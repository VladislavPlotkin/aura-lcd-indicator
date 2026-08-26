#!/bin/bash
# Демо: случайные фразы на LCD (русский)
# LCD: 16 символов на строку, 2 строки

CDIR="$(cd "$(dirname "$0")" && pwd)"
CTL="$CDIR/../aura-indicator/bin/aura-ctl"
pad() { local s="$1"; printf "%-16s" "${s:0:16}"; }
join() { echo "$(pad "$1")$(pad "$2")"; }

# Пары фраз: строка0 | строка1 (каждая ≤ 16 ASCII символов)
phrases=(
  # IT / Программирование
  "Hello world!|Privet mir!"
  "sudo make me|a sandwich"
  "eto ne bag,|eto feature"
  "Doveriaj mne,|ya developer"
  "Kofe: VKL | Son: VYL"
  "Pognaaaali!|Debugging..."
  "404|Um otsutstvuet"
  "Zagruzka...|Podozhdite"
  "Vstavte monetu|dlya prodolzh"
  "git commit|-m 'rabotaet'"
  "while(true)|kofe++"
  "Kompiliruetsya|otpravlyaj!"
  "Seg fault|core dumped"
  "V teorii,|na practice"
  "chmod 777|reshaet vsyo"
  "Ya ispravlyu|zavtra? ;)"
  "Deploj v|pyatnitsu!"
  "Testy idut|podozhdite"
  "V productiyu|derzhi pivo"
  "Rabotaet na|moem PC"

  # Жизнь / Мемы
  "Soxranjaj|spokoistvie"
  "Prosto sdelaj!|...zavtra"
  "Vse v poryadke|ya v norme"
  "U tebya byla|ODNA zadacha!"
  "Syuzhetnyi|povorot: feature"
  "Byt ili ne|byt"
  "Mne nuzhen kofei|i son"
  "Ponedelnik|snova pomogite"
  "Pjatnica!|Uzhe ponedelnik?"
  "Pochti|doshli"
  "Yeshche raz|ty smozhesh"
  "Ne pani-|kuj, vsyo ok"
  "Terpenie| - eto drug"
  "Nikogda ne|sdavajsya"
  "Esli srazu ne|poluchilos..."
  "Praktika|sovershenstvuet"
  "Est. Spat.|Kodit. Povtoryat."
  "42|Otvet na vsyo"
  "Ne zabud svoy|polotence"
  "Segodnya horo-|shiy den"
)

echo "=== Случайные фразы ==="

idx=$(( RANDOM % ${#phrases[@]} ))
IFS='|' read -r line0 line1 <<< "${phrases[$idx]}"

$CTL lcd "$(join "$line0" "$line1")"
$CTL set led2 static 0,200,255

echo "Фраза #$idx: $line0 | $line1"
echo "Готово!"
