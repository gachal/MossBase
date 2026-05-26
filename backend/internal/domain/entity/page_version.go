package entity

import "time"

type PageVersion struct {
	ID            uint64
	PageID        uint64
	VersionNumber int
	Title         string
	Content       string
	ContentHTML   string
	EditedBy      uint64
	CreatedAt     time.Time
}
