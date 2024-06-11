package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

		if len(requestBody.Zipcode) != 8 {
			http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		url := "http://serviceb:8081"
		postBody, _ := json.Marshal(map[string]string{
			"cep": requestBody.Zipcode,
		})
		responseBody := bytes.NewBuffer(postBody)

		resp, err := http.Post(url, "application/json", responseBody)
		if err != nil {
			http.Error(w, "Error making request to external service", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "External service returned an error", resp.StatusCode)
			return
		}

		// Leitura da resposta do Serviço B
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response from external service", http.StatusInternalServerError)
			return
		}

		// Escreve a resposta do Serviço B para o usuário que chamou o Serviço A
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	})

	fmt.Println("Service A is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
