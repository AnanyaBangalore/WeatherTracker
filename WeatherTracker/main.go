package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json:"Name"`
	Main struct {
		Kelvin         float64 `json:"temp"`         // Temperature in Kelvin
		Celsius        float64 `json:"temp_celsius"` // Temperature in Celsius
		KelvinResponse float64 `json:"temp_kelvin"`  // Temperature in Kelvin (exported for JSON)
	} `json:"main"`
}

func loadApiConfig(filename string) (apiConfigData, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return apiConfigData{}, err
	}
	var c apiConfigData
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiConfigData{}, err
	}
	return c, nil
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadApiConfig("./apiConfig")
	if err != nil {
		return weatherData{}, err
	}
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city)
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()

	var d weatherData
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return weatherData{}, err
	}

	// Convert temperature from Kelvin to Celsius
	celsius := d.Main.Kelvin - 273.15
	d.Main.Celsius = celsius // Store Celsius temperature in the struct

	// Also store the Kelvin temperature in the exported field 'KelvinResponse'
	d.Main.KelvinResponse = d.Main.Kelvin

	// Print the temperature in Celsius for debugging
	fmt.Printf("The temperature in %s is %.2f°C (%.2fK)\n", d.Name, celsius, d.Main.Kelvin)

	return d, nil
}

func main() {
	fmt.Println("Started the port on 8080")
	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		data, err := query(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(data) // Send the data with temperature in Celsius and Kelvin
	})
	http.ListenAndServe(":8080", nil)
}
