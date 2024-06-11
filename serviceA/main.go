package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
			semconv.ServiceNameKey.String("servicea"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp
}

func main() {
	tp := initTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer = otel.Tracer("servicea")

	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "weatherHandler")
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

		url := "http://serviceb:8081"
		postBody, _ := json.Marshal(map[string]string{
			"cep": requestBody.Zipcode,
		})
		responseBody := bytes.NewBuffer(postBody)

		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
		resp, err := client.Post(url, "application/json", responseBody)
		if err != nil {
			http.Error(w, "Error making request to external service", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "External service returned an error", resp.StatusCode)
			return
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response from external service", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	})

	fmt.Println("Service A is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
