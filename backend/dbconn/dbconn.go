package dbconn

import
(
"time"
"gorm.io/gorm"
"gorm.io/driver/sqlite"
"github.com/ConfusedPolarBear/garden/internal/util"
)

func Open() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	return db, err
}

func CreateReading(db *gorm.DB, reading util.Reading) error {
	reading.Time = time.Now()
	if err := db.Create(&reading).Error; err != nil {
		return err
	}
	return nil
}

func CreateGardenSystem(db *gorm.DB, gardenSystem util.GardenSystem) error {
	gardenSystem.LastSeen = time.Now()
	if err := db.Create(&gardenSystem).Error; err != nil {
		return err
	}
	return nil
}

func GetGardenSystem(db *gorm.DB, systemIdentifier string) (util.GardenSystem, error) {
	var err error
	system := util.GardenSystem{
		Identifier: systemIdentifier,
	}
	if err = db.Where(system).First(&system).Error; err != nil {
		panic(err)
	}
	system.Announcement, err = GetGardenSystemInfo(db, systemIdentifier)
	if err != nil {
		panic(err)
	}
	return system, nil
}
 

func CreateGardenSystemInfo(db *gorm.DB, gardenSystem util.GardenSystemInfo) error {
	if err := db.Create(&gardenSystem).Error; err != nil {
		return err
	}
	return nil
}

func GetGardenSystemInfo(db *gorm.DB, systemIdentifier string) (util.GardenSystemInfo, error) {
	systemInfo := util.GardenSystemInfo{
		Identifier: systemIdentifier,
	}
	if err := db.Where(systemInfo).First(&systemInfo).Error; err != nil {
		panic(err)
	}
	return systemInfo, nil
}
 