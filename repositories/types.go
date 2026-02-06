package repositories

import (
	"metachan/database"

	"gorm.io/gorm"
)

var DB *gorm.DB = database.DB

type idType interface {
	~int | ~string
}
