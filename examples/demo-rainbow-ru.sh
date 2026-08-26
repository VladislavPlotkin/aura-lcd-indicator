#!/bin/bash
# Демо: радуга на индикаторе (русский)

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; s="${s:0:16}"; local len=${#s}; printf '%s' "$s"; local i; for ((i=len; i<16; i++)); do printf ' '; done; }
join() { echo "$(pad "$1")$(pad "$2")"; }

RAINBOW=(
    "255,0,0"
    "255,127,0"
    "255,255,0"
    "0,255,0"
    "0,0,255"
    "75,0,130"
    "143,0,255"
)

RAINBOW_NAMES=(
    "Красный"
    "Оранжевый"
    "Жёлтый"
    "Зелёный"
    "Синий"
    "Индиго"
    "Фиолетовый"
)

echo "=== Демо радуга ==="

$CTL lcd "$(join "Радуга!" "7 цветов!")"

for i in "${!RAINBOW[@]}"; do
    $CTL set led2 static "${RAINBOW[$i]}"
    echo "  ${RAINBOW_NAMES[$i]}: ${RAINBOW[$i]}"
    sleep 0.7
done

$CTL lcd "$(join "Радуга 2x" "быстро!")"

for round in 1 2; do
    for color in "${RAINBOW[@]}"; do
        $CTL set led2 static "$color"
        sleep 0.3
    done
done

$CTL lcd "$(join "Мигание" "радугой!")"

for i in 1 2 3; do
    $CTL set led2 static 255,0,0
    sleep 0.2
    $CTL set led2 static 0,255,0
    sleep 0.2
    $CTL set led2 static 0,0,255
    sleep 0.2
done

$CTL off led2
echo "Радуга завершена!"
