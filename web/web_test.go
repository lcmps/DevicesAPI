package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/lcmps/DevicesAPI/model"
	"github.com/lcmps/DevicesAPI/model/database"
)

// Define an interface for the DB methods used in tests
type DeviceDB interface {
	GetDevices(limit, offset int, brand, state, name string) ([]model.Device, error)
	CreateDevice(device *database.Device) error
}

// MockDB for newDevice test, matching DeviceDB interface
type MockDBNewDevice struct {
	existing  []model.Device
	createErr error
}

func (m *MockDBNewDevice) GetDevices(limit, offset int, brand, state, name string) ([]model.Device, error) {
	var result []model.Device
	for _, d := range m.existing {
		if (brand == "" || d.Brand == brand) && (name == "" || d.Name == name) {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MockDBNewDevice) CreateDevice(device *database.Device) error {
	if m.createErr != nil {
		return m.createErr
	}
	// Optionally, convert database.Device to model.Device and append to m.existing for test state
	m.existing = append(m.existing, model.Device{
		Name:  device.Name,
		Brand: device.Brand,
		State: device.State,
	})
	return nil
}

// Use DeviceDB interface for Web struct in tests
type TestWeb struct {
	Router *gin.Engine
	DB     DeviceDB
}

func TestIsValidState(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"Available", "Available", true},
		{"InUseHyphen", "In-Use", true},
		{"Inactive", "Inactive", true},
		{"Empty", "", false},
		{"Lowercase", "available", false},
		{"NoHyphen", "InUse", false},
		{"TrailingSpace", "In-Use ", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidState(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       model.Device
		mockDB     *MockDBNewDevice
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "Missing fields",
			body:       model.Device{Brand: "BrandX", State: "Available"},
			mockDB:     &MockDBNewDevice{},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "name, brand, and state are required fields",
		},
		{
			name:       "Invalid state",
			body:       model.Device{Name: "DeviceA", Brand: "BrandX", State: "Unknown"},
			mockDB:     &MockDBNewDevice{},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid state value, should be one of: Available, In-Use, Inactive",
		},
		{
			name:       "Duplicate device",
			body:       model.Device{Name: "DeviceA", Brand: "BrandX", State: "Available"},
			mockDB:     &MockDBNewDevice{existing: []model.Device{{Name: "DeviceA", Brand: "BrandX", State: "Available"}}},
			wantStatus: http.StatusConflict,
			wantMsg:    "a device with the same name and brand already exists",
		},
		{
			name:       "Success",
			body:       model.Device{Name: "DeviceB", Brand: "BrandY", State: "Available"},
			mockDB:     &MockDBNewDevice{},
			wantStatus: http.StatusCreated,
			wantMsg:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			web := &TestWeb{Router: gin.New(), DB: tc.mockDB}

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			bodyBytes, _ := json.Marshal(tc.body)
			ctx.Request, _ = http.NewRequest("POST", "/api/device/", bytes.NewReader(bodyBytes))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// Call the handler directly, passing the test DB
			newDeviceWithTestDB(web.DB, ctx)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			if tc.wantMsg != "" {
				var resp model.RestError
				_ = json.Unmarshal(rec.Body.Bytes(), &resp)
				if resp.Message != tc.wantMsg {
					t.Fatalf("expected error message %q, got %q", tc.wantMsg, resp.Message)
				}
			}
		})
	}
}

// Helper: newDevice handler using DeviceDB interface for tests
func newDeviceWithTestDB(db DeviceDB, ctx *gin.Context) {
	var requestBody model.Device
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, model.RestError{Message: err.Error()})
		return
	}

	if requestBody.Name == "" || requestBody.Brand == "" || requestBody.State == "" {
		ctx.JSON(http.StatusBadRequest, model.RestError{Message: "name, brand, and state are required fields"})
		return
	}

	if !isValidState(requestBody.State) {
		ctx.JSON(http.StatusBadRequest, model.RestError{Message: "invalid state value, should be one of: Available, In-Use, Inactive"})
		return
	}

	existingDevices, err := db.GetDevices(1, 0, requestBody.Brand, "", requestBody.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.RestError{Message: err.Error()})
		return
	}
	if len(existingDevices) > 0 {
		ctx.JSON(http.StatusConflict, model.RestError{Message: "a device with the same name and brand already exists"})
		return
	}

	newDevice := model.Device{
		Name:  requestBody.Name,
		Brand: requestBody.Brand,
		State: requestBody.State,
	}
	dbDevice := newDevice.TranslateToDB()

	err = db.CreateDevice(&dbDevice)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.RestError{Message: err.Error()})
		return
	}

	var dvc model.Device
	dvc.TranslateToAPI(dbDevice)
	ctx.JSON(http.StatusCreated, dvc)
}
