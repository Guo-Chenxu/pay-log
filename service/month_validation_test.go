package service

import "testing"

func TestValidateYearMonthAcceptsValidMonth(t *testing.T) {
	if err := validateYearMonth(2026, 6); err != nil {
		t.Fatalf("validateYearMonth returned error: %v", err)
	}
}

func TestValidateYearMonthRejectsMonthOutsideRange(t *testing.T) {
	tests := []struct {
		name  string
		month int
	}{
		{name: "month zero", month: 0},
		{name: "month thirteen", month: 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateYearMonth(2026, tt.month); err == nil {
				t.Fatalf("validateYearMonth returned nil error, want error")
			}
		})
	}
}

func TestValidateYearMonthRangeRejectsInvalidRange(t *testing.T) {
	tests := []struct {
		name       string
		startYear  int
		startMonth int
		endYear    int
		endMonth   int
	}{
		{name: "start month zero", startYear: 2026, startMonth: 0, endYear: 2026, endMonth: 1},
		{name: "end month thirteen", startYear: 2026, startMonth: 1, endYear: 2026, endMonth: 13},
		{name: "start after end", startYear: 2026, startMonth: 7, endYear: 2026, endMonth: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateYearMonthRange(tt.startYear, tt.startMonth, tt.endYear, tt.endMonth); err == nil {
				t.Fatalf("validateYearMonthRange returned nil error, want error")
			}
		})
	}
}
