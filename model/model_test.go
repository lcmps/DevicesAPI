package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lcmps/DevicesAPI/model"
	"github.com/lcmps/DevicesAPI/model/database"
)

func TestDevice_TranslateToAPI(t *testing.T) {
	created := time.Date(2023, 10, 5, 14, 48, 0, 0, time.UTC)
	dbDevice := database.Device{
		ID:        uuid.MustParse("3fa85f64-5717-4562-b3fc-2c963f66afa6"),
		Name:      "Device1",
		Brand:     "BrandA",
		State:     "Available",
		CreatedAt: created,
	}
	var dvc model.Device
	dvc.TranslateToAPI(dbDevice)
	if dvc.ID != dbDevice.ID.String() {
		t.Fatalf("expected ID %v, got %v", dbDevice.ID.String(), dvc.ID)
	}
	if dvc.Name != dbDevice.Name {
		t.Fatalf("expected Name %v, got %v", dbDevice.Name, dvc.Name)
	}
	if dvc.Brand != dbDevice.Brand {
		t.Fatalf("expected Brand %v, got %v", dbDevice.Brand, dvc.Brand)
	}
	if dvc.State != dbDevice.State {
		t.Fatalf("expected State %v, got %v", dbDevice.State, dvc.State)
	}
	expectedTime := created.Format("2006-01-02T15:04:05Z07:00")
	if dvc.CreatedAt != expectedTime {
		t.Fatalf("expected CreatedAt %v, got %v", expectedTime, dvc.CreatedAt)
	}
}

func TestDevice_TranslateToDB(t *testing.T) {
	dvc := model.Device{
		ID:        "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		Name:      "Device1",
		Brand:     "BrandA",
		State:     "Available",
		CreatedAt: "2023-10-05T14:48:00Z",
	}
	dbDevice := dvc.TranslateToDB()
	if dbDevice.ID.String() != dvc.ID {
		t.Fatalf("expected ID %v, got %v", dvc.ID, dbDevice.ID.String())
	}
	if dbDevice.Name != dvc.Name {
		t.Fatalf("expected Name %v, got %v", dvc.Name, dbDevice.Name)
	}
	if dbDevice.Brand != dvc.Brand {
		t.Fatalf("expected Brand %v, got %v", dvc.Brand, dbDevice.Brand)
	}
	if dbDevice.State != dvc.State {
		t.Fatalf("expected State %v, got %v", dvc.State, dbDevice.State)
	}
}

func TestDeviceList_TranslateToAPI(t *testing.T) {
	created1 := time.Date(2023, 10, 5, 14, 48, 0, 0, time.UTC)
	created2 := time.Date(2023, 10, 6, 15, 30, 0, 0, time.UTC)
	dbDevices := []database.Device{
		{
			ID:        uuid.MustParse("3fa85f64-5717-4562-b3fc-2c963f66afa6"),
			Name:      "Device1",
			Brand:     "BrandA",
			State:     "Available",
			CreatedAt: created1,
		},
		{
			ID:        uuid.MustParse("4fa85f64-5717-4562-b3fc-2c963f66afa7"),
			Name:      "Device2",
			Brand:     "BrandB",
			State:     "Inactive",
			CreatedAt: created2,
		},
	}

	var dvcList model.DeviceList
	dvcList.TranslateToAPI(dbDevices)

	if dvcList.Total != 2 {
		t.Fatalf("expected Total 2, got %d", dvcList.Total)
	}
	if len(dvcList.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(dvcList.Devices))
	}

	// Check first device
	if dvcList.Devices[0].ID != dbDevices[0].ID.String() {
		t.Fatalf("expected ID %v, got %v", dbDevices[0].ID.String(), dvcList.Devices[0].ID)
	}
	if dvcList.Devices[0].Name != dbDevices[0].Name {
		t.Fatalf("expected Name %v, got %v", dbDevices[0].Name, dvcList.Devices[0].Name)
	}
	if dvcList.Devices[0].Brand != dbDevices[0].Brand {
		t.Fatalf("expected Brand %v, got %v", dbDevices[0].Brand, dvcList.Devices[0].Brand)
	}
	if dvcList.Devices[0].State != dbDevices[0].State {
		t.Fatalf("expected State %v, got %v", dbDevices[0].State, dvcList.Devices[0].State)
	}
	expectedTime1 := created1.Format("2006-01-02T15:04:05Z07:00")
	if dvcList.Devices[0].CreatedAt != expectedTime1 {
		t.Fatalf("expected CreatedAt %v, got %v", expectedTime1, dvcList.Devices[0].CreatedAt)
	}

	// Check second device
	if dvcList.Devices[1].ID != dbDevices[1].ID.String() {
		t.Fatalf("expected ID %v, got %v", dbDevices[1].ID.String(), dvcList.Devices[1].ID)
	}
	if dvcList.Devices[1].Name != dbDevices[1].Name {
		t.Fatalf("expected Name %v, got %v", dbDevices[1].Name, dvcList.Devices[1].Name)
	}
	if dvcList.Devices[1].Brand != dbDevices[1].Brand {
		t.Fatalf("expected Brand %v, got %v", dbDevices[1].Brand, dvcList.Devices[1].Brand)
	}
	if dvcList.Devices[1].State != dbDevices[1].State {
		t.Fatalf("expected State %v, got %v", dbDevices[1].State, dvcList.Devices[1].State)
	}
	expectedTime2 := created2.Format("2006-01-02T15:04:05Z07:00")
	if dvcList.Devices[1].CreatedAt != expectedTime2 {
		t.Fatalf("expected CreatedAt %v, got %v", expectedTime2, dvcList.Devices[1].CreatedAt)
	}
}
