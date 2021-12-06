package db

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitializeDatabase() {
	var err error

	logrus.Trace("[db] initializing database connection")

	if err := util.Mkdir("data"); err != nil {
		logrus.Fatalf("[db] unable to create data directory: %s", err)
	}

	// TODO: support other database backends other than sqlite
	db, err = gorm.Open(sqlite.Open("data/garden.db"), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("[db] unable to open database: %s", db)
	}

	logrus.Debug("[db] connected to database")

	if err := db.AutoMigrate(&util.GardenSystem{}, &util.GardenSystemInfo{}, &util.Reading{}, &util.Sensor{}); err != nil {
		panic(err)
	}

	logrus.Debug("[db] migrations completed successfully")
}

func CreateReading(reading util.Reading) error {
	reading.CreatedAt = time.Now()
	if err := db.Create(&reading).Error; err != nil {
		return err
	}
	return nil
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

func GetCoordinator() (util.GardenSystem, error) {
	// Coordinators are systems that report a channel and are not connected through the mesh.
	var id string

	db.
		Raw("SELECT garden_system_id FROM garden_system_infos WHERE is_mesh = false AND channel >= 1 LIMIT 1").
		Scan(&id)

	return GetSystem(id, false)
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

func ArchiveOldReadings() {
	ticker := time.NewTicker(time.Hour * 24 * 7) // Can test this with smaller values like time.Second * 5

	go func() {
		for {
			t := <-ticker.C
			fmt.Println("Tick at", t)
			var readings []util.Reading
			db.Find(&readings)
			file, _ := os.Create(strconv.Itoa(t.Day()) + "-" + t.Month().String() + ".csv")
			writer := csv.NewWriter(file)

			var data = [][]string{{"GardenSystemID", "Temperature", "Humidity", "CreatedAt"}}
			for _, value := range data {
				err := writer.Write(value)
				if err != nil {
					log.Fatal(err)
				}
			}

			for _, reading := range readings {
				var data = [][]string{{
					reading.GardenSystemID,
					fmt.Sprintf("%f", reading.Temperature),
					fmt.Sprintf("%f", reading.Humidity),
					reading.CreatedAt.String(),
				}}
				for _, value := range data {
					err := writer.Write(value)
					if err != nil {
						log.Fatal(err)
					}
				}
				fmt.Println(reading.GardenSystemID, reading.Temperature, reading.Humidity, reading.CreatedAt)
			}

			writer.Flush()
			file.Close()
			db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&util.Reading{}) // This deletes all the readings
		}
	}()
}

func PopulateTestData() {
	for i := 0; i < 10; i++ {
		t := rand.Float32() * 100
		h := rand.Float32() * 100

		testReading := util.Reading{
			GardenSystemID: "Test",
			Error:          false,
			Temperature:    t,
			Humidity:       h,
		}

		if err := CreateReading(testReading); err != nil {
			panic(err)
		}

		reading := &util.Reading{
			Temperature: t,
		}

		if err := db.Where(reading).First(reading).Error; err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", reading)
	}

}
