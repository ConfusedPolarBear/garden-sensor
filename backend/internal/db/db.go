package db

import (
	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var db *gorm.DB

func InitializeDatabase() {
	var err error

	logrus.Trace("[db] initializing database connection")

	// TODO: support other database backends other than sqlite
	db, err = gorm.Open(sqlite.Open("garden.db"), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("[db] unable to open database: %s", db)
	}

	logrus.Debug("[db] connected to database")

	if err := db.AutoMigrate(&util.GardenSystem{}, &util.GardenSystemInfo{}, &util.Reading{}, &util.Sensor{}); err != nil {
		panic(err)
	}

	logrus.Debug("[db] migrations completed successfully")
}

func CreateSystem(system util.GardenSystem) error {
	// Upsert the base system and the announcement
	return db.
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&system).
		Error
}

func GetSystem(id string, preloadReadings bool) (util.GardenSystem, error) {
	var system util.GardenSystem

	base := db.
		Preload("Announcement").
		Preload("Announcement.Sensors")

	if preloadReadings {
		base.Preload("Readings")
	}

	err := base.
		Where("identifier = ?", id).
		First(&system).
		Error

	if err == nil {
		loadLatestReading(&system)
	}

	return system, err
}

func GetAllSystems() []util.GardenSystem {
	var systems []util.GardenSystem
	db.
		Preload("Announcement").
		Preload("Announcement.Sensors").
		Find(&systems)

	for i := range systems {
		loadLatestReading(&systems[i])
	}

	return systems
}

func UpdateSystem(system util.GardenSystem) error {
	return db.Save(&system).Error
}

// Loads the latest reading for this system. This is done to avoid preloading the entire slice of Readings as that would
// be inefficient.
func loadLatestReading(system *util.GardenSystem) {
	// Uses Limit() and Find() as opposed to a simple First() because First() will log an error if no readings exist,
	// which happens when a node boots for the first time. It isn't harmful in anyway, it just is bad UX.
	db.
		Order("created_at DESC").
		Where("garden_system_id = ?", system.Identifier).
		Limit(1).
		Find(&system.LastReading)
}
