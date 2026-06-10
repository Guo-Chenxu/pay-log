package service

import "github.com/Guo-Chenxu/pay-log/pkg/bizerr"

func validateYearMonth(year, month int) error {
	if month < 1 || month > 12 {
		return bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "month must be between 1 and 12")
	}
	return nil
}

func validateYearMonthRange(startYear, startMonth, endYear, endMonth int) error {
	if err := validateYearMonth(startYear, startMonth); err != nil {
		return err
	}
	if err := validateYearMonth(endYear, endMonth); err != nil {
		return err
	}
	if startYear*12+startMonth > endYear*12+endMonth {
		return bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "start month must be before or equal to end month")
	}
	return nil
}
