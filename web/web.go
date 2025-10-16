package web

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func Serve() {
	r := gin.Default()

	api := r.Group("/api/device")
	{
		// Create a new device
		api.POST("/", newDevice)

		// Fully and/or partially update an existing device.
		api.PUT("/:id", updateDevice)

		// Fetch a single device (by ID).
		api.GET("/:id", getDeviceByID)

		// fetch all devices.
		// devices by brand.
		// devices by state.
		api.GET("/", getDeviceByFilter)

		// Delete a single device.
		api.DELETE("/:id", deleteDevice)
	}

	err := r.Run(":" + os.Getenv("PORT"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func newDevice(ctx *gin.Context) {
}

func updateDevice(ctx *gin.Context) {
}

func getDeviceByID(ctx *gin.Context) {
}

func getDeviceByFilter(ctx *gin.Context) {
}

func deleteDevice(ctx *gin.Context) {
}
