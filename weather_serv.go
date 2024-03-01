package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"picker_weather_serv/models"
	"time"
)

// RowKv This structure is used to load the NATS key/value store.
// key   = timestamp
// value = JSON data
type RowKv struct {
	key   time.Time
	value []byte
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")

	// Fetch geo-coordinates of store(s)
	latitude := 40.7128
	longitude := -74.0060

	// fetch weather data of the store(s)
	getCurrentWeather(latitude, longitude, apiKey)
	getForecastData(latitude, longitude, apiKey)
}

func getCurrentWeather(latitude float64, longitude float64, apiKey string) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric", latitude, longitude, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var weatherData models.WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Printf("Current weather at coordinates (Lat: %.2f, Lon: %.2f):\n", latitude, longitude)
	fmt.Printf("Temperature: %.1f°C\n", weatherData.Main.Temp)
	fmt.Printf("Description: %s\n", weatherData.Weather[0].Description)
	fmt.Printf("Humidity: %d%%\n", weatherData.Main.Humidity)
	fmt.Printf("Wind Speed: %.1f m/s\n", weatherData.Wind.Speed)
}

func getForecastData(latitude float64, longitude float64, apiKey string) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric", latitude, longitude, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var forecastData models.ForecastData
	err = json.Unmarshal(body, &forecastData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Marshal the struct back into formatted JSON
	//formattedData, err := json.MarshalIndent(forecastData, "", "    ")
	//if err != nil {
	//	fmt.Println("Error marshalling JSON:", err)
	//	return
	//}
	//
	//// Write formatted JSON data to a file
	//if err := ioutil.WriteFile("formatted_forecast_data.json", formattedData, 0644); err != nil {
	//	fmt.Println("Error writing file:", err)
	//	return
	//}

	fmt.Println(forecastData)
	fmt.Println(string(body))
}
