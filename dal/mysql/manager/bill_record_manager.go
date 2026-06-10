package manager

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/Guo-Chenxu/pay-log/consts"
	mysqldb "github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

type BillRecordManager struct{ db *gorm.DB }

var (
	billMgr     *BillRecordManager
	billMgrOnce sync.Once
)

func NewBillRecordManager() *BillRecordManager {
	billMgrOnce.Do(func() { billMgr = &BillRecordManager{db: mysqldb.GetDB()} })
	return billMgr
}

func (m *BillRecordManager) Exists(userID int64, channel int8, txTime time.Time, amount int64, description string) (bool, error) {
	var count int64
	err := m.db.Model(&model.BillRecord{}).
		Where("user_id=? AND channel=? AND transaction_time=? AND amount=? AND description=? AND is_deleted = ?",
			userID, channel, txTime, amount, description, false).
		Count(&count).Error
	return count > 0, err
}

func (m *BillRecordManager) Create(r *model.BillRecord) error {
	return m.db.Create(r).Error
}

func (m *BillRecordManager) ListByMonth(userID int64, year, month, offset, limit int, sortOrder string) ([]*model.BillRecord, int64, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	var records []*model.BillRecord
	var total int64
	q := m.db.Model(&model.BillRecord{}).
		Where("user_id=? AND transaction_time>=? AND transaction_time<? AND is_deleted = ?", userID, start, end, false)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "transaction_time DESC"
	switch sortOrder {
	case "asc":
		order = "amount ASC"
	case "desc":
		order = "amount DESC"
	}
	err := q.Order(order).Offset(offset).Limit(limit).Find(&records).Error
	return records, total, err
}

func (m *BillRecordManager) ListByRange(userID int64, start, end time.Time, offset, limit int, sortOrder ...string) ([]*model.BillRecord, int64, error) {
	var records []*model.BillRecord
	var total int64
	q := m.db.Model(&model.BillRecord{}).
		Where("user_id=? AND transaction_time>=? AND transaction_time<? AND is_deleted = ?", userID, start, end, false)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "transaction_time DESC"
	if len(sortOrder) > 0 && sortOrder[0] == "asc" {
		order = "amount ASC"
	} else if len(sortOrder) > 0 && sortOrder[0] == "desc" {
		order = "amount DESC"
	}
	err := q.Order(order).Offset(offset).Limit(limit).Find(&records).Error
	return records, total, err
}

func (m *BillRecordManager) SoftDelete(id, userID int64) error {
	return m.db.Model(&model.BillRecord{}).
		Where("id=? AND user_id=?", id, userID).
		Update("is_deleted", true).Error
}

// YearMonth is a year+month pair from a GROUP BY query.
type YearMonth struct {
	Year  int
	Month int
}

func (m *BillRecordManager) ListDistinctMonths(userID int64) ([]YearMonth, error) {
	var results []YearMonth
	err := m.db.Model(&model.BillRecord{}).
		Select("YEAR(transaction_time) as year, MONTH(transaction_time) as month").
		Where("user_id=? AND is_deleted = ?", userID, false).
		Group("YEAR(transaction_time), MONTH(transaction_time)").
		Order("year DESC, month DESC").
		Scan(&results).Error
	return results, err
}

func (m *BillRecordManager) ListDistinctMonthsInRange(userID int64, start, end time.Time) ([]YearMonth, error) {
	var results []YearMonth
	err := m.db.Model(&model.BillRecord{}).
		Select("YEAR(transaction_time) as year, MONTH(transaction_time) as month").
		Where("user_id=? AND transaction_time>=? AND transaction_time<? AND is_deleted = ?", userID, start, end, false).
		Group("YEAR(transaction_time), MONTH(transaction_time)").
		Order("year ASC, month ASC").
		Scan(&results).Error
	return results, err
}

// PeriodAgg holds aggregated bill totals for a time period.
type PeriodAgg struct {
	TotalIncome      int64
	TotalExpense     int64
	AlipayIncome     int64
	AlipayExpense    int64
	WechatIncome     int64
	WechatExpense    int64
	InvestmentAmount int64
}

func (m *BillRecordManager) AggregatePeriod(userID int64, start, end time.Time) (*PeriodAgg, error) {
	type row struct {
		Channel      int8
		BillType     int8
		IsInvestment bool
		Total        int64
	}
	var rows []row
	err := m.db.Model(&model.BillRecord{}).
		Select("channel, bill_type, is_investment, SUM(amount) as total").
		Where("user_id=? AND transaction_time>=? AND transaction_time<? AND is_deleted = ?", userID, start, end, false).
		Group("channel, bill_type, is_investment").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	agg := &PeriodAgg{}
	for _, r := range rows {
		applyPeriodAggRow(agg, r.Channel, r.BillType, r.IsInvestment, r.Total)
	}
	return agg, nil
}

func applyPeriodAggRow(agg *PeriodAgg, channel, billType int8, isInvestment bool, total int64) {
	if isInvestment {
		agg.InvestmentAmount += total
		return
	}
	switch {
	case billType == int8(consts.BillTypeIncome) && channel == int8(consts.ChannelAlipay):
		agg.AlipayIncome += total
	case billType == int8(consts.BillTypeIncome) && channel == int8(consts.ChannelWechat):
		agg.WechatIncome += total
	case billType == int8(consts.BillTypeExpense) && channel == int8(consts.ChannelAlipay):
		agg.AlipayExpense += total
	case billType == int8(consts.BillTypeExpense) && channel == int8(consts.ChannelWechat):
		agg.WechatExpense += total
	}
	agg.TotalIncome = agg.AlipayIncome + agg.WechatIncome
	agg.TotalExpense = agg.AlipayExpense + agg.WechatExpense
}
