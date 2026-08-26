#!/bin/bash
# Демо: имитация прогресса / статусов (русский)
# LCD: 16 символов на строку, 2 строки

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; s="${s:0:16}"; local len=${#s}; printf '%s' "$s"; local i; for ((i=len; i<16; i++)); do printf ' '; done; }
join() { echo "$(pad "$1")$(pad "$2")"; }

echo "=== Имитация статусов ==="

$CTL lcd "$(join "Думаю..." "шаг 1/5")"
sleep 2

$CTL lcd "$(join "Планирую" "шаг 2/5")"
$CTL set led2 static 0,0,255
sleep 2

$CTL lcd "$(join "Кодирую..." "шаг 3/5")"
sleep 2

$CTL lcd "$(join "Ревью" "шаг 4/5")"
$CTL blink 255,255,0 3 300
sleep 2

$CTL lcd "$(join "Тестирую" "шаг 5/5")"
$CTL set led2 static 255,165,0
sleep 2

$CTL lcd "$(join "ГОТОВО!" "УСПЕХ!")"
$CTL set led2 static 0,255,0
$CTL blink 0,255,0 3 200
sleep 1

$CTL off led2
echo "Готово! (зелёный = успех)"
