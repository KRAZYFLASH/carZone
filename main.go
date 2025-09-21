package main

import (
	"log"

	"github.com/KRAZYFLASH/SimpleApp-CarManagement/driver"
	"github.com/joho/godotenv"

	carStore "github.com/KRAZYFLASH/SimpleApp-CarManagement/carZone/store/car"
	engineStore "github.com/KRAZYFLASH/SimpleApp-CarManagement/carZone/store/engine"
	carService "github.com/KRAZYFLASH/SimpleApp-CarManagement/carZone/service/car"
	engineService "github.com/KRAZYFLASH/SimpleApp-CarManagement/carZone/service/engine"
	"github.com/joho/godotenv"
)

func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	driver.InitDB()

	defer driver.CloseDB()

	db := driver.GetDB()

	carStore := carStore.New(db)
	carService := carService.NewCarService(carStore)
	engineStore := engineStore.New(db)
	engineService := engineService.NewEngineService(engineStore)
}