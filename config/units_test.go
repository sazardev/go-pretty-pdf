package config

import "testing"

func TestParsePaperSize(t *testing.T) {
	tests := []struct {
		name   string
		wantW  float64
		wantH  float64
		wantOK bool
	}{
		{PaperLetter, 8.5, 11, true},
		{"Legal", 8.5, 14, true},
		{"A4", 8.27, 11.69, true},
		{"tabloid", 0, 0, false},
		{"", 0, 0, false},
	}
	for _, tt := range tests {
		w, h, ok := ParsePaperSize(tt.name)
		if ok != tt.wantOK || w != tt.wantW || h != tt.wantH {
			t.Errorf("ParsePaperSize(%q) = (%v, %v, %v), want (%v, %v, %v)", tt.name, w, h, ok, tt.wantW, tt.wantH, tt.wantOK)
		}
	}
}

func TestParseCSSUnit(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"1in", 1},
		{"25.4mm", 1},
		{"2.54cm", 1},
		{"72pt", 1},
		{"96px", 1},
		{"10bogus", 0},
	}
	for _, tt := range tests {
		got := ParseCSSUnit(tt.in)
		diff := got - tt.want
		if diff < -0.001 || diff > 0.001 {
			t.Errorf("ParseCSSUnit(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
