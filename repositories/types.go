package repositories

import (
	"metachan/database"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB = database.DB

type idType interface {
	~int | ~string
}

type animeStub struct {
	MALID      int
	UpdatedAt  time.Time
	EnrichedAt *time.Time
}

type animeCharacterRow struct {
	Role string
}

type characterStub struct {
	MALID      int
	EnrichedAt *time.Time
}

type personStub struct {
	MALID      int
	EnrichedAt *time.Time
}
