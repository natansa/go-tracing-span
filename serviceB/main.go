package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/natansa/temperatura-cep/services"
)

type RequestBody struct {
	Zipcode string `json:"cep"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		zipcodeHandler := services.NewZipcodeHandler()
		cityName, err := zipcodeHandler.FetchCityNameFromZipcode(requestBody.Zipcode)

		if err != nil {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}

		weatherService := services.NewWeatherService()
		tempCelsius, err := weatherService.FetchWeather(cityName)

		if err != nil {
			http.Error(w, "error fetching weather information", http.StatusInternalServerError)
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

	fmt.Println("Service A is running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
