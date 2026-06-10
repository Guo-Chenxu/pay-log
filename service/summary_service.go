package service

import (
	"context"
	"sync"
	"time"

	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/pageutil"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
	voresp "github.com/Guo-Chenxu/pay-log/vo/resp"
)

type SummaryService struct {
	billMgr  *manager.BillRecordManager
	monthMgr *manager.MonthSummaryManager
}

var (
	sumSvc     *SummaryService
	sumSvcOnce sync.Once
)

func NewSummaryService() *SummaryService {
	sumSvcOnce.Do(func() {
		sumSvc = &SummaryService{
			billMgr:  manager.NewBillRecordManager(),
			monthMgr: manager.NewMonthSummaryManager(),
		}
	})
	return sumSvc
}

func (s *SummaryService) SummarizeMonth(ctx context.Context, userID int64, year, month int) error {
	if err := validateYearMonth(year, month); err != nil {
		return err
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	agg, err := s.billMgr.AggregatePeriod(userID, start, end)
	if err != nil {
		return err
	}
	return s.monthMgr.Upsert(&model.MonthSummary{
		ID: snowflake.GenerateID(), UserID: userID, Year: year, Month: month,
		TotalIncome: agg.TotalIncome, TotalExpense: agg.TotalExpense,
		AlipayIncome: agg.AlipayIncome, AlipayExpense: agg.AlipayExpense,
		WechatIncome: agg.WechatIncome, WechatExpense: agg.WechatExpense,
		InvestmentAmount: agg.InvestmentAmount,
	})
}

// SummarizeAllUnsummarized summarizes every month that has bill records but no cached summary yet.
func (s *SummaryService) SummarizeAllUnsummarized(ctx context.Context, userID int64) error {
	months, err := s.billMgr.ListDistinctMonths(userID)
	if err != nil {
		return err
	}
	summarized, err := s.monthMgr.ListSummarizedMonths(userID)
	if err != nil {
		return err
	}
	for _, ym := range months {
		if summarized[ym.Year] != nil && summarized[ym.Year][ym.Month] {
			continue
		}
		if err := s.SummarizeMonth(ctx, userID, ym.Year, ym.Month); err != nil {
			return err
		}
	}
	return nil
}

func (s *SummaryService) GetMonthOverview(ctx context.Context, userID int64, page, pageSize int) (*response.PaginatedData[*voresp.MonthSummaryResp], error) {
	page, pageSize = pageutil.NormalizePage(page, pageSize)

	allMonths, err := s.billMgr.ListDistinctMonths(userID)
	if err != nil {
		return nil, err
	}
	total := int64(len(allMonths))

	offset := (page - 1) * pageSize
	if offset >= int(total) {
		return &response.PaginatedData[*voresp.MonthSummaryResp]{
			Items: []*voresp.MonthSummaryResp{}, Page: page, PageSize: pageSize,
			Total: total, HasNext: false,
		}, nil
	}
	end := offset + pageSize
	if end > int(total) {
		end = int(total)
	}
	pageMonths := allMonths[offset:end]

	// Fetch cached summaries for this page's range (months sorted DESC, so last=oldest)
	startYM := pageMonths[len(pageMonths)-1]
	endYM := pageMonths[0]
	cached, err := s.monthMgr.ListByRange(userID, startYM.Year, startYM.Month, endYM.Year, endYM.Month)
	if err != nil {
		return nil, err
	}
	cachedMap := make(map[int]map[int]*model.MonthSummary)
	for _, c := range cached {
		if cachedMap[c.Year] == nil {
			cachedMap[c.Year] = make(map[int]*model.MonthSummary)
		}
		cachedMap[c.Year][c.Month] = c
	}

	items := make([]*voresp.MonthSummaryResp, 0, len(pageMonths))
	for _, ym := range pageMonths {
		if m, ok := cachedMap[ym.Year][ym.Month]; ok {
			items = append(items, voresp.FromMonthSummary(m))
			continue
		}
		// No cache — aggregate on the fly
		mStart := time.Date(ym.Year, time.Month(ym.Month), 1, 0, 0, 0, 0, time.Local)
		agg, err := s.billMgr.AggregatePeriod(userID, mStart, mStart.AddDate(0, 1, 0))
		if err != nil {
			return nil, err
		}
		items = append(items, voresp.FromPeriodAggMonthSummary(ym.Year, ym.Month, agg))
	}

	return &response.PaginatedData[*voresp.MonthSummaryResp]{
		Items: items, Page: page, PageSize: pageSize,
		Total: total, HasNext: int64(page*pageSize) < total,
	}, nil
}

func (s *SummaryService) GetMonthDetail(ctx context.Context, userID int64, year, month int) (*voresp.MonthSummaryResp, error) {
	if err := validateYearMonth(year, month); err != nil {
		return nil, err
	}
	m, err := s.monthMgr.Get(userID, year, month)
	if err == nil {
		return voresp.FromMonthSummary(m), nil
	}
	// No cache — aggregate on the fly
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	agg, err := s.billMgr.AggregatePeriod(userID, start, start.AddDate(0, 1, 0))
	if err != nil {
		return nil, err
	}
	return voresp.FromPeriodAggMonthSummary(year, month, agg), nil
}

func (s *SummaryService) GetRangeDetail(ctx context.Context, userID int64, startYear, startMonth, endYear, endMonth, page, pageSize int) (*voresp.RangeDetailResp, error) {
	if err := validateYearMonthRange(startYear, startMonth, endYear, endMonth); err != nil {
		return nil, err
	}
	start := time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(endYear, time.Month(endMonth), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)

	agg, err := s.billMgr.AggregatePeriod(userID, start, endTime)
	if err != nil {
		return nil, err
	}

	// Per-month breakdown: prefer cached, fall back to real-time
	billMonths, err := s.billMgr.ListDistinctMonthsInRange(userID, start, endTime)
	if err != nil {
		return nil, err
	}
	cached, err := s.monthMgr.ListByRange(userID, startYear, startMonth, endYear, endMonth)
	if err != nil {
		return nil, err
	}
	cachedMap := make(map[int]map[int]*model.MonthSummary)
	for _, c := range cached {
		if cachedMap[c.Year] == nil {
			cachedMap[c.Year] = make(map[int]*model.MonthSummary)
		}
		cachedMap[c.Year][c.Month] = c
	}
	months := make([]*voresp.MonthSummaryResp, 0, len(billMonths))
	for _, ym := range billMonths {
		if m, ok := cachedMap[ym.Year][ym.Month]; ok {
			months = append(months, voresp.FromMonthSummary(m))
			continue
		}
		mStart := time.Date(ym.Year, time.Month(ym.Month), 1, 0, 0, 0, 0, time.Local)
		mAgg, err := s.billMgr.AggregatePeriod(userID, mStart, mStart.AddDate(0, 1, 0))
		if err != nil {
			return nil, err
		}
		months = append(months, voresp.FromPeriodAggMonthSummary(ym.Year, ym.Month, mAgg))
	}

	return &voresp.RangeDetailResp{
		Agg:    voresp.FromPeriodAgg(agg),
		Months: months,
	}, nil
}
