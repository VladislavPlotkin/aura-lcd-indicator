/**
 * ASUS AURA ARGB — упрощенный захват RMT RAW (RX) + вывод на LED и LCD
 * ========================================================================
 * Copyright (c) 2026 Vladislav Plotkin
 * SPDX-License-Identifier: MIT
 *
 * Упрощенная версия с поддержкой компиляции под ESP32 и ESP32-C3.
 * Убраны тесты, конвертеры символов и лишние режимы.
 * LCD: принудительно установлена 2-я кодовая таблица символов (A00/A02).
 */

#include <Arduino.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include <string.h>

// ============================================================================
// Конфигурация — пины для ESP32 / ESP32-C3
// ============================================================================
#if defined(ARDUINO_ESP32C3_DEV)
  // ESP32-C3
  #define AURA_PIN        1
  #define LED_PIN         2       // синий LED
  #define EXT_LED_PIN     4       // красный LED
  #define GREEN_LED_PIN   3       // зелёный LED
  
  // LCD пины для C3
  #define LCD_RS   9
  #define LCD_EN   10
  #define LCD_D4   5
  #define LCD_D5   6
  #define LCD_D6   7
  #define LCD_D7   8

  #define RX_MEM_BLOCKS RMT_MEM_NUM_BLOCKS_2
#else
  // Оригинальный ESP32
  #define AURA_PIN    16
  #define LED_PIN     2        // встроенный синий LED
  #define EXT_LED_PIN 15       // внешний красный LED
  #define GREEN_LED_PIN 4      // внешний зелёный LED

  // LCD пины для ESP32
  #define LCD_RS   14
  #define LCD_EN   27
  #define LCD_D4   32
  #define LCD_D5   33
  #define LCD_D6   25
  #define LCD_D7   26

  #define RX_MEM_BLOCKS RMT_MEM_NUM_BLOCKS_6
#endif

#define LCD_COLS 16
#define LCD_ROWS 2
#define LCD_SIZE (LCD_COLS * LCD_ROWS)

#define LEDC_FREQ   5000
#define LEDC_BITS   8
#define RMT_BAUD    921600
#define MAX_ITEMS   3000
#define TICK_HZ     40000000UL  // 40 МГц -> 25 нс такт
#define IDLE_TICKS  4000        // 100 мкс порог сброса-паузы

static const uint32_t tick_ns = 1000000000UL / TICK_HZ; // = 25 нс

// ============================================================================
// Состояние
// ============================================================================
static bool       g_rx_ok = false;
static rmt_data_t g_rx_buf[MAX_ITEMS];
static uint32_t   g_frame = 0;

// Буфер для LCD
static char lcd_buf[LCD_SIZE];
static bool lcd_dirty = false;
static unsigned long s_last_lcd_update = 0;

// Декодированные цвета (массив: 12 LED * (R,G,B))
static uint8_t lcd_colors[12][3];

// ============================================================================
// Драйвер LCD (HD44780 4-бита)
// ============================================================================

static void lcd_pulse_en(void) {
    digitalWrite(LCD_EN, HIGH);
    delayMicroseconds(1);
    digitalWrite(LCD_EN, LOW);
    delayMicroseconds(5);
}

static void lcd_write_nibble(uint8_t nib) {
    digitalWrite(LCD_D4, (nib & 0x01) ? HIGH : LOW);
    digitalWrite(LCD_D5, (nib & 0x02) ? HIGH : LOW);
    digitalWrite(LCD_D6, (nib & 0x04) ? HIGH : LOW);
    digitalWrite(LCD_D7, (nib & 0x08) ? HIGH : LOW);
    lcd_pulse_en();
}

static void lcd_write(uint8_t val, bool is_data) {
    digitalWrite(LCD_RS, is_data ? HIGH : LOW);
    lcd_write_nibble(val >> 4);
    lcd_write_nibble(val & 0x0F);
    delayMicroseconds(100);
}

static void lcd_init(void) {
    pinMode(LCD_RS, OUTPUT);
    pinMode(LCD_EN, OUTPUT);
    pinMode(LCD_D4, OUTPUT);
    pinMode(LCD_D5, OUTPUT);
    pinMode(LCD_D6, OUTPUT);
    pinMode(LCD_D7, OUTPUT);

    delay(150);

    digitalWrite(LCD_RS, LOW);
    delayMicroseconds(100);
    lcd_write_nibble(0x03); delayMicroseconds(4500);
    lcd_write_nibble(0x03); delayMicroseconds(150);
    lcd_write_nibble(0x03); delayMicroseconds(100);
    lcd_write_nibble(0x02); delayMicroseconds(100);

    lcd_write(0x2A, false);  // 4-bit, 2 строки, 5x8 + вторая страница знакогенератора
    lcd_write(0x0C, false);  // дисплей вкл
    lcd_write(0x06, false);  // инкремент без сдвига
    lcd_write(0x01, false);  // очистка
    delay(5);
}

static void lcd_set_cursor(uint8_t row, uint8_t col) {
    uint8_t addr = (row == 0 ? 0x80 : 0xC0) | (col & 0x0F);
    lcd_write(addr, false);
}

// Вывод обычной ASCII строки без конвертации (для теста)
static void lcd_print_ascii(const char *str, uint8_t row, uint8_t col) {
    lcd_set_cursor(row, col);
    while (*str) {
        lcd_write((uint8_t)*str++, true);
    }
}

// Переключение на вторую страницу знакогенератора (CG ROM A02) командой 0x2A.
// Команда не очищает экран и не сбрасывает курсор, поэтому безопасно
// вызывать повторно в любой момент.
static void lcd_set_second_table(void) {
    lcd_write(0x2A, false); // Function Set: 4-бит, 2 строки, 5x8, 2-я таблица CG
    delay(5);
}

// Обновление LCD по буферу
static void lcd_update_display(void) {
    if (!lcd_dirty) return;
    
    lcd_set_second_table(); // гарантируем работу со 2-й страницей знакогенератора
    lcd_set_cursor(0, 0);
    for (int i = 0; i < LCD_COLS; i++) {
        lcd_write((uint8_t)lcd_buf[i], true);
    }
    lcd_set_cursor(1, 0);
    for (int i = LCD_COLS; i < LCD_SIZE; i++) {
        lcd_write((uint8_t)lcd_buf[i], true);
    }
    lcd_dirty = false;
}

// Заполнение буфера LCD цветами LED
static void unpack_to_lcd(const uint8_t colors[][3]) {
    bool changed = false;
    for (int i = 0; i < LCD_SIZE; i++) {
        int led_idx = i / 3 + 1;      // +1: пропускаем LED0 (индикатор)
        int ch = i % 3;               // 0=R, 1=G, 2=B
        
        uint8_t code = colors[led_idx][ch];
        // Без фильтра: выводим любой код, который есть в знакогенераторе
        // (0x00-0x1F — спецсимволы/CGRAM, 0x20-0x7F — ASCII, 0xA0-0xFF — кириллица и символы).
        char c = (char)code;
        
        if (c != lcd_buf[i]) changed = true;
        lcd_buf[i] = c;
    }
    
    if (changed) {
        lcd_dirty = true;
    }
}

// ============================================================================
// Декодирование кадра WS2812B (GRB, MSB first)
// ============================================================================

static int decode_full_frame(const rmt_data_t *buf, int total, uint8_t out[36]) {
    int bits = 0, byte = 0, nbytes = 0;
    for (int i = 0; i < total && nbytes < 36; i++) {
        uint32_t d0 = buf[i].duration0 * tick_ns;
        if (d0 >= 5000) break;            // сброс-пауза -> конец кадра
        int bit = (d0 >= 600) ? 1 : 0;    // HIGH >= 600 нс => бит 1
        byte = (byte << 1) | bit;
        if (++bits == 8) { out[nbytes++] = byte; byte = 0; bits = 0; }
    }
    return nbytes;
}

// ============================================================================
// Настройка
// ============================================================================

void setup() {
    Serial.begin(RMT_BAUD);
    delay(100);

    // LED
    ledcAttach(LED_PIN, LEDC_FREQ, LEDC_BITS);
    ledcAttach(EXT_LED_PIN, LEDC_FREQ, LEDC_BITS);
    ledcAttach(GREEN_LED_PIN, LEDC_FREQ, LEDC_BITS);
    ledcWrite(LED_PIN, 0);
    ledcWrite(EXT_LED_PIN, 0);
    ledcWrite(GREEN_LED_PIN, 0);

    #if defined(ARDUINO_ESP32C3_DEV)
        Serial.println("\n[BOOT] Упрощенный AURA ARGB Capture (ESP32-C3)");
    #else
        Serial.println("\n[BOOT] Упрощенный AURA ARGB Capture (ESP32)");
    #endif

    Serial.printf("[RMT] Init RX on GPIO%d...\n", AURA_PIN);
    
    if (!rmtInit(AURA_PIN, RMT_RX_MODE, RX_MEM_BLOCKS, TICK_HZ)) {
        Serial.println("[RMT] RX init FAILED!");
    } else {
        g_rx_ok = true;
        Serial.println("[RMT] RX init OK.");
    }
    rmtSetRxMaxThreshold(AURA_PIN, IDLE_TICKS);

    // LCD init и принудительная установка 2-й таблицы
    lcd_init();
    lcd_set_second_table();  // Принудительно ставим 2-ю таблицу
    lcd_print_ascii("AURA RGB", 0, 0);
    lcd_print_ascii("Waiting...", 1, 0);
    
    Serial.println("[LCD] Инициализирован, 2-я кодовая таблица установлена.");
    Serial.println("[RMT] Ожидание данных...");
}

// ============================================================================
// Основной цикл
// ============================================================================

void loop() {
    size_t num = MAX_ITEMS;
    
    // Попытка чтения кадра
    if (rmtRead(AURA_PIN, g_rx_buf, &num, 100)) {
        g_frame++;
        
        // Декодируем все 12 светодиодов (36 байт)
        uint8_t all_leds[36];
        int n = decode_full_frame(g_rx_buf, (int)num, all_leds);

        // Выводим LED0 на физические светодиоды
        ledcWrite(EXT_LED_PIN,   all_leds[1]);   // R0
        ledcWrite(GREEN_LED_PIN, all_leds[0]);   // G0
        ledcWrite(LED_PIN,       all_leds[2]);   // B0
        
        // Заполняем массив цветов для LCD
        for (int i = 0; i < 12; i++) {
            int base = i * 3;   // формат GRB
            lcd_colors[i][0] = all_leds[base + 1];   // R
            lcd_colors[i][1] = all_leds[base + 0];   // G
            lcd_colors[i][2] = all_leds[base + 2];   // B
        }
        
        // Обновляем LCD (ограничиваем частоту, т.к. экран медленный)
        unsigned long now = millis();
        if (now - s_last_lcd_update >= 500) { // обновление 2 раза в сек
            unpack_to_lcd(lcd_colors);
            lcd_update_display();
            s_last_lcd_update = now;
        }
        
        // Только для отладки в Serial (LED0)
        static uint8_t last_r = 255, last_g = 255, last_b = 255;
        if (all_leds[0] != last_g || all_leds[1] != last_r || all_leds[2] != last_b) {
            Serial.printf("LED0: R=%3d G=%3d B=%3d\n", all_leds[1], all_leds[0], all_leds[2]);
            last_r = all_leds[1]; last_g = all_leds[0]; last_b = all_leds[2];
        }
    } else {
        // Нет данных
        delay(10);
    }
    
    delay(1);
}
