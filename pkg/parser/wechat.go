package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/Guo-Chenxu/pay-log/pkg/money"
	"github.com/xuri/excelize/v2"
)

// ParseWechatXLSX parses a WeChat Pay XLSX bill file.
// The first 16 rows are header/comments; row 17 (index 16) is column header; data starts at row 18 (index 17).
func ParseWechatXLSX(path string) ([]ParsedRecord, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("wechat: no sheets found")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	const dataStartRow = 17 // 0-indexed: skip rows 0-16 (header + column header)
	var records []ParsedRecord
	for i := dataStartRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 9 {
			continue
		}
		rec, err := parseWechatRow(row)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func parseWechatRow(row []string) (ParsedRecord, error) {
	if len(row) < 9 {
		return ParsedRecord{}, fmt.Errorf("wechat: row has %d columns, want at least 9", len(row))
	}
	txTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(row[0]), time.Local)
	if err != nil {
		return ParsedRecord{}, err
	}
	amtStr := strings.TrimSpace(row[5])
	amtStr = strings.TrimPrefix(amtStr, "¥")
	amtStr = strings.ReplaceAll(amtStr, ",", "")
	amount, err := money.YuanStringToCents(amtStr)
	if err != nil {
		return ParsedRecord{}, err
	}
	billType := parseBillType(strings.TrimSpace(row[4]))
	merchantOrderNo := ""
	if len(row) > 9 {
		merchantOrderNo = strings.TrimSpace(row[9])
	}
	remark := ""
	if len(row) > 10 {
		remark = strings.TrimSpace(row[10])
	}
	return ParsedRecord{
		TransactionTime: txTime,
		Category:        strings.TrimSpace(row[1]),
		Counterparty:    strings.TrimSpace(row[2]),
		Description:     strings.TrimSpace(row[3]),
		BillType:        billType,
		Amount:          amount,
		PaymentMethod:   strings.TrimSpace(row[6]),
		Status:          strings.TrimSpace(row[7]),
		OrderNo:         strings.TrimSpace(row[8]),
		MerchantOrderNo: merchantOrderNo,
		Remark:          remark,
	}, nil
}
