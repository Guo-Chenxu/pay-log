package money

import (
	"math"
	"testing"
)

func TestYuanStringToCents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "integer yuan", input: "12", want: 1200},
		{name: "one decimal", input: "12.3", want: 1230},
		{name: "two decimals", input: "12.34", want: 1234},
		{name: "zero", input: "0", want: 0},
		{name: "leading and trailing spaces", input: " 7.05 ", want: 705},
		{name: "max int64 cents", input: "92233720368547758.07", want: math.MaxInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YuanStringToCents(tt.input)
			if err != nil {
				t.Fatalf("YuanStringToCents(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("YuanStringToCents(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestYuanStringToCentsRejectsInvalidInput(t *testing.T) {
	inputs := []string{"", " ", "-1", "1.234", "abc", "1.", ".1", "1,234.56", "92233720368547758.08"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if got, err := YuanStringToCents(input); err == nil {
				t.Fatalf("YuanStringToCents(%q) = %d, want error", input, got)
			}
		})
	}
}

func TestCentsToYuanString(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "zero", input: 0, want: "0.00"},
		{name: "single digit cents", input: 5, want: "0.05"},
		{name: "one yuan", input: 100, want: "1.00"},
		{name: "yuan and cents", input: 1234, want: "12.34"},
		{name: "negative defensive formatting", input: -1234, want: "-12.34"},
		{name: "minimum int64 defensive formatting", input: math.MinInt64, want: "-92233720368547758.08"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CentsToYuanString(tt.input); got != tt.want {
				t.Fatalf("CentsToYuanString(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
