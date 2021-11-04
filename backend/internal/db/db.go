package db

import (
	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	// Ideally, this would be done with one call to Delete() and it would delete all dependent data.
	// However, that doesn't work since the data for sensors and system info is left dangling in the database.
	err := db.Transaction(func(tx *gorm.DB) error {
		// Delete the old system info and abort on error.
		if err := tx.Delete(&util.GardenSystemInfo{}, "garden_system_id = ?", system.Identifier).Error; err != nil {
			return err
		}

		// Delete the old sensors and abort on error.
		if err := tx.Delete(&util.Sensor{}, "garden_system_info_id = ?", system.Identifier).Error; err != nil {
			return err
		}

		// Delete the old system and abort on error.
		if err := tx.Delete(&system).Error; err != nil {
			return err
		}

		// Create the new system.
		return tx.
			Create(&system).
			Error
	})

	return err
}

func GetSystem(id string, preloadReadings bool) (util.GardenSystem, error) {
	var system util.GardenSystem

	base := db.
		Preload("Announcement").
		Preload("Announcement.Sensors")

	// Cap the number of preloaded readings to preserve performance
	if preloadReadings {
		base.Preload("Readings", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1440)
		})
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

func DeleteSystem(id string) error {
	// TODO: switch to using gorm's deletion methods instead calls to exec
	err := db.
		Exec(`DELETE FROM readings WHERE garden_system_id = ?`, id).
		Exec(`DELETE FROM sensors WHERE garden_system_info_id = ?`, id).
		Exec(`DELETE FROM garden_system_infos WHERE garden_system_id = ?`, id).
		Exec(`DELETE FROM garden_systems WHERE identifier = ?`, id).Error

	return err
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
