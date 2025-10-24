package stat

import (
	"gorm.io/datatypes"
	"time"
	"url/short/pkg/db"
)

type StatRepository struct {
	*db.DB
}

func NewStatRepository(DB *db.DB) *StatRepository {
	return &StatRepository{
		DB: DB,
	}
}

func (repo StatRepository) AddClick(linkId uint) {
	var stat Stat
	currentDate := datatypes.Date(time.Now())
	repo.DB.Find(&stat, "link_id = ? AND date = ?", linkId, currentDate)
	if stat.ID == 0 {

		repo.DB.Create(&Stat{
			LinkId: linkId,
			Clicks: 1,
			Date:   currentDate,
		})
	} else {
		stat.Clicks += 1
		repo.DB.Save(&stat)
	}

}

func (repo StatRepository) GetStats(by string, from, to time.Time) []GetStatResponse {
	var stats []GetStatResponse
	var selectQuery string

	switch by {
	case GroupByDay:
		selectQuery = "to_char(date, 'YYYY-MM-DD') as period, sum(clicks) as sum"
	case GroupByMonth:
		selectQuery = "to_char(date, 'YYYY-MM') as period, sum(clicks) as sum"
	}

	repo.DB.Table("stats").
		Select(selectQuery).
		Where("date BETWEEN ? AND ?", from, to).
		Group("period").
		Order("period desc").
		Scan(&stats)

	return stats
}
