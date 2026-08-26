// Copyright (c) 2026 Vladislav Plotkin
// SPDX-License-Identifier: MIT
package aura

import (
	"testing"
)

func TestBuildDirectPackets_CountField(t *testing.T) {
	// Simulate LCD send: 33 colors (1 indicator + 32 LCD)
	colors := make([][3]byte, 33)
	for i := range colors {
		colors[i] = [3]byte{byte(i), byte(i + 1), byte(i + 2)}
	}

	packets := BuildDirectPackets(0, colors, 0)

	if len(packets) != 2 {
		t.Fatalf("expected 2 packets, got %d", len(packets))
	}

	// Packet 1: LEDs 0-19, no apply flag
	p1 := packets[0]
	if p1[0] != 0xEC {
		t.Errorf("pkt[0] = 0x%02X, want 0xEC", p1[0])
	}
	if p1[1] != 0x40 {
		t.Errorf("pkt[1] = 0x%02X, want 0x40", p1[1])
	}
	if p1[2] != 0x00 {
		t.Errorf("pkt[2] = 0x%02X, want 0x00 (no apply flag)", p1[2])
	}
	if p1[3] != 0x00 {
		t.Errorf("pkt[3] = 0x%02X, want 0x00 (start LED)", p1[3])
	}
	// count should be the number of LEDs (1-based)
	if p1[4] != 20 {
		t.Errorf("pkt[4] = %d, want 20 (20 LEDs)", p1[4])
	}

	// Packet 2: LEDs 20-32, apply flag
	p2 := packets[1]
	if p2[0] != 0xEC {
		t.Errorf("pkt[0] = 0x%02X, want 0xEC", p2[0])
	}
	if p2[1] != 0x40 {
		t.Errorf("pkt[1] = 0x%02X, want 0x40", p2[1])
	}
	if p2[2] != 0x80 {
		t.Errorf("pkt[2] = 0x%02X, want 0x80 (apply flag)", p2[2])
	}
	if p2[3] != 20 {
		t.Errorf("pkt[3] = %d, want 20 (start LED)", p2[3])
	}
	// count should be the number of LEDs (1-based)
	if p2[4] != 13 {
		t.Errorf("pkt[4] = %d, want 13 (13 LEDs)", p2[4])
	}

	// Verify total LED count in both packets
	totalLEDs := int(p1[4]) + int(p2[4])
	if totalLEDs != 33 {
		t.Errorf("total LEDs = %d, want 33", totalLEDs)
	}

	// Verify color data in first packet
	if p1[5] != 0 { // R of first color = colors[0][0] = 0
		t.Errorf("p1 data byte 0 = 0x%02X, want 0x00", p1[5])
	}
	if p1[6] != 1 { // G of first color = colors[0][1] = 1
		t.Errorf("p1 data byte 1 = 0x%02X, want 0x01", p1[6])
	}
	if p1[7] != 2 { // B of first color = colors[0][2] = 2
		t.Errorf("p1 data byte 2 = 0x%02X, want 0x02", p1[7])
	}
}

func TestBuildDirectPackets_SingleLED(t *testing.T) {
	// Test with single LED
	colors := [][3]byte{{255, 0, 0}}
	packets := BuildDirectPackets(0, colors, 0)
	if len(packets) != 1 {
		t.Fatalf("expected 1 packet, got %d", len(packets))
	}
	p := packets[0]
	if p[4] != 1 { // 1 LED → count = 1
		t.Errorf("pkt[4] = %d, want 1 (1 LED)", p[4])
	}
	if p[2] != 0x80 { // single packet = last = apply flag
		t.Errorf("pkt[2] = 0x%02X, want 0x80", p[2])
	}
}

func TestBuildDirectPackets_ExactMaxLEDs(t *testing.T) {
	// Test with exactly 20 LEDs (fits in one packet)
	colors := make([][3]byte, 20)
	packets := BuildDirectPackets(0, colors, 0)
	if len(packets) != 1 {
		t.Fatalf("expected 1 packet for 20 LEDs, got %d", len(packets))
	}
	p := packets[0]
	if p[4] != 20 { // 20 LEDs → count = 20
		t.Errorf("pkt[4] = %d, want 20", p[4])
	}
	if p[2] != 0x80 { // single packet = last = apply flag
		t.Errorf("pkt[2] = 0x%02X, want 0x80", p[2])
	}
}

func TestBuildLEDColors_Positions(t *testing.T) {
	// After SplitToDisplay padding, input is always 32 chars
	text := SplitToDisplay("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123", 2, 16)
	colors := BuildLEDColors(text, 2, 16)

	// Expected: 1 indicator + ceil(32/3) = 12 entries
	if len(colors) != 12 {
		t.Fatalf("expected 12 color entries, got %d", len(colors))
	}

	// LED0 should be black (indicator placeholder)
	if colors[0] != [3]byte{0, 0, 0} {
		t.Errorf("colors[0] = %v, want {0,0,0}", colors[0])
	}

	// LED1 R = 'A', G = 'B', B = 'C'
	if colors[1][0] != 'A' {
		t.Errorf("LED1 R = %c, want 'A'", colors[1][0])
	}
	if colors[1][1] != 'B' {
		t.Errorf("LED1 G = %c, want 'B'", colors[1][1])
	}
	if colors[1][2] != 'C' {
		t.Errorf("LED1 B = %c, want 'C'", colors[1][2])
	}

	// LED6: positions 15(R), 16(G), 17(B) — row boundary!
	// With SplitToDisplay padding: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123  "
	// Pos 15 = 'P', Pos 16 = 'Q', Pos 17 = 'R'
	if colors[6][0] != 'P' {
		t.Errorf("LED6 R = %c, want 'P' (pos 15)", colors[6][0])
	}
	if colors[6][1] != 'Q' {
		t.Errorf("LED6 G = %c, want 'Q' (pos 16)", colors[6][1])
	}
	if colors[6][2] != 'R' {
		t.Errorf("LED6 B = %c, want 'R' (pos 17)", colors[6][2])
	}

	// Positions 29,30,31: '3' at 29, spaces at 30 and 31
	if colors[10][2] != '3' { // position 29 = '3'
		t.Errorf("LED10 B (pos 29) = %c, want '3'", colors[10][2])
	}
	if colors[11][0] != 0x20 { // position 30 = space (padded)
		t.Errorf("LED11 R (pos 30) = 0x%02X, want space", colors[11][0])
	}
	if colors[11][1] != 0x20 { // position 31 = space (padded)
		t.Errorf("LED11 G (pos 31) = 0x%02X, want space", colors[11][1])
	}
}

func TestBuildLEDColors_Full32(t *testing.T) {
	text := "0123456789ABCDEFGHIJKLMNOPQRSTUV" // exactly 32 chars
	colors := BuildLEDColors(text, 2, 16)

	if len(colors) != 12 {
		t.Fatalf("expected 12 color entries, got %d", len(colors))
	}

	// LED11: positions 30(R), 31(G), unused(B) → 'U' at 30, 'V' at 31
	if colors[11][0] != 'U' {
		t.Errorf("LED11 R (pos 30) = %c, want 'U'", colors[11][0])
	}
	if colors[11][1] != 'V' {
		t.Errorf("LED11 G (pos 31) = %c, want 'V'", colors[11][1])
	}
	if colors[11][2] != 0x00 { // unused B → 0x00 (never set)
		t.Errorf("LED11 B = 0x%02X, want 0x00 (unused)", colors[11][2])
	}
}

func TestSendLCDTextFullRoundTrip(t *testing.T) {
	// Simulate the full pipeline: full-screen text → state → packets
	text := "0123456789ABCDEFGHIJKLMNOPQRSTUV" // exactly 32 chars
	display := SplitToDisplay(text, 2, 16)
	if len([]rune(display)) != 32 {
		t.Fatalf("display len = %d, want 32", len([]rune(display)))
	}

	colors := BuildLEDColors(display, 2, 16)
	if len(colors) != 12 {
		t.Fatalf("BuildLEDColors returned %d entries, want 12", len(colors))
	}

	// Build full color array (as GetFullColors would)
	lcdColors := make([][3]byte, 0, 32)
	if len(colors) > 1 {
		lcdColors = append(lcdColors, colors[1:]...)
	}
	for len(lcdColors) < 32 {
		lcdColors = append(lcdColors, [3]byte{0, 0, 0})
	}

	// Build state
	var state State
	for i := 0; i < 32 && i < len(lcdColors); i++ {
		state.LCD[i] = lcdColors[i]
	}

	full := GetFullColors(&state)
	if len(full) != 33 {
		t.Fatalf("GetFullColors returned %d entries, want 33", len(full))
	}

	// Verify LED11 (strip index 11) has correct data
	// full[11] = state.LCD[10] = colors[11] = LED11 encoding
	if full[11][0] != 'U' { // position 30 (R) = 'U'
		t.Errorf("full[11].R = %c, want 'U'", full[11][0])
	}
	if full[11][1] != 'V' { // position 31 (G) = 'V'
		t.Errorf("full[11].G = %c, want 'V'", full[11][1])
	}

	// Build packets and verify they include LED11
	packets := BuildDirectPackets(0, full, 0)
	if len(packets) == 0 {
		t.Fatal("no packets generated")
	}

	// LED11 is at full index 11, which is in packet 1 (indices 0-19)
	// The LED data for LED11 starts at byte offset 5 + 11*3 = 38 in packet 1
	if len(packets) >= 1 {
		p := packets[0]
		led11Offset := 5 + 11*3
		if p[led11Offset] != 'U' {
			t.Errorf("packet1 LED11 R = %c, want 'U'", p[led11Offset])
		}
		if p[led11Offset+1] != 'V' {
			t.Errorf("packet1 LED11 G = %c, want 'V'", p[led11Offset+1])
		}
	}
}

func TestCharToLCDCode_ASCII(t *testing.T) {
	tests := []struct {
		input rune
		want  byte
	}{
		{'A', 0x41},
		{'z', 0x7A},
		{' ', 0x20},
		{'0', 0x30},
		{'9', 0x39},
		{'!', 0x21},
	}
	for _, tt := range tests {
		got := CharToLCDCode(tt.input)
		if got != tt.want {
			t.Errorf("CharToLCDCode(%q) = 0x%02X, want 0x%02X", tt.input, got, tt.want)
		}
	}
}

func TestCharToLCDCode_Cyrillic(t *testing.T) {
	tests := []struct {
		input rune
		want  byte
		name  string
	}{
		{'А', 0x41, "А (same as A)"},
		{'Б', 0xA0, "Б"},
		{'В', 0x42, "В (same as B)"},
		{'Г', 0xA1, "Г"},
		{'Д', 0xE0, "Д"},
		{'Е', 0x45, "Е (same as E)"},
		{'Ж', 0xA3, "Ж"},
		{'З', 0xA4, "З"},
		{'И', 0xA5, "И"},
		{'Й', 0xA6, "Й"},
		{'К', 0x4B, "К (same as K)"},
		{'Л', 0xA7, "Л"},
		{'М', 0x4D, "М (same as M)"},
		{'Н', 0x48, "Н (same as H)"},
		{'О', 0x4F, "О (same as O)"},
		{'П', 0xA8, "П"},
		{'Р', 0x50, "Р (same as P)"},
		{'С', 0x43, "С (same as C)"},
		{'Т', 0x54, "Т (same as T)"},
		{'У', 0xA9, "У"},
		{'Ф', 0xAA, "Ф"},
		{'Х', 0x58, "Х (same as X)"},
		{'Ц', 0xE1, "Ц"},
		{'Ч', 0xAB, "Ч"},
		{'Ш', 0xAC, "Ш"},
		{'Щ', 0xE2, "Щ"},
		{'Ъ', 0xAD, "Ъ"},
		{'Ы', 0xAE, "Ы"},
		{'Ь', 0x62, "Ь (same as b)"},
		{'Э', 0xAF, "Э"},
		{'Ю', 0xB0, "Ю"},
		{'Я', 0xB1, "Я"},
		{'а', 0x61, "а (same as a)"},
		{'б', 0xB2, "б"},
		{'в', 0xB3, "в"},
		{'г', 0xB4, "г"},
		{'д', 0xE3, "д"},
		{'е', 0x65, "е (same as e)"},
		{'ж', 0xB6, "ж"},
		{'з', 0xB7, "з"},
		{'и', 0xB8, "и"},
		{'й', 0xB9, "й"},
		{'к', 0xBA, "к"},
		{'л', 0xBB, "л"},
		{'м', 0xBC, "м"},
		{'н', 0xBD, "н"},
		{'о', 0x6F, "о (same as o)"},
		{'п', 0xBE, "п"},
		{'р', 0x70, "р (same as p)"},
		{'с', 0x63, "с (same as c)"},
		{'т', 0xBF, "т"},
		{'у', 0x79, "у (same as y)"},
		{'ф', 0xE4, "ф"},
		{'х', 0x78, "х (same as x)"},
		{'ц', 0xE5, "ц"},
		{'ч', 0xC0, "ч"},
		{'ш', 0xC1, "ш"},
		{'щ', 0xE6, "щ"},
		{'ъ', 0xC2, "ъ"},
		{'ы', 0xC3, "ы"},
		{'ь', 0xC4, "ь"},
		{'э', 0xC5, "э"},
		{'ю', 0xC6, "ю"},
		{'я', 0xC7, "я"},
		{'Ё', 0xA2, "Ё"},
		{'ё', 0xB5, "ё"},
	}
	for _, tt := range tests {
		got := CharToLCDCode(tt.input)
		if got != tt.want {
			t.Errorf("CharToLCDCode(%q) %s = 0x%02X, want 0x%02X", tt.input, tt.name, got, tt.want)
		}
	}
}

func TestBuildLEDColorsFromCodes(t *testing.T) {
	// 32 сырых кода CGROM, каждый равен своему индексу (0x00..0x1F).
	codes := make([]byte, 32)
	for i := range codes {
		codes[i] = byte(i)
	}
	colors := BuildLEDColorsFromCodes(codes, 2, 16)
	if len(colors) != 12 {
		t.Fatalf("BuildLEDColorsFromCodes len = %d, want 12", len(colors))
	}
	// Граница строк: код 15 -> LED6[0] (row0, посл. символ),
	// код 16 -> LED6[1] (row1, первый символ). Они в одном LED, подряд.
	if colors[6][0] != 0x0F {
		t.Errorf("code 15 -> colors[6][0] = 0x%02X, want 0x0F", colors[6][0])
	}
	if colors[6][1] != 0x10 {
		t.Errorf("code 16 -> colors[6][1] = 0x%02X, want 0x10", colors[6][1])
	}
	// Последний код 31 -> LED11[1] (row1, посл. символ).
	if colors[11][1] != 0x1F {
		t.Errorf("code 31 -> colors[11][1] = 0x%02X, want 0x1F", colors[11][1])
	}
	// Первый код 0 -> LED1[0] (row0, первый символ).
	if colors[1][0] != 0x00 {
		t.Errorf("code 0 -> colors[1][0] = 0x%02X, want 0x00", colors[1][0])
	}
}

func TestCharToLCDCode_UnknownPassesThrough(t *testing.T) {
	// После отмены фильтра любой символ пропускается как есть (младший байт),
	// чтобы на LCD можно было вывести любой код знакогенератора.
	unknowns := []rune{'€', '★', 'ä', 'ñ', '中'}
	for _, r := range unknowns {
		got := CharToLCDCode(r)
		want := byte(r & 0xFF)
		if got != want {
			t.Errorf("CharToLCDCode(%q) = 0x%02X, want 0x%02X (low byte passthrough)", r, got, want)
		}
	}
}
