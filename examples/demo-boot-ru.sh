#!/bin/bash
# Демо: приветствие при загрузке (русский)
# Показывает приветствие на LCD и плавно меняет цвет LED

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; s="${s:0:16}"; local len=${#s}; printf '%s' "$s"; local i; for ((i=len; i<16; i++)); do printf ' '; done; }
join() { echo "$(pad "$1")$(pad "$2")"; }

echo "=== Приветствие ==="

$CTL lcd "$(join "AURA RGB v2" "Загрузка...")"
$CTL set led2 static 0,0,255
sleep 2

$CTL lcd "$(join "Система готова" "Готов к работе")"
$CTL blink 0,255,0 3 200
sleep 1

# Плавное переключение цветов
for color in "255,0,0" "0,255,0" "0,0,255" "255,255,255"; do
    $CTL set led2 static "$color"
    sleep 0.5
done

$CTL lcd "$(join "Всё ок!" "Погнали!")"
$CTL set led2 static 0,255,0
sleep 1

$CTL off led2
echo "Приветствие завершено!"
