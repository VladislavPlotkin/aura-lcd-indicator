#!/bin/bash
# Запустить все демо по очереди
# ./demo-all.sh          — запустить всё
# ./demo-all.sh rainbow  — запустить конкретное

CDIR="$(cd "$(dirname "$0")" && pwd)"
DEMOS=(
    demo-full.sh
    demo-lines.sh
    demo-leds.sh
    demo-moods.sh
    demo-quotes.sh
    demo-matrix.sh
    demo-game.sh
    demo-rainbow.sh
    demo-status.sh
)
DEMOS_RU=(
    demo-status-ru.sh
    demo-rainbow-ru.sh
    demo-quotes-ru.sh
    demo-moods-ru.sh
    demo-boot-ru.sh
    demo-pulse-ru.sh
)

if [ -n "$1" ]; then
    # Запустить конкретный
    name="demo-$1.sh"
    if [ -f "$CDIR/$name" ]; then
        echo "=== Running $name ==="
        bash "$CDIR/$name"
    else
        echo "Unknown demo: $1"
        echo "Available (EN): ${DEMOS[*]}"
        echo "Available (RU): ${DEMOS_RU[*]}"
        exit 1
    fi
elif [ "$1" = "ru" ]; then
    # Запустить все русские демо
    for demo in "${DEMOS_RU[@]}"; do
        echo ""
        echo "================================================"
        echo ">>> $demo"
        echo "================================================"
        bash "$CDIR/$demo"
        sleep 1
    done
    echo ""
    echo "All RU demos done!"
else
    # Запустить все
    for demo in "${DEMOS[@]}"; do
        echo ""
        echo "================================================"
        echo ">>> $demo"
        echo "================================================"
        bash "$CDIR/$demo"
        sleep 1
    done
    echo ""
    echo "All demos done!"
fi
