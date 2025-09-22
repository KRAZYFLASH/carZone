package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/KRAZYFLASH/carZone/driver"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	carHandler "github.com/KRAZYFLASH/carZone/handler/car"
	engineHandler "github.com/KRAZYFLASH/carZone/handler/engine"
	carService "github.com/KRAZYFLASH/carZone/service/car"
	engineService "github.com/KRAZYFLASH/carZone/service/engine"
	carStore "github.com/KRAZYFLASH/carZone/store/car"
	engineStore "github.com/KRAZYFLASH/carZone/store/engine"
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

	carHandler := carHandler.NewCarHandler(carService)
	engineHandler := engineHandler.NewEngineHandler(engineService)

	router := mux.NewRouter()

	schemaFile := "store/schema.sql"
	if err := executeSchemaFile(db, schemaFile); err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	router.HandleFunc("/cars/{id}", carHandler.GetCarById).Methods("GET")
	router.HandleFunc("/cars", carHandler.GetCarByBrand).Methods("GET")
	router.HandleFunc("/cars", carHandler.CreateCar).Methods("POST")
	router.HandleFunc("/cars/{id}", carHandler.UpdateCar).Methods("PUT")
	router.HandleFunc("/cars/{id}", carHandler.DeleteCar).Methods("DELETE")

	router.HandleFunc("/engine/{id}", engineHandler.GetEngineById).Methods("GET")
	router.HandleFunc("/engine", engineHandler.CreateEngine).Methods("POST")
	router.HandleFunc("/engine/{id}", engineHandler.UpdateEngine).Methods("PUT")
	router.HandleFunc("/engine/{id}", engineHandler.DeleteEngine).Methods("DELETE")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server is running on %s", addr)

	log.Fatal(http.ListenAndServe(addr, router))

}


func executeSchemaFile(db *sql.DB, fileName string) error {
	sqlfile, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	_, err = db.Exec(string(sqlfile))

	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	return nil
}