#!/bin/bash
# Запустить все русские демо по очереди
# ./demo-all-ru.sh          — запустить всё
# ./demo-all-ru.sh rainbow  — запустить конкретное

CDIR="$(cd "$(dirname "$0")" && pwd)"
DEMOS=(
    demo-status-ru.sh
    demo-rainbow-ru.sh
    demo-quotes-ru.sh
    demo-moods-ru.sh
)

if [ -n "$1" ]; then
    # Запустить конкретный
    name="demo-${1}-ru.sh"
    if [ -f "$CDIR/$name" ]; then
        echo "=== Запуск $name ==="
        bash "$CDIR/$name"
    else
        echo "Неизвестное демо: $1"
        echo "Доступные: ${DEMOS[*]}"
        exit 1
    fi
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
    echo "Все демо завершены!"
fi
