package service

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
	"github.com/Guo-Chenxu/pay-log/pkg/pageutil"
	"github.com/Guo-Chenxu/pay-log/pkg/parser"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
	voresp "github.com/Guo-Chenxu/pay-log/vo/resp"
)

type BillService struct {
	billMgr *manager.BillRecordManager
}

var (
	billSvc     *BillService
	billSvcOnce sync.Once
)

func NewBillService() *BillService {
	billSvcOnce.Do(func() {
		billSvc = &BillService{billMgr: manager.NewBillRecordManager()}
	})
	return billSvc
}

// ImportFile parses the temp file at path for the given channel (1=alipay, 2=wechat),
// inserts non-duplicate records, then deletes the temp file.
func (s *BillService) ImportFile(ctx context.Context, userID int64, path string, channel int8) (*voresp.UploadResultResp, error) {
	defer os.Remove(path)

	if !consts.IsValidChannel(channel) {
		return nil, bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "channel must be 1 (alipay) or 2 (wechat)")
	}

	var records []parser.ParsedRecord
	var err error
	if channel == int8(consts.ChannelAlipay) {
		records, err = parser.ParseAlipayCSV(path)
	} else {
		records, err = parser.ParseWechatXLSX(path)
	}
	if err != nil {
		return nil, err
	}

	result := &voresp.UploadResultResp{}
	for _, r := range records {
		exists, err := s.billMgr.Exists(userID, channel, r.TransactionTime, r.Amount, r.Description)
		if err != nil {
			result.Failed++
			continue
		}
		if exists {
			result.Skipped++
			continue
		}
		rec := &model.BillRecord{
			ID:              snowflake.GenerateID(),
			UserID:          userID,
			Channel:         channel,
			TransactionTime: r.TransactionTime,
			Counterparty:    r.Counterparty,
			Description:     r.Description,
			Category:        r.Category,
			BillType:        r.BillType,
			Amount:          r.Amount,
			PaymentMethod:   r.PaymentMethod,
			Status:          r.Status,
			OrderNo:         r.OrderNo,
			MerchantOrderNo: r.MerchantOrderNo,
			Remark:          r.Remark,
			IsInvestment:    r.IsInvestment,
		}
		if err = s.billMgr.Create(rec); err != nil {
			result.Failed++
		} else {
			result.Success++
		}
	}
	return result, nil
}

type normalizedManualBill struct {
	Amount       int64
	BillType     int8
	IsInvestment bool
}

func normalizeManualBill(req voreq.AddManualBillReq) (normalizedManualBill, error) {
	if !consts.IsValidChannel(req.Channel) {
		return normalizedManualBill{}, bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "channel must be 1 (alipay) or 2 (wechat)")
	}
	if !consts.IsValidBillType(req.BillType) {
		return normalizedManualBill{}, bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "bill_type must be 1 (income), 2 (expense), or 3 (neutral)")
	}
	amount, err := money.YuanStringToCents(req.Amount)
	if err != nil {
		return normalizedManualBill{}, bizerr.NewCustomErrorWithExtra(bizerr.ParamException.Code, "amount must be a non-negative yuan string with up to 2 decimals")
	}
	billType := req.BillType
	if req.IsInvestment {
		billType = int8(consts.BillTypeNeutral)
	}
	return normalizedManualBill{Amount: amount, BillType: billType, IsInvestment: req.IsInvestment}, nil
}

func (s *BillService) AddManual(ctx context.Context, userID int64, req voreq.AddManualBillReq) error {
	normalized, err := normalizeManualBill(req)
	if err != nil {
		return err
	}
	rec := &model.BillRecord{
		ID:              snowflake.GenerateID(),
		UserID:          userID,
		Channel:         req.Channel,
		TransactionTime: req.TransactionTime,
		Counterparty:    req.Counterparty,
		Description:     req.Description,
		Category:        req.Category,
		BillType:        normalized.BillType,
		Amount:          normalized.Amount,
		PaymentMethod:   req.PaymentMethod,
		Remark:          req.Remark,
		IsInvestment:    normalized.IsInvestment,
	}
	return s.billMgr.Create(rec)
}

func (s *BillService) ListByMonth(ctx context.Context, userID int64, year, month, page, pageSize int, sortOrder string) (*response.PaginatedData[*voresp.BillRecordResp], error) {
	if err := validateYearMonth(year, month); err != nil {
		return nil, err
	}
	page, pageSize = pageutil.NormalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	items, total, err := s.billMgr.ListByMonth(userID, year, month, offset, pageSize, sortOrder)
	if err != nil {
		return nil, err
	}
	return &response.PaginatedData[*voresp.BillRecordResp]{
		Items: voresp.FromBillRecords(items), Page: page, PageSize: pageSize,
		Total: total, HasNext: int64(page*pageSize) < total,
	}, nil
}

func (s *BillService) ListByRange(ctx context.Context, userID int64, startYear, startMonth, endYear, endMonth, page, pageSize int, sortOrder string) (*response.PaginatedData[*voresp.BillRecordResp], error) {
	if err := validateYearMonthRange(startYear, startMonth, endYear, endMonth); err != nil {
		return nil, err
	}
	page, pageSize = pageutil.NormalizePage(page, pageSize)
	start := time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
	end := time.Date(endYear, time.Month(endMonth), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	offset := (page - 1) * pageSize
	items, total, err := s.billMgr.ListByRange(userID, start, end, offset, pageSize, sortOrder)
	if err != nil {
		return nil, err
	}
	return &response.PaginatedData[*voresp.BillRecordResp]{
		Items: voresp.FromBillRecords(items), Page: page, PageSize: pageSize,
		Total: total, HasNext: int64(page*pageSize) < total,
	}, nil
}

func (s *BillService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.billMgr.SoftDelete(id, userID)
}
