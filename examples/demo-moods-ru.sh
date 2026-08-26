#!/bin/bash
# Демо: настроения (цвета индикатора)
# Каждое настроение = определённый цвет + подпись на LCD

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; printf "%-16s" "${s:0:16}"; }
join() { echo "$(pad "$1")$(pad "$2")"; }

echo "=== Демо настроений ==="

# Название: цвет | LCD строка 0 | LCD строка 1
moods=(
  "255,0,0|Radost|Veselo"
  "0,0,255|Grust|Toska"
  "255,165,0|Zlost|B٪sh"
  "255,255,0|Udivlenie|Ogo!"
  "0,255,0|Pokoy|Tishina"
  "128,0,128|Lyubov|Serdtse"
  "0,200,200|Ustalost|Zzz..."
  "255,100,50|Voodushevlenie|Vpered!"
  "200,200,200|Neitral|Obychno"
)

for mood in "${moods[@]}"; do
  IFS='|' read -r color line0 line1 <<< "$mood"
  $CTL lcd "$(join "$line0" "$line1")"
  $CTL set led2 static "$color"
  echo "  $line0 $line1: $color"
  sleep 3
done

$CTL off led2
echo "Демо завершено!"
