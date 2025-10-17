package db_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lcmps/DevicesAPI/db"
	"github.com/lcmps/DevicesAPI/model/database"
)

type MockResult struct {
	Error error
}

type MockConnector struct {
	createErr  error
	lastDevice *database.Device
}

func (m *MockConnector) Create(device interface{}) *MockResult {
	if d, ok := device.(*database.Device); ok {
		m.lastDevice = d
	}
	return &MockResult{Error: m.createErr}
}

func TestNew(t *testing.T) {
	keys := []string{"POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}
	backup := map[string]string{}
	for _, k := range keys {
		backup[k] = os.Getenv(k)
	}
	t.Cleanup(func() {
		for k, v := range backup {
			_ = os.Setenv(k, v)
		}
	})

	t.Run("MissingPostgresEnv", func(t *testing.T) {
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
		if _, err := db.New(); err == nil {
			t.Fatal("expected error when postgres env vars are missing, got nil")
		}
	})

	t.Run("UnreachableHost", func(t *testing.T) {
		_ = os.Setenv("POSTGRES_HOST", "invalid-host.local")
		_ = os.Setenv("POSTGRES_USER", "user")
		_ = os.Setenv("POSTGRES_PASSWORD", "pass")
		_ = os.Setenv("POSTGRES_DB", "dbname")
		if _, err := db.New(); err == nil {
			t.Fatal("expected error when cannot connect to postgres, got nil")
		}
	})

	t.Run("SuccessfulConnection", func(t *testing.T) {
		_ = os.Setenv("POSTGRES_HOST", "localhost")
		_ = os.Setenv("POSTGRES_USER", "postgres")
		_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
		_ = os.Setenv("POSTGRES_DB", "device_api")

		dbInstance, err := db.New()
		if err != nil {
			t.Skipf("skipping: could not connect to test database: %v", err)
		}
		if dbInstance == nil || dbInstance.Connector == nil {
			t.Fatal("expected non-nil db and connector")
		}
	})
}

func TestInit_Integration(t *testing.T) {
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}
	err = dbInstance.Init()
	if err != nil {
		t.Fatalf("expected nil error from Init, got %v", err)
	}

	t.Run("InitError", func(t *testing.T) {
		_ = os.Setenv("POSTGRES_HOST", "invalid_host_for_test")
		badDB, err := db.New()
		if err == nil && badDB != nil {
			err = badDB.Init()
			if err == nil {
				t.Fatalf("expected error from Init with invalid host, got nil")
			}
		}
		_ = os.Setenv("POSTGRES_HOST", "localhost")
	})
}

func TestCreateDevice_Integration(t *testing.T) {
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}

	if err := dbInstance.Connector.AutoMigrate(&database.Device{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	device := &database.Device{Name: "TestDevice", Brand: "TestBrand", State: "Available"}
	err = dbInstance.CreateDevice(device)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = dbInstance.CreateDevice(nil)
	if err == nil {
		t.Fatalf("expected error when creating device with nil pointer, got nil")
	}
}

func TestUpdateDevice_Integration(t *testing.T) {

	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}

	if err := dbInstance.Connector.AutoMigrate(&database.Device{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	device := &database.Device{Name: "UpdateMe", Brand: "BrandX", State: "Available"}
	if err := dbInstance.CreateDevice(device); err != nil {
		t.Fatalf("failed to create device for update: %v", err)
	}

	device.State = "Inactive"
	err = dbInstance.UpdateDevice(*device)
	if err != nil {
		t.Fatalf("expected nil error on update, got %v", err)
	}

	nonExistent := *device
	nonExistent.ID, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")
	err = dbInstance.UpdateDevice(nonExistent)
	if err == nil || err.Error() != "no device found with the given ID" {
		t.Fatalf("expected error for non-existent device, got %v", err)
	}

	sqlDB, _ := dbInstance.Connector.DB()
	_ = sqlDB.Close()
	err = dbInstance.UpdateDevice(*device)
	if err == nil || !strings.Contains(err.Error(), "failed to update device") {
		t.Fatalf("expected error for failed update device, got %v", err)
	}
}

func TestGetDeviceByID_Integration(t *testing.T) {

	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}

	if err := dbInstance.Connector.AutoMigrate(&database.Device{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	device := &database.Device{Name: "FetchMe", Brand: "BrandY", State: "Available"}
	if err := dbInstance.CreateDevice(device); err != nil {
		t.Fatalf("failed to create device for fetch: %v", err)
	}

	fetched, err := dbInstance.GetDeviceByID(device.ID.String())
	if err != nil {
		t.Fatalf("expected nil error on fetch, got %v", err)
	}
	if fetched.ID != device.ID {
		t.Fatalf("expected device ID %v, got %v", device.ID, fetched.ID)
	}

	_, err = dbInstance.GetDeviceByID("00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatalf("expected error for non-existent device, got nil")
	}
}

func TestGetDevices_Integration(t *testing.T) {
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}

	if err := dbInstance.Connector.AutoMigrate(&database.Device{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	dev1 := &database.Device{Name: "Alpha", Brand: "BrandA", State: "Available"}
	dev2 := &database.Device{Name: "Beta", Brand: "BrandB", State: "Inactive"}
	dev3 := &database.Device{Name: "Gamma", Brand: "BrandA", State: "Available"}
	for _, d := range []*database.Device{dev1, dev2, dev3} {
		if err := dbInstance.CreateDevice(d); err != nil {
			t.Fatalf("failed to create device: %v", err)
		}
	}

	brandADevices, err := dbInstance.GetDevices(10, 0, "BrandA", "", "")
	if err != nil {
		t.Fatalf("expected nil error for brand filter, got %v", err)
	}
	for _, d := range brandADevices {
		if d.Brand != "BrandA" {
			t.Fatalf("expected BrandA, got %v", d.Brand)
		}
	}

	alphaDevices, err := dbInstance.GetDevices(10, 0, "", "", "Alpha")
	if err != nil {
		t.Fatalf("expected nil error for name filter, got %v", err)
	}
	found := false
	for _, d := range alphaDevices {
		if d.Name == "Alpha" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected to find device with name Alpha")
	}

	_, err = dbInstance.GetDevices(10, 0, "", "INVALID_STATE_FOR_ERROR", "")
	if err == nil || !strings.Contains(err.Error(), "failed to get devices") {
		t.Fatalf("expected error for failed get devices, got %v", err)
	}
}

func TestDeleteDevice_Integration(t *testing.T) {
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_USER", "postgres")
	_ = os.Setenv("POSTGRES_PASSWORD", "postgres")
	_ = os.Setenv("POSTGRES_DB", "device_api")

	dbInstance, err := db.New()
	if err != nil {
		t.Skipf("skipping: could not connect to test database: %v", err)
	}

	if err := dbInstance.Connector.AutoMigrate(&database.Device{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	device := &database.Device{Name: "DeleteMe", Brand: "BrandZ", State: "Available"}
	if err := dbInstance.CreateDevice(device); err != nil {
		t.Fatalf("failed to create device for delete: %v", err)
	}

	err = dbInstance.DeleteDevice(device.ID.String())
	if err != nil {
		t.Fatalf("expected nil error on delete, got %v", err)
	}

	err = dbInstance.DeleteDevice("00000000-0000-0000-0000-000000000000")
	if err == nil || err.Error() != "no device found with the given ID" {
		t.Fatalf("expected error for non-existent device, got %v", err)
	}

	err = dbInstance.DeleteDevice("not-a-uuid")
	if err == nil {
		t.Fatalf("expected error for invalid UUID, got nil")
	}
}
