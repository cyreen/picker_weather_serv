package wtfetchr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"picker_weather_serv/models"
	"time"
)

func GetCurrentWeather(latitude float64, longitude float64, apiKey string) (models.ForecastResponse, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric", latitude, longitude, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return models.ForecastResponse{}, fmt.Errorf("error fetching weather data: %w", err)
	}
	defer resp.Body.Close()

	var weatherData models.WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return models.ForecastResponse{}, fmt.Errorf("error decoding JSON: %w", err)
	}

	//Extract datetime and temperature from the forecastData
	var forecastResponse models.ForecastResponse
	forecastResponse = models.ForecastResponse{
		Datetime:    time.Now().Format("2006-01-02 15:04:05"),
		Temperature: weatherData.Main.Temp,
	}

	return forecastResponse, nil
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
