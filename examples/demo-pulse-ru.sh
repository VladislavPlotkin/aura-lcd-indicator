#!/bin/bash
# Демо: пульсация яркости LED
# Плавное увеличение и уменьшение яркости

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; printf "%-16s" "${s:0:16}"; }
join() { echo "$(pad "$1")$(pad "$2")"; }

echo "=== Пульсация ==="

$CTL lcd "$(join "Пульсация" "8 циклов")"

# Красный пульс
echo "  Красный..."
for i in $(seq 0 5 255); do
    $CTL set led2 static "$i,0,0"
    sleep 0.02
done
for i in $(seq 255 -5 0); do
    $CTL set led2 static "$i,0,0"
    sleep 0.02
done

# Синий пульс
echo "  Синий..."
for i in $(seq 0 5 255); do
    $CTL set led2 static "0,0,$i"
    sleep 0.02
done
for i in $(seq 255 -5 0); do
    $CTL set led2 static "0,0,$i"
    sleep 0.02
done

# Зелёный пульс
echo "  Зелёный..."
for i in $(seq 0 5 255); do
    $CTL set led2 static "0,$i,0"
    sleep 0.02
done
for i in $(seq 255 -5 0); do
    $CTL set led2 static "0,$i,0"
    sleep 0.02
done

# Белый пульс
echo "  Белый..."
for i in $(seq 0 5 255); do
    $CTL set led2 static "$i,$i,$i"
    sleep 0.02
done
for i in $(seq 255 -5 0); do
    $CTL set led2 static "$i,$i,$i"
    sleep 0.02
done

$CTL off led2
$CTL lcd "$(join "Готово!" "Конец")"
echo "Пульсация завершена!"
