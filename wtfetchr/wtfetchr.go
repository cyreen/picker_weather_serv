package wtfetchr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"picker_weather_serv/models"
)

func GetCurrentWeather(latitude float64, longitude float64, apiKey string) {
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

// returns a 3-hourly forecast for 5 days
func GetFiveDaysForecast(latitude float64, longitude float64, apiKey string) ([]models.ForecastResponse, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric", latitude, longitude, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching weather data: %w", err)
	}
	defer resp.Body.Close()

	var forecastData models.ForecastData
	if err := json.NewDecoder(resp.Body).Decode(&forecastData); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	// Extract datetime and temperature from the forecastData
	var forecastResponses []models.ForecastResponse
	for _, item := range forecastData.List {
		forecastResponses = append(forecastResponses, models.ForecastResponse{
			Datetime:    item.DtTxt,
			Temperature: item.Main.Temp,
		})
	}

	return forecastResponses, nil
}
