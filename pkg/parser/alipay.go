package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ParseAlipayCSV parses a GBK-encoded Alipay CSV bill file.
func ParseAlipayCSV(path string) ([]ParsedRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := transform.NewReader(f, simplifiedchinese.GBK.NewDecoder())
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Skip lines until separator line (starts with "---"), then skip the column header line
	var headerFound bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "---") {
			headerFound = true
			if !scanner.Scan() { // consume the column header line
				if err := scanner.Err(); err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("alipay: header line not found")
			}
			break
		}
	}
	if !headerFound {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("alipay: separator line not found")
	}

	// Collect remaining lines as CSV data
	var sb strings.Builder
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	r := csv.NewReader(strings.NewReader(sb.String()))
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	var records []ParsedRecord
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 11 {
			continue
		}
		rec, err := parseAlipayRow(row)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func parseAlipayRow(row []string) (ParsedRecord, error) {
	txTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(row[0]), time.Local)
	if err != nil {
		return ParsedRecord{}, err
	}
	amount, err := money.YuanStringToCents(strings.TrimSpace(row[6]))
	if err != nil {
		return ParsedRecord{}, err
	}
	billType := parseBillType(strings.TrimSpace(row[5]))
	description := strings.TrimSpace(row[4])
	isInvestment := billType == int8(consts.BillTypeNeutral) && strings.Contains(description, "买入")
	return ParsedRecord{
		TransactionTime: txTime,
		Category:        strings.TrimSpace(row[1]),
		Counterparty:    strings.TrimSpace(row[2]),
		Description:     description,
		BillType:        billType,
		Amount:          amount,
		PaymentMethod:   strings.TrimSpace(row[7]),
		Status:          strings.TrimSpace(row[8]),
		OrderNo:         strings.TrimSpace(row[9]),
		MerchantOrderNo: strings.TrimSpace(row[10]),
		IsInvestment:    isInvestment,
	}, nil
}

func parseBillType(s string) int8 {
	switch s {
	case "收入":
		return int8(consts.BillTypeIncome)
	case "支出":
		return int8(consts.BillTypeExpense)
	default:
		return int8(consts.BillTypeNeutral)
	}
}
