package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/natansa/temperatura-cep/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
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
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}

func zipcodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
	ctx, span := tracer.Start(ctx, "serviceb-zipcodeHandler")
	defer span.End()

	time.Sleep(time.Millisecond * 500)

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
}

func main() {
	tp := initTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer = otel.Tracer("serviceb")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Handle("/metrics", promhttp.Handler())
	r.Method("POST", "/", otelhttp.NewHandler(http.HandlerFunc(zipcodeHandler), "zipcodeHandler"))

	fmt.Println("Service B is running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
