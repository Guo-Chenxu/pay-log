package service

import "testing"

func TestValidateAIRangeAcceptsChronologicalMonths(t *testing.T) {
	if err := validateAISummaryRange(2026, 1, 2026, 12); err != nil {
		t.Fatalf("validateAISummaryRange returned error: %v", err)
	}
	if err := validateAISummaryRange(2025, 12, 2026, 1); err != nil {
		t.Fatalf("validateAISummaryRange returned error across years: %v", err)
	}
}

func TestValidateAIServiceSummaryRangeRejectsInvalidRangeBeforeWork(t *testing.T) {
	if err := (&AIService{}).ValidateSummaryRange(2026, 13, 2026, 1); err == nil {
		t.Fatalf("ValidateSummaryRange returned nil error, want error")
	}
}

func TestValidateAISummaryRangeRejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		name       string
		startYear  int
		startMonth int
		endYear    int
		endMonth   int
	}{
		{name: "invalid start month low", startYear: 2026, startMonth: 0, endYear: 2026, endMonth: 1},
		{name: "invalid start month high", startYear: 2026, startMonth: 13, endYear: 2026, endMonth: 1},
		{name: "invalid end month low", startYear: 2026, startMonth: 1, endYear: 2026, endMonth: 0},
		{name: "invalid end month high", startYear: 2026, startMonth: 1, endYear: 2026, endMonth: 13},
		{name: "start after end", startYear: 2026, startMonth: 2, endYear: 2026, endMonth: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAISummaryRange(tt.startYear, tt.startMonth, tt.endYear, tt.endMonth); err == nil {
				t.Fatalf("validateAISummaryRange returned nil error, want error")
			}
		})
	}
}
