package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/natansa/temperatura-cep/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type RequestBody struct {
	Zipcode string `json:"cep"`
}

var tracer trace.Tracer

func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP trace exporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("serviceb"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}

func main() {
	tp := initTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer = otel.Tracer("serviceb")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "zipcodeHandler")
		defer span.End()

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestBody RequestBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &requestBody); err != nil {
			http.Error(w, "Invalid request body format", http.StatusBadRequest)
			return
		}

		if len(requestBody.Zipcode) != 8 {
			http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		zipcodeHandler := services.NewZipcodeHandler()
		cityName, err := zipcodeHandler.FetchCityNameFromZipcode(requestBody.Zipcode)

		if err != nil {
			http.Error(w, "Cannot find zipcode", http.StatusNotFound)
			return
		}

		weatherService := services.NewWeatherService()
		_, tempSpan := tracer.Start(ctx, "FetchWeather")
		tempCelsius, err := weatherService.FetchWeather(cityName)
		tempSpan.End()

		if err != nil {
			http.Error(w, "Error fetching weather information", http.StatusInternalServerError)
			return
		}

		tempFahrenheit := services.CelsiusToFahrenheit(tempCelsius)
		tempKelvin := services.CelsiusToKelvin(tempCelsius)

		response := map[string]interface{}{
			"city":   cityName,
			"temp_C": tempCelsius,
			"temp_F": tempFahrenheit,
			"temp_K": tempKelvin,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Service B is running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
