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

func CreateGardenSystem(db *gorm.DB, gardenSystem util.GardenSystemInfo) error {
	if err := db.Create(&gardenSystem).Error; err != nil {
		return err
	}
	return nil
}

func GetGardenSystem(db *gorm.DB, systemIdentifier string) (util.GardenSystemInfo, error) {
	systemInfo := util.GardenSystemInfo{
		Identifier: systemIdentifier,
	}

	//result := db.Where(systemInfo)
	//fmt.Println(result)
	if err := db.Where(systemInfo).First(&systemInfo).Error; err != nil {
		panic(err)
	}
	return systemInfo, nil
}
 