package utils

import (
	"os"

	"gorm.io/gorm"
)

type Filter struct {
	Status string `form:"status"` // e.g., ?status=upcoming
	Search string `form:"search"` // e.g., ?search=concert
	Online bool   `form:"online"` // e.g., ?online=true
	Limit  int    `form:"limit,default=10"`
	Offset int    `form:"offset,default=0"`
}

func ByStatus(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == "" {
			return db
		}
		return db.Where("status = ?", status)
	}
}

func BySearch(title string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if title == "" {
			return db
		}
		if os.Getenv("APP_ENV") == "production" {
			return db.Where("title ILIKE ?", "%"+title+"%")
		} else {
			return db.Where("title LIKE ?", "%"+title+"%")
		}
	}
}

func ByOnline(online bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !online {
			return db
		}
		return db.Where("is_online = ?", true)
	}
}

func ByLimit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit == 0 {
			return db
		}
		return db.Limit(limit)
	}
}

func ByOffset(offset int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if offset == 0 {
			return db
		}
		return db.Offset(offset)
	}
}

func ApplyEventListFilters(query *gorm.DB, filter Filter) *gorm.DB {
	return query.Scopes(
		ByStatus(filter.Status),
		BySearch(filter.Search),
		ByOnline(filter.Online),
		ByLimit(filter.Limit),
		ByOffset(filter.Offset),
	)
}

func ApplyUpcomingEventFilters(query *gorm.DB, filter Filter) *gorm.DB {
	return query.Scopes(
		// ByStatus(filter.Status),
		// BySearch(filter.Search),
		// ByOnline(filter.Online),
		ByLimit(filter.Limit),
		ByOffset(filter.Offset),
	)
}
