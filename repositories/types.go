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
	MALID     int
	UpdatedAt time.Time
}
