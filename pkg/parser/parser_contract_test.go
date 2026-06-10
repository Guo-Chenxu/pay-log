package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/xuri/excelize/v2"
)

func TestParseAlipayRowUsesCentsConstantsAndInvestmentBool(t *testing.T) {
	row := []string{
		"2026-06-11 21:30:00",
		"投资理财",
		"基金公司",
		"",
		"买入基金",
		"不计收支",
		"1000.05",
		"余额宝",
		"交易成功",
		"order-1",
		"merchant-1",
	}

	rec, err := parseAlipayRow(row)
	if err != nil {
		t.Fatalf("parseAlipayRow returned error: %v", err)
	}
	if rec.Amount != 100005 {
		t.Fatalf("Amount = %d, want 100005 cents", rec.Amount)
	}
	if rec.BillType != int8(consts.BillTypeNeutral) {
		t.Fatalf("BillType = %d, want neutral", rec.BillType)
	}
	if !rec.IsInvestment {
		t.Fatalf("IsInvestment = false, want true")
	}
}

func TestParseWechatRowUsesCentsAndConstants(t *testing.T) {
	row := []string{
		"2026-06-11 08:00:00",
		"餐饮",
		"早餐店",
		"早餐",
		"支出",
		"¥1,234.56",
		"零钱",
		"支付成功",
		"order-2",
		"merchant-2",
		"备注",
	}

	rec, err := parseWechatRow(row)
	if err != nil {
		t.Fatalf("parseWechatRow returned error: %v", err)
	}
	if rec.Amount != 123456 {
		t.Fatalf("Amount = %d, want 123456 cents", rec.Amount)
	}
	if rec.BillType != int8(consts.BillTypeExpense) {
		t.Fatalf("BillType = %d, want expense", rec.BillType)
	}
	if rec.IsInvestment {
		t.Fatalf("IsInvestment = true, want false")
	}
}

func TestParseWechatAcceptsRowsWithoutTrailingRemark(t *testing.T) {
	row := []string{
		"2026-06-11 08:00:00",
		"餐饮",
		"早餐店",
		"早餐",
		"支出",
		"¥1,234.56",
		"零钱",
		"支付成功",
		"order-2",
		"merchant-2",
	}

	rec, err := parseWechatRow(row)
	if err != nil {
		t.Fatalf("parseWechatRow returned error: %v", err)
	}
	if rec.Remark != "" {
		t.Fatalf("Remark = %q, want empty", rec.Remark)
	}

	path := writeWechatXLSX(t, row)
	recs, err := ParseWechatXLSX(path)
	if err != nil {
		t.Fatalf("ParseWechatXLSX returned error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("ParseWechatXLSX returned %d records, want 1", len(recs))
	}
}

func TestParseWechatAcceptsRowsWithoutTrailingMerchantOrderNoAndRemark(t *testing.T) {
	row := []string{
		"2026-06-11 08:00:00",
		"餐饮",
		"早餐店",
		"早餐",
		"支出",
		"¥1,234.56",
		"零钱",
		"支付成功",
		"order-2",
	}

	rec, err := parseWechatRow(row)
	if err != nil {
		t.Fatalf("parseWechatRow returned error: %v", err)
	}
	if rec.OrderNo != "order-2" {
		t.Fatalf("OrderNo = %q, want order-2", rec.OrderNo)
	}
	if rec.MerchantOrderNo != "" {
		t.Fatalf("MerchantOrderNo = %q, want empty", rec.MerchantOrderNo)
	}
	if rec.Remark != "" {
		t.Fatalf("Remark = %q, want empty", rec.Remark)
	}

	path := writeWechatXLSX(t, row)
	recs, err := ParseWechatXLSX(path)
	if err != nil {
		t.Fatalf("ParseWechatXLSX returned error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("ParseWechatXLSX returned %d records, want 1", len(recs))
	}
}

func TestParseAlipayCSVReturnsErrorWhenSeparatorHasNoHeader(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "alipay-*.csv")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	if _, err := file.WriteString("------------------------\n"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	recs, err := ParseAlipayCSV(file.Name())
	if err == nil {
		t.Fatalf("ParseAlipayCSV returned nil error and %d records, want malformed format error", len(recs))
	}
	if !strings.Contains(err.Error(), "header") {
		t.Fatalf("ParseAlipayCSV error = %q, want header error", err.Error())
	}
}

func TestParseAlipayCSVSurfacesScannerErrorBeforeSeparator(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "alipay-*.csv")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	if _, err := file.WriteString(strings.Repeat("x", 1024*1024+1)); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	recs, err := ParseAlipayCSV(file.Name())
	if err == nil {
		t.Fatalf("ParseAlipayCSV returned nil error and %d records, want scanner error", len(recs))
	}
	if strings.Contains(err.Error(), "separator line not found") {
		t.Fatalf("ParseAlipayCSV error = %q, want scanner error", err.Error())
	}
}

func writeWechatXLSX(t *testing.T, dataRow []string) string {
	t.Helper()

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("Close xlsx returned error: %v", err)
		}
	}()

	sheet := f.GetSheetName(0)
	for col, value := range dataRow {
		cell, err := excelize.CoordinatesToCellName(col+1, 18)
		if err != nil {
			t.Fatalf("CoordinatesToCellName returned error: %v", err)
		}
		if err := f.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("SetCellValue returned error: %v", err)
		}
	}

	path := t.TempDir() + "/wechat.xlsx"
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs returned error: %v", err)
	}
	return path
}
