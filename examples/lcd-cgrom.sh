#!/bin/bash
# lcd-cgrom.sh — пошаговый просмотр знакогенератора (CGROM) LCD 1602A.
#
# На каждом шаге на LCD выводятся 32 символа CGROM:
#   строка 0 (row0): коды  base .. base+15
#   строка 1 (row1): коды  base+16 .. base+31
# Переход к следующим 32 символам выполняется по нажатию любой клавиши.
# В терминале печатается текущий диапазон выводимых кодов.
#
# Переключение СТРАНИЦЫ CGROM (таблицы знакогенератора):
#   страница выбирается на самом контроллере LCD, а не кодами символов.
#   Послать по Serial ESP32 (напр. /dev/ttyACM0, 921600 8N1) команду P<n>:
#     printf 'P2\n' > /dev/ttyACM0   # страница 2 (кириллица A02)
#     printf 'P0\n' > /dev/ttyACM0   # страница 0 (ASCII 0x20-0x7F)
#   Доступны P0..P3. Затем запустить этот скрипт, чтобы увидеть содержимое
#   выбранной страницы.
#
# Используется raw-режим:  aura-ctl lcd --raw <hex>

set -u

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"

if [ ! -x "$CTL" ]; then
	echo "Не найден бинарь aura-ctl: $CTL" >&2
	echo "Соберите его: cd aura-indicator && go build -o bin/aura-ctl ./cmd/aura-ctl/" >&2
	exit 1
fi

TOTAL=256
STEP=32
base=0

echo "Просмотр CGROM LCD (по 32 символа за шаг)."
echo "Нажимайте любую клавишу для перехода к следующим 32 символам; выход — Ctrl-C."
echo

while [ "$base" -lt "$TOTAL" ]; do
	end=$(( base + STEP - 1 ))

	# Построить hex-строку: 32 кода, по 2 hex-цифры на код.
	hex=""
	for ((c = base; c < base + STEP; c++)); do
		hex="$hex$(printf '%02x' "$c")"
	done

	# Отправить сырые коды CGROM на LCD (row0 + row1).
	"$CTL" lcd --raw "$hex"

	# Показать диапазон в терминале.
	printf 'CGROM  0x%02X-0x%02X   (десятичные %3d-%3d)\n' "$base" "$end" "$base" "$end"

	printf 'Нажмите любую клавишу для следующих 32 символов... '
	read -n 1 -s -r
	echo

	base=$(( base + STEP ))
done

echo "Готово: весь CGROM (256 символов) показан."
