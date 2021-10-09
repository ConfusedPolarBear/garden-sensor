package dbconn

import
(
"gorm.io/gorm"
"gorm.io/driver/sqlite"
)

func Open() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	return db, err
}