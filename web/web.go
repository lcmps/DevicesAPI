package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lcmps/DevicesAPI/db"
	"github.com/lcmps/DevicesAPI/model"
)

type Web struct {
	Router *gin.Engine
	DB     *db.DB
}

func New(connection *db.DB) *Web {
	gin.SetMode(gin.ReleaseMode)

	return &Web{
		Router: gin.Default(),
		DB:     connection,
	}
}

func isValidState(state string) bool {
	switch state {
	case "Available", "In-Use", "Inactive":
		return true
	default:
		return false
	}
}

func (w *Web) Serve() {
	api := w.Router.Group("/api/device")
	{
		// Create a new device
		api.POST("/", w.newDevice)

		// Fully and/or partially update an existing device.
		api.PUT("/:id", w.updateDevice)

		// Fetch a single device (by ID).
		api.GET("/:id", w.getDeviceByID)

		// fetch all devices.
		// devices by name (partial match).
		// devices by brand.
		// devices by state.
		api.GET("/", w.getDeviceByFilter)

		// Delete a single device.
		api.DELETE("/:id", w.deleteDevice)
	}

	fmt.Println("Starting server on port " + os.Getenv("PORT"))
	err := w.Router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (w *Web) newDevice(ctx *gin.Context) {
	var requestBody model.Device
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Name == "" || requestBody.Brand == "" || requestBody.State == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name, brand, and state are required fields"})
		return
	}

	// checking if the provided state is one of the 3 valid values.
	if !isValidState(requestBody.State) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state value, should be one of: Available, In-Use, Inactive"})
		return
	}

	// Checking if an entry with the same name and brand already exists (and is not deleted).
	// This could've been forced through the following unique constraint statement in the database:
	// -----> CONSTRAINT name_brand_unique UNIQUE (name, brand)
	// but since the document didn't specify that, I've implemented it in the application logic,
	// so if my guess that name+brand should be unique is wrong, it can be easily changed.
	existingDevices, err := w.DB.GetDevices(1, 0, requestBody.Brand, "", requestBody.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(existingDevices) > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "a device with the same name and brand already exists"})
		return
	}

	newDevice := model.Device{
		Name:  requestBody.Name,
		Brand: requestBody.Brand,
		State: requestBody.State,
	}
	dbDevice := newDevice.TranslateToDB()

	err = w.DB.CreateDevice(&dbDevice)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch the device just created to get the real ID and CreatedAt
	created, err := w.DB.GetDevices(1, 0, dbDevice.Brand, dbDevice.State, dbDevice.Name)
	if err != nil || len(created) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch created device"})
		return
	}

	var dvc model.Device
	dvc.TranslateToAPI(created[0])

	ctx.JSON(http.StatusCreated, dvc)
}

func (w *Web) updateDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	var requestBody model.Device
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := w.DB.GetDeviceByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Name and brand properties cannot be updated if the device is in use, so I need to check that first.
	// If the device is in use and either the name or brand is being changed, return an error.
	// Otherwise, proceed with the update.
	if device.State == "In-Use" {
		nameChanged := requestBody.Name != "" && requestBody.Name != device.Name
		brandChanged := requestBody.Brand != "" && requestBody.Brand != device.Brand
		if nameChanged || brandChanged {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot update name or brand: device is currently in use"})
			return
		}
	}

	// Check if any field would actually change
	noChange := true
	if requestBody.Name != "" && requestBody.Name != device.Name {
		noChange = false
	}
	if requestBody.Brand != "" && requestBody.Brand != device.Brand {
		noChange = false
	}
	if requestBody.State != "" && requestBody.State != device.State {
		noChange = false
	}

	if noChange {
		var dvc model.Device
		dvc.TranslateToAPI(device)
		ctx.JSON(http.StatusOK, dvc)
		return
	}

	// Updating only the fields that are provided in the request body.
	if requestBody.Name != "" {
		device.Name = requestBody.Name
	}
	if requestBody.Brand != "" {
		device.Brand = requestBody.Brand
	}
	if requestBody.State != "" {
		// checking if the provided state is one of the 3 valid values.
		if !isValidState(requestBody.State) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state value, should be one of: Available, In-Use, Inactive"})
			return
		}
		device.State = requestBody.State
	}

	err = w.DB.UpdateDevice(device)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dvc model.Device
	dvc.TranslateToAPI(device)

	ctx.JSON(http.StatusOK, dvc)
}

func (w *Web) getDeviceByID(ctx *gin.Context) {
	id := ctx.Param("id")

	device, err := w.DB.GetDeviceByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dvc model.Device
	dvc.TranslateToAPI(device)

	ctx.JSON(http.StatusOK, dvc)
}

// getDeviceByFilter
// accepts the following query parameters:
// - limit: number of records to return (default: 50)
// - start: starting index (default: 0)
// - name: filter by device name (optional, partial match)
// - brand: filter by device brand (optional)
// - state: filter by device state (optional), with the following possible values:
//   - Available
//   - In-use
//   - Inactive
func (w *Web) getDeviceByFilter(ctx *gin.Context) {
	// Setting default values for limit and start parameters if none are provided through the URL query.
	// since returning all records could be heavy on the server and network.
	// Default limit is 50 records, default start is 0th record.
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}

	startStr := ctx.DefaultQuery("start", "0")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start must be an integer"})
		return
	}

	brand := ctx.DefaultQuery("brand", "")
	state := ctx.DefaultQuery("state", "")
	name := ctx.DefaultQuery("name", "")

	devices, err := w.DB.GetDevices(limit, start, brand, state, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dvcList model.DeviceList
	dvcList.TranslateToAPI(devices)

	ctx.JSON(http.StatusOK, dvcList)
}

func (w *Web) deleteDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	// Fetch device to check its state
	device, err := w.DB.GetDeviceByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if device.State == "In-Use" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete device: device is currently in use"})
		return
	}

	// Proceed with deletion
	err = w.DB.DeleteDevice(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
