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

	// LED1 R = 'A' (index 44), G = 'B' (index 45), B = 'C' (index 46)
	if colors[1][0] != 44 {
		t.Errorf("LED1 R = %d, want 44 (index of 'A')", colors[1][0])
	}
	if colors[1][1] != 45 {
		t.Errorf("LED1 G = %d, want 45 (index of 'B')", colors[1][1])
	}
	if colors[1][2] != 46 {
		t.Errorf("LED1 B = %d, want 46 (index of 'C')", colors[1][2])
	}

	// LED6: positions 15(R), 16(G), 17(B) — row boundary!
	// With SplitToDisplay padding: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123  "
	// Pos 15 = 'P' (index 59), Pos 16 = 'Q' (index 60), Pos 17 = 'R' (index 61)
	if colors[6][0] != 59 {
		t.Errorf("LED6 R = %d, want 59 (index of 'P', pos 15)", colors[6][0])
	}
	if colors[6][1] != 60 {
		t.Errorf("LED6 G = %d, want 60 (index of 'Q', pos 16)", colors[6][1])
	}
	if colors[6][2] != 61 {
		t.Errorf("LED6 B = %d, want 61 (index of 'R', pos 17)", colors[6][2])
	}

	// Positions 29,30,31: '3' at 29, spaces at 30 and 31
	// '3' → index 30, space → index 0
	if colors[10][2] != 30 { // position 29 = '3'
		t.Errorf("LED10 B (pos 29) = %d, want 30", colors[10][2])
	}
	if colors[11][0] != 0 { // position 30 = space (padded)
		t.Errorf("LED11 R (pos 30) = 0x%02X, want 0 (space)", colors[11][0])
	}
	if colors[11][1] != 0 { // position 31 = space (padded)
		t.Errorf("LED11 G (pos 31) = 0x%02X, want 0 (space)", colors[11][1])
	}
}

func TestBuildLEDColors_Full32(t *testing.T) {
	text := "0123456789ABCDEFGHIJKLMNOPQRSTUV" // exactly 32 chars
	colors := BuildLEDColors(text, 2, 16)

	if len(colors) != 12 {
		t.Fatalf("expected 12 color entries, got %d", len(colors))
	}

	// LED11: positions 30(R), 31(G), unused(B) → 'U' at 30 (index 64), 'V' at 31 (index 65)
	if colors[11][0] != 64 {
		t.Errorf("LED11 R (pos 30) = %d, want 64 (index of 'U')", colors[11][0])
	}
	if colors[11][1] != 65 {
		t.Errorf("LED11 G (pos 31) = %d, want 65 (index of 'V')", colors[11][1])
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
	if full[11][0] != 64 { // position 30 (R) = 'U' (index 64)
		t.Errorf("full[11].R = %d, want 64", full[11][0])
	}
	if full[11][1] != 65 { // position 31 (G) = 'V' (index 65)
		t.Errorf("full[11].G = %d, want 65", full[11][1])
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
		if p[led11Offset] != 64 {
			t.Errorf("packet1 LED11 R = %d, want 64", p[led11Offset])
		}
		if p[led11Offset+1] != 65 {
			t.Errorf("packet1 LED11 G = %d, want 65", p[led11Offset+1])
		}
	}
}

func TestCharToLCDIndex_ASCII(t *testing.T) {
	tests := []struct {
		input rune
		want  byte
	}{
		{'A', 44},
		{'z', 101},
		{' ', 0},
		{'0', 27},
		{'9', 36},
		{'!', 12},
	}
	for _, tt := range tests {
		got := CharToLCDIndex(tt.input)
		if got != tt.want {
			t.Errorf("CharToLCDIndex(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCharToLCDIndex_Cyrillic(t *testing.T) {
	tests := []struct {
		input rune
		want  byte
		name  string
	}{
		{'А', 44, "А (same as A)"},
		{'Б', 109, "Б"},
		{'В', 45, "В (same as B)"},
		{'Г', 110, "Г"},
		{'Д', 152, "Д"},
		{'Е', 48, "Е (same as E)"},
		{'Ж', 112, "Ж"},
		{'З', 113, "З"},
		{'И', 114, "И"},
		{'Й', 115, "Й"},
		{'К', 54, "К (same as K)"},
		{'Л', 116, "Л"},
		{'М', 56, "М (same as M)"},
		{'Н', 51, "Н (same as H)"},
		{'О', 58, "О (same as O)"},
		{'П', 117, "П"},
		{'Р', 59, "Р (same as P)"},
		{'С', 46, "С (same as C)"},
		{'Т', 63, "Т (same as T)"},
		{'У', 118, "У"},
		{'Ф', 119, "Ф"},
		{'Х', 67, "Х (same as X)"},
		{'Ц', 153, "Ц"},
		{'Ч', 120, "Ч"},
		{'Ш', 121, "Ш"},
		{'Щ', 154, "Щ"},
		{'Ъ', 122, "Ъ"},
		{'Ы', 123, "Ы"},
		{'Ь', 77, "Ь (same as b)"},
		{'Э', 124, "Э"},
		{'Ю', 125, "Ю"},
		{'Я', 126, "Я"},
		{'а', 76, "а (same as a)"},
		{'б', 127, "б"},
		{'в', 128, "в"},
		{'г', 129, "г"},
		{'д', 155, "д"},
		{'е', 80, "е (same as e)"},
		{'ж', 131, "ж"},
		{'з', 132, "з"},
		{'и', 133, "и"},
		{'й', 134, "й"},
		{'к', 135, "к"},
		{'л', 136, "л"},
		{'м', 137, "м"},
		{'н', 138, "н"},
		{'о', 90, "о (same as o)"},
		{'п', 139, "п"},
		{'р', 91, "р (same as p)"},
		{'с', 78, "с (same as c)"},
		{'т', 140, "т"},
		{'у', 100, "у (same as y)"},
		{'ф', 156, "ф"},
		{'х', 99, "х (same as x)"},
		{'ц', 157, "ц"},
		{'ч', 141, "ч"},
		{'ш', 142, "ш"},
		{'щ', 158, "щ"},
		{'ъ', 143, "ъ"},
		{'ы', 144, "ы"},
		{'ь', 145, "ь"},
		{'э', 146, "э"},
		{'ю', 147, "ю"},
		{'я', 148, "я"},
		{'Ё', 111, "Ё"},
		{'ё', 130, "ё"},
	}
	for _, tt := range tests {
		got := CharToLCDIndex(tt.input)
		if got != tt.want {
			t.Errorf("CharToLCDIndex(%q) %s = %d, want %d", tt.input, tt.name, got, tt.want)
		}
	}
}

func TestBuildLEDColorsFromCodes(t *testing.T) {
	// Test with valid CGROM codes that map to indices
	// Using CGROM codes: 0x41('A'→idx 44), 0x42('B'→idx 45), 0x43('C'→idx 46), etc.
	codes := []byte{
		0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A,
		0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, // Row 0: A-P
		0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5A,
		0x5B, 0x5C, 0x5D, 0x5E, 0x5F, 0x60, // Row 1: Q-`
	}
	colors := BuildLEDColorsFromCodes(codes, 2, 16)
	if len(colors) != 12 {
		t.Fatalf("BuildLEDColorsFromCodes len = %d, want 12", len(colors))
	}
	// First LED (LED1) should have 'A'(44), 'B'(45), 'C'(46)
	if colors[1][0] != 44 {
		t.Errorf("code 0x41('A') -> colors[1][0] = %d, want 44", colors[1][0])
	}
	if colors[1][1] != 45 {
		t.Errorf("code 0x42('B') -> colors[1][1] = %d, want 45", colors[1][1])
	}
	if colors[1][2] != 46 {
		t.Errorf("code 0x43('C') -> colors[1][2] = %d, want 46", colors[1][2])
	}
	// Row boundary: code 0x50('P', idx 59) -> colors[6][0]
	// Row 1 starts with 0x51('Q', idx 60) -> colors[6][1]
	if colors[6][0] != 59 {
		t.Errorf("code 0x50('P') -> colors[6][0] = %d, want 59", colors[6][0])
	}
	if colors[6][1] != 60 {
		t.Errorf("code 0x51('Q') -> colors[6][1] = %d, want 60", colors[6][1])
	}
	// Last code 0x60(`, idx 75) -> colors[11][1] (row1, last char)
	if colors[11][1] != 75 {
		t.Errorf("code 0x60 -> colors[11][1] = %d, want 75", colors[11][1])
	}
	// First code 0x41 -> colors[1][0]
	if colors[1][0] != 44 {
		t.Errorf("code 0x41 -> colors[1][0] = %d, want 44", colors[1][0])
	}
}

func TestCharToLCDIndex_UnknownReturnsSpace(t *testing.T) {
	// Unknown characters should return index 0 (space)
	unknowns := []rune{'€', '★', 'ä', 'ñ', '中'}
	for _, r := range unknowns {
		got := CharToLCDIndex(r)
		if got != 0 {
			t.Errorf("CharToLCDIndex(%q) = %d, want 0 (space)", r, got)
		}
	}
}

func TestCGROMToIndex(t *testing.T) {
	// Test known CGROM codes → indices
	tests := []struct {
		cgrom byte
		index byte
	}{
		{0x20, 0},   // space
		{0x41, 44},  // 'A'
		{0x42, 45},  // 'B'
		{0x43, 46},  // 'C'
		{0x50, 59},  // 'P'
		{0x61, 76},  // 'a'
		{0xA0, 109}, // 'Б'
		{0xA1, 110}, // 'Г'
		{0xA2, 111}, // 'Ё'
		{0xFF, 160}, // special char
	}
	for _, tt := range tests {
		got := CGROMToIndex(tt.cgrom)
		if got != tt.index {
			t.Errorf("CGROMToIndex(0x%02X) = %d, want %d", tt.cgrom, got, tt.index)
		}
	}
}

func TestCGROMToIndex_Unknown(t *testing.T) {
	// Unknown CGROM codes should return LCD_CGromSize (out of range)
	unknownCodes := []byte{0x0F, 0x10, 0x00, 0x01, 0xFE}
	for _, code := range unknownCodes {
		got := CGROMToIndex(code)
		if got != LCD_CGromSize {
			t.Errorf("CGROMToIndex(0x%02X) = %d, want %d (out of range)", code, got, LCD_CGromSize)
		}
	}
}
