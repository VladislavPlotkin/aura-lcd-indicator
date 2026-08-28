// Copyright (c) 2026 Vladislav Plotkin
// SPDX-License-Identifier: MIT
package aura

// HD44780 LCD characters (ASCII range)
const (
	LCDRows = 2
	LCDCols = 16
)

// BuildLEDColors строит массив LED-цветов из текста.
// Первый LED (индекс 0) — заглушка (чёрный), используется для индикатора.
// Каждый LED кодирует 3 символа: R, G, B каналы.
// Возвращает [][3]byte длиной (rows*cols + 2) / 3 + 1 ...
// Правильная формула:
//
//	totalChars = rows * cols
//	numLEDs = 1 + (totalChars + 2) / 3  // +1 для LED0 (индикатор)
//
// colors[0] = {0,0,0} — зарезервировано для индикатора
// Для i от 0 до len(runes)-1:
//
//	ledIdx = i / 3 + 1    // +1 пропускает LED0
//	channel = i % 3       // 0=R, 1=G, 2=B
//	colors[ledIdx][channel] = CharToLCDIndex(runes[i])
func BuildLEDColors(text string, rows, cols int) [][3]byte {
	totalChars := rows * cols
	numLEDs := 1 + (totalChars+2)/3 // +1 для LED0, ceil деление
	colors := make([][3]byte, numLEDs)

	runes := []rune(text)
	for i := 0; i < totalChars && i < len(runes); i++ {
		ledIdx := i/3 + 1
		channel := i % 3
		colors[ledIdx][channel] = CharToLCDIndex(runes[i])
	}

	return colors
}

// BuildLEDColorsFromCodes строит массив LED-цветов напрямую из сырых
// байт-кодов CGROM, без маппинга и без интерпретации UTF-8.
// Используется raw-режимом для передачи произвольных кодов (0x00-0xFF).
// Позиция 16 — граница строк (row0: 0-15, row1: 16-31).
func BuildLEDColorsFromCodes(codes []byte, rows, cols int) [][3]byte {
	totalChars := rows * cols
	numLEDs := 1 + (totalChars+2)/3 // +1 для LED0, ceil деление
	colors := make([][3]byte, numLEDs)

	for i := 0; i < totalChars && i < len(codes); i++ {
		ledIdx := i/3 + 1 // +1 пропускает LED0 (индикатор)
		channel := i % 3  // 0=R, 1=G, 2=B
		// Convert CGROM code to index
		idx := CGROMToIndex(codes[i])
		if int(idx) >= LCD_CGromSize {
			idx = 0 // unknown code → space (index 0)
		}
		colors[ledIdx][channel] = idx
	}
	return colors
}

// SendLCDRaw отправляет произвольные байт-коды CGROM на LCD (raw-режим).
// codes — сырые коды знакогенератора (0x00-0xFF). Обрезаются/дополняются
// до 32 (totalChars). Хвост дополняется 0x00 (пустой символ CGRAM0).
func SendLCDRaw(codes []byte, devicePath string) (string, error) {
	total := LCDRows * LCDCols

	buf := make([]byte, total)
	if len(codes) > total {
		copy(buf, codes[:total])
	} else {
		copy(buf, codes)
	}

	colors := BuildLEDColorsFromCodes(buf, LCDRows, LCDCols)

	// Извлекаем LCD часть (LED1..LED32), паддим до 32 записей
	lcdColors := make([][3]byte, 0, LCDLedCount)
	if len(colors) > 1 {
		lcdColors = append(lcdColors, colors[1:]...)
	}
	for len(lcdColors) < LCDLedCount {
		lcdColors = append(lcdColors, [3]byte{0, 0, 0})
	}

	state := LoadState()
	SetLCDColors(state, lcdColors)
	if err := SaveState(state); err != nil {
		return "", err
	}

	warning := ""
	if len(codes) > total {
		warning = "raw codes truncated to 32"
	}
	return sendColors(GetFullColors(state), devicePath, warning)
}

// SendLCDText отправляет текст на LCD через direct-режим led2.
//
// Если row < 0 — full-screen режим (32 символа, 2 строки).
// Если row == 0 или row == 1 — single-line режим:
//
//	загружается существующий state, обновляется только указанная строка,
//	вторая строка сохраняется, хвост строки зачищается пробелами.
//
// Передаёт в led2 полный 33-LED буфер (индикатор + LCD).
//
// Возвращает предупреждение, если текст был обрезан.
func SendLCDText(text string, row int, devicePath string) (string, error) {
	warning := ""

	if row < 0 {
		// Full-screen: 32 символа
		display := SplitToDisplay(text, LCDRows, LCDCols)
		if len([]rune(text)) > LCDRows*LCDCols {
			warning = "text truncated to 32 chars"
		}
		colors := BuildLEDColors(display, LCDRows, LCDCols)

		// Извлекаем LCD часть (LED1..LED32), паддим до 32 записей
		lcdColors := make([][3]byte, 0, LCDLedCount)
		if len(colors) > 1 {
			lcdColors = append(lcdColors, colors[1:]...)
		}
		for len(lcdColors) < LCDLedCount {
			lcdColors = append(lcdColors, [3]byte{0, 0, 0})
		}

		// Сохраняем в state, отправляем полный буфер (индикатор + LCD)
		state := LoadState()
		SetLCDColors(state, lcdColors)
		if err := SaveState(state); err != nil {
			return "", err
		}
		return sendColors(GetFullColors(state), devicePath, warning)
	}

	// Single-line mode
	if row >= LCDRows {
		return "", nil // ignore invalid row
	}

	// Загружаем текущее состояние
	state := LoadState()

	// Чистим строку: берём первые LCDCols символов
	runes := []rune(text)
	effectiveLen := len(runes)
	if effectiveLen > LCDCols {
		effectiveLen = LCDCols
		warning = "text truncated to 16 chars"
	}

	// Заполняем каналы для строки row
	// Маппинг: строка 0 = позиции 0-15, строка 1 = позиции 16-31
	startPos := row * LCDCols
	for p := 0; p < effectiveLen; p++ {
		absP := startPos + p
		ledIdx := absP / 3
		channel := absP % 3 // 0=R, 1=G, 2=B
		state.LCD[ledIdx][channel] = CharToLCDIndex(runes[p])
	}
	// Зачищаем хвост строки пробелами
	for p := effectiveLen; p < LCDCols; p++ {
		absP := startPos + p
		ledIdx := absP / 3
		channel := absP % 3
		state.LCD[ledIdx][channel] = CharToLCDIndex(' ')
	}

	if err := SaveState(state); err != nil {
		return "", err
	}

	fullColors := GetFullColors(state)
	return sendColors(fullColors, devicePath, warning)
}

// sendColors открывает устройство и отправляет цвета в led2.
func sendColors(colors [][3]byte, devicePath, warning string) (string, error) {
	dev, err := OpenDevice(devicePath)
	if err != nil {
		return "", err
	}
	defer dev.Close()

	if err := dev.SetDirect("led2", colors); err != nil {
		return "", err
	}

	return warning, nil
}

// SetIndicatorDirect устанавливает цвет индикатора (LED0) через state + direct send.
func SetIndicatorDirect(r, g, b byte, devicePath string) error {
	state := LoadState()
	SetIndicator(state, r, g, b)
	if err := SaveState(state); err != nil {
		return err
	}

	dev, err := OpenDevice(devicePath)
	if err != nil {
		return err
	}
	defer dev.Close()

	return dev.SetDirect("led2", GetFullColors(state))
}
