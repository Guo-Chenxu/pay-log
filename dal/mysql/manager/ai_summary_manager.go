package manager

import (
	"sync"

	"gorm.io/gorm"

	mysqldb "github.com/Guo-Chenxu/pay-log/dal/mysql"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

type AISummaryManager struct{ db *gorm.DB }

var (
	aiMgr     *AISummaryManager
	aiMgrOnce sync.Once
)

func NewAISummaryManager() *AISummaryManager {
	aiMgrOnce.Do(func() { aiMgr = &AISummaryManager{db: mysqldb.GetDB()} })
	return aiMgr
}

func (m *AISummaryManager) SoftDeleteByRange(userID int64, startYear, startMonth, endYear, endMonth int) error {
	return m.db.Model(&model.AISummary{}).
		Where("user_id=? AND start_year=? AND start_month=? AND end_year=? AND end_month=? AND is_deleted = ?",
			userID, startYear, startMonth, endYear, endMonth, false).
		Update("is_deleted", true).Error
}

func (m *AISummaryManager) Create(s *model.AISummary) error {
	return m.db.Create(s).Error
}

func (m *AISummaryManager) GetByRange(userID int64, startYear, startMonth, endYear, endMonth int) (*model.AISummary, error) {
	var s model.AISummary
	err := m.db.Where("user_id=? AND start_year=? AND start_month=? AND end_year=? AND end_month=? AND is_deleted = ?",
		userID, startYear, startMonth, endYear, endMonth, false).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
