package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type openWeatherData struct {
	Name string `json:"name"`

	Sys struct {
		Country string `json:"country"`
	} `json:"sys"`

	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`

	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`

	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

type weatherResponse struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Temperature float64 `json:"temperature_celsius"`
	FeelsLike   float64 `json:"feels_like_celsius"`
	Humidity    int     `json:"humidity_percent"`
	Pressure    int     `json:"pressure_hpa"`
	Condition   string  `json:"condition"`
	Description string  `json:"description"`
	WindSpeed   float64 `json:"wind_speed_mps"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from go!\n"))
}

func query(city string) (weatherResponse, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")

	if apiKey == "" {
		return weatherResponse{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}

	resp, err := http.Get(
		"https://api.openweathermap.org/data/2.5/weather?APPID=" +
			apiKey +
			"&q=" + url.QueryEscape(city) +
			"&units=metric",
	)
	if err != nil {
		return weatherResponse{}, err
	}

	defer resp.Body.Close()

	var data openWeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return weatherResponse{}, err
	}

	return weatherResponse{
		City:        data.Name,
		Country:     data.Sys.Country,
		Temperature: data.Main.Temp,
		FeelsLike:   data.Main.FeelsLike,
		Humidity:    data.Main.Humidity,
		Pressure:    data.Main.Pressure,
		Condition:   data.Weather[0].Main,
		Description: data.Weather[0].Description,
		WindSpeed:   data.Wind.Speed,
	}, nil
}

func main() {
	http.HandleFunc("/", hello)

	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			city := strings.TrimPrefix(r.URL.Path, "/weather/")

			if city == "" {
				http.Error(w, "city is required", http.StatusBadRequest)
				return
			}

			data, err := query(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(data)
		})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}
