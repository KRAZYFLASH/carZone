package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/KRAZYFLASH/carZone/driver"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	carHandler "github.com/KRAZYFLASH/carZone/handler/car"
	engineHandler "github.com/KRAZYFLASH/carZone/handler/engine"
	carService "github.com/KRAZYFLASH/carZone/service/car"
	engineService "github.com/KRAZYFLASH/carZone/service/engine"
	carStore "github.com/KRAZYFLASH/carZone/store/car"
	engineStore "github.com/KRAZYFLASH/carZone/store/engine"

	loginHandler "github.com/KRAZYFLASH/carZone/handler/login"
	middleware "github.com/KRAZYFLASH/carZone/middleware"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not found; continuing")
	}

	// --- Tracing init (sekali saja) ---
	ctx := context.Background()
	tp, err := startTracing(ctx)
	if err != nil {
		log.Fatalf("tracing init: %v", err)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer func() { _ = tp.Shutdown(ctx) }()
	// -----------------------------------

	driver.InitDB()
	defer driver.CloseDB()

	db := driver.GetDB()

	cs := carStore.New(db)
	csvc := carService.NewCarService(cs)
	es := engineStore.New(db)
	esvc := engineService.NewEngineService(es)

	ch := carHandler.NewCarHandler(csvc)
	eh := engineHandler.NewEngineHandler(esvc)

	router := mux.NewRouter()

	router.Use(otelmux.Middleware("CarZone"))
	router.Use(middleware.MetricsMiddleware)



	if err := executeSchemaFile(db, "store/schema.sql"); err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	router.HandleFunc("/login", loginHandler.LoginHandler).Methods("POST")

	protected := router.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/cars/{id}", ch.GetCarById).Methods("GET")
	protected.HandleFunc("/cars", ch.GetCarByBrand).Methods("GET")
	protected.HandleFunc("/cars", ch.CreateCar).Methods("POST")
	protected.HandleFunc("/cars/{id}", ch.UpdateCar).Methods("PUT")
	protected.HandleFunc("/cars/{id}", ch.DeleteCar).Methods("DELETE")

	protected.HandleFunc("/engine/{id}", eh.GetEngineById).Methods("GET")
	protected.HandleFunc("/engine", eh.CreateEngine).Methods("POST")
	protected.HandleFunc("/engine/{id}", eh.UpdateEngine).Methods("PUT")
	protected.HandleFunc("/engine/{id}", eh.DeleteEngine).Methods("DELETE")

	router.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" { port = "8000" }

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server is running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func executeSchemaFile(db *sql.DB, fileName string) error {
	sqlfile, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}
	if _, err = db.Exec(string(sqlfile)); err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}
	return nil
}

func startTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(map[string]string{"Content-Type": "application/json"}),
	)
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exp,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("CarZone"),
		)),
	)
	return tp, nil
}
