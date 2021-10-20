package dbconn

import
(
"gorm.io/gorm"
"gorm.io/driver/sqlite"
"github.com/ConfusedPolarBear/garden/internal/util"
)

func Open() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	return db, err
}

func CreateReading(db *gorm.DB, reading util.Reading) error {
	if err := db.Create(&reading).Error; err != nil {
		return err
	}
	return nil
}
 