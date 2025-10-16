package db

import (
	"fmt"
	"os"

	"github.com/lcmps/DevicesAPI/model/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Connector *gorm.DB
}

func New() (*DB, error) {
	connString := fmt.Sprintf(`host=%s user=%s password=%s dbname=%s sslmode=disable`,
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	conn, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &DB{Connector: conn}, nil
}

func (db *DB) Init() error {
	// Since i'm using UUID as ID and trigram index, I need to enable both extensions on postgres
	// Also using DO/BEGIN to create the ENUM type as a compatibility measure in case the type already
	// exists AND the postgres version doesn't support 'IF NOT EXISTS' on types
	setupQueries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
		`CREATE EXTENSION IF NOT EXISTS pg_trgm;`,
		`DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'device_state') THEN
				CREATE TYPE device_state AS ENUM ('Available', 'In-Use', 'Inactive');
			END IF;
		END$$;`,
	}

	// Creating a few indexes to optimize queries
	// [0] Index on state to speedup filtering by state
	// [1] Partial index on state where deleted is false to optimize filtering active devices by state
	// [2] Trigram index on name to optimize searching by partial name matches
	postSetupQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_devices_state ON devices(state);`,
		`CREATE INDEX IF NOT EXISTS idx_devices_active_state ON devices(state) WHERE deleted = FALSE;`,
		`CREATE INDEX IF NOT EXISTS idx_devices_name_trgm ON devices USING gin (name gin_trgm_ops);`,
	}

	for _, query := range setupQueries {
		if err := db.Connector.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to execute setup query: %w", err)
		}
	}

	if err := db.Connector.AutoMigrate(&database.Device{}); err != nil {
		return fmt.Errorf("failed to migrate device: %w", err)
	}

	for _, query := range postSetupQueries {
		if err := db.Connector.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to execute post-setup query: %w", err)
		}
	}

	fmt.Println("Database Migrated")
	return nil
}

func (db *DB) CreateDevice(device *database.Device) error {
	result := db.Connector.Create(device)
	if result.Error != nil {
		return fmt.Errorf("failed to create device: %w", result.Error)
	}
	return nil
}

func (db *DB) UpdateDevice(device database.Device) error {
	// Update the device in the database
	// UPDATE devices SET ... WHERE id = ? AND deleted = FALSE
	result := db.Connector.Model(&database.Device{}).Where("id = ? AND deleted = FALSE", device.ID).Updates(device)
	if result.Error != nil {
		return fmt.Errorf("failed to update device: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no device found with the given ID")
	}

	return nil
}

func (db *DB) GetDeviceByID(id string) (database.Device, error) {
	var device database.Device

	// Select * FROM devices WHERE id = ? AND deleted = FALSE LIMIT 1
	result := db.Connector.Where("id = ? AND deleted = FALSE", id).First(&device)
	if result.Error != nil {
		return device, fmt.Errorf("failed to get device by ID: %w", result.Error)
	}

	return device, nil
}

func (db *DB) GetDevices(limit int, offset int, brand, state, name string) ([]database.Device, error) {
	var deviceList []database.Device

	query := db.Connector.Where("deleted = FALSE")

	if brand != "" {
		query = query.Where("brand = ?", brand)
	}
	if state != "" {
		query = query.Where("state = ?", state)
	}
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	result := query.Limit(limit).Offset(offset).Find(&deviceList)
	if result.Error != nil {
		return deviceList, fmt.Errorf("failed to get devices: %w", result.Error)
	}

	return deviceList, nil
}

func (db *DB) DeleteDevice(id string) error {
	// Using soft delete on a device by setting the deleted flag to TRUE on the database
	// UPDATE devices SET deleted = TRUE WHERE id = ? AND deleted = FALSE
	result := db.Connector.Model(&database.Device{}).Where("id = ? AND deleted = FALSE", id).Update("deleted", true)
	if result.Error != nil {
		return fmt.Errorf("failed to delete device: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no device found with the given ID")
	}

	return nil
}
