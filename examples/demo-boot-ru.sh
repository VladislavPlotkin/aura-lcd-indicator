#!/bin/bash
# Демо: приветствие при загрузке (русский)
# Показывает приветствие на LCD и плавно меняет цвет LED

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; printf "%-16s" "${s:0:16}"; }
join() { echo "$(pad "$1")$(pad "$2")"; }

echo "=== Приветствие ==="

$CTL lcd "$(join "AURA RGB v2" "Загрузка...")"
$CTL set led2 static 0,0,255
sleep 2

$CTL lcd "$(join "Sistema gotova" "Gotov k rabote")"
$CTL blink 0,255,0 3 200
sleep 1

# Плавное переключение цветов
for color in "255,0,0" "0,255,0" "0,0,255" "255,255,255"; do
    $CTL set led2 static "$color"
    sleep 0.5
done

$CTL lcd "$(join "Vse ok!" "Pognaaaali!")"
$CTL set led2 static 0,255,0
sleep 1

$CTL off led2
echo "Приветствие завершено!"
