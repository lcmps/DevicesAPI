package db

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type DB struct {
	Connector *gorm.DB
}

type Device struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Deleted   bool      `gorm:"not null;default:false" json:"deleted"`
	Name      string    `gorm:"type:varchar(250);not null" json:"name"`
	Brand     string    `gorm:"type:varchar(250);not null" json:"brand"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	State     string    `gorm:"type:device_state;not null;default:'Available'" json:"state"`
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

	if err := db.Connector.AutoMigrate(&Device{}); err != nil {
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
