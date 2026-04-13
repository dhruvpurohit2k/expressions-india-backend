package utils

import (
	"gorm.io/gorm"
)

type Filter struct {
	Status    string `form:"status"`    // upcoming | completed | cancelled
	Search    string `form:"search"`    // title search
	Online    string `form:"online"`    // "true" | "false" | "" (empty = no filter)
	Paid      string `form:"paid"`      // "true" | "false" | "" (empty = no filter)
	SortOrder string `form:"sortOrder"` // asc | desc (by start_date), default desc
	Limit     int    `form:"limit,default=10"`
	Offset    int    `form:"offset,default=0"`
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
		return db.Where("LOWER(title) LIKE LOWER(?)", "%"+title+"%")
	}
}

func ByOnline(online string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if online == "" {
			return db
		}
		return db.Where("is_online = ?", online == "true")
	}
}

func ByPaid(paid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if paid == "" {
			return db
		}
		return db.Where("is_paid = ?", paid == "true")
	}
}

func ByDateSort(sortOrder string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sortOrder == "asc" {
			return db.Order("start_date ASC")
		}
		return db.Order("start_date DESC")
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
		return db.Offset(offset)
	}
}

// ApplyEventListBaseFilters applies only the WHERE/ORDER filters — no LIMIT/OFFSET.
// Use this for COUNT queries so the total reflects the full filtered set.
func ApplyEventListBaseFilters(query *gorm.DB, filter Filter) *gorm.DB {
	return query.Scopes(
		ByStatus(filter.Status),
		BySearch(filter.Search),
		ByOnline(filter.Online),
		ByPaid(filter.Paid),
		ByDateSort(filter.SortOrder),
	)
}

func ApplyEventListFilters(query *gorm.DB, filter Filter) *gorm.DB {
	return query.Scopes(
		ByStatus(filter.Status),
		BySearch(filter.Search),
		ByOnline(filter.Online),
		ByPaid(filter.Paid),
		ByDateSort(filter.SortOrder),
		ByLimit(filter.Limit),
		ByOffset(filter.Offset),
	)
}

type PodcastFilter struct {
	Search    string `form:"search"`
	SortOrder string `form:"sortOrder"` // asc | desc, default desc
	Limit     int    `form:"limit,default=15"`
	Offset    int    `form:"offset,default=0"`
}

type ArticleFilter struct {
	Search    string `form:"search"`
	Category  string `form:"category"`
	SortOrder string `form:"sortOrder"` // asc | desc, default desc
	Limit     int    `form:"limit,default=15"`
	Offset    int    `form:"offset,default=0"`
}

type CourseFilter struct {
	Search    string `form:"search"`
	Audiences string `form:"audiences"` // comma-separated audience names
	SortField string `form:"sortField"` // createdAt | updatedAt, default createdAt
	SortOrder string `form:"sortOrder"` // asc | desc, default desc
	Limit     int    `form:"limit,default=15"`
	Offset    int    `form:"offset,default=0"`
}

type EnquiryFilter struct {
	Name   string `form:"name"`
	Email  string `form:"email"`
	Phone  string `form:"phone"`
	Limit  int    `form:"limit,default=15"`
	Offset int    `form:"offset,default=0"`
}

type JournalFilter struct {
	Search string `form:"search"`
	Year   int    `form:"year"`
	Limit  int    `form:"limit,default=15"`
	Offset int    `form:"offset,default=0"`
}

type EnrollmentFilter struct {
	Search string `form:"search"` // matches name, email, or phone
	Limit  int    `form:"limit,default=15"`
	Offset int    `form:"offset,default=0"`
}

func ApplyUpcomingEventFilters(query *gorm.DB, filter Filter) *gorm.DB {
	return query.Scopes(
		ByLimit(filter.Limit),
		ByOffset(filter.Offset),
	)
}
