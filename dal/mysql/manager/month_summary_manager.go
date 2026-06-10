package manager

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	mysqldb "github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

type MonthSummaryManager struct{ db *gorm.DB }

var (
	monthMgr     *MonthSummaryManager
	monthMgrOnce sync.Once
)

func NewMonthSummaryManager() *MonthSummaryManager {
	monthMgrOnce.Do(func() { monthMgr = &MonthSummaryManager{db: mysqldb.GetDB()} })
	return monthMgr
}

func (m *MonthSummaryManager) Upsert(s *model.MonthSummary) error {
	return m.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "year"}, {Name: "month"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"total_income", "total_expense", "alipay_income", "alipay_expense",
			"wechat_income", "wechat_expense", "investment_amount", "update_time",
		}),
	}).Create(s).Error
}

func (m *MonthSummaryManager) Get(userID int64, year, month int) (*model.MonthSummary, error) {
	var s model.MonthSummary
	err := m.db.Where("user_id=? AND year=? AND month=? AND is_deleted = ?", userID, year, month, false).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *MonthSummaryManager) ListPaged(userID int64, offset, limit int) ([]*model.MonthSummary, int64, error) {
	var items []*model.MonthSummary
	var total int64
	q := m.db.Model(&model.MonthSummary{}).Where("user_id=? AND is_deleted = ?", userID, false)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("year DESC, month DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (m *MonthSummaryManager) ListByRange(userID int64, startYear, startMonth, endYear, endMonth int) ([]*model.MonthSummary, error) {
	var items []*model.MonthSummary
	err := m.db.Where(
		"user_id=? AND is_deleted = ? AND (year*12+month) BETWEEN ? AND ?",
		userID, false, startYear*12+startMonth, endYear*12+endMonth,
	).Order("year ASC, month ASC").Find(&items).Error
	return items, err
}

// ListSummarizedMonths returns all (year, month) pairs already summarized for a user.
func (m *MonthSummaryManager) ListSummarizedMonths(userID int64) (map[int]map[int]bool, error) {
	type ym struct {
		Year  int
		Month int
	}
	var rows []ym
	err := m.db.Model(&model.MonthSummary{}).
		Select("year, month").
		Where("user_id=? AND is_deleted = ?", userID, false).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int]map[int]bool)
	for _, r := range rows {
		if result[r.Year] == nil {
			result[r.Year] = make(map[int]bool)
		}
		result[r.Year][r.Month] = true
	}
	return result, nil
}
