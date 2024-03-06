package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyreen/CAP-store/grafana_config_updater/logger"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	db "picker_weather_serv/dbconn"
	"picker_weather_serv/models"
	wt "picker_weather_serv/wtfetchr"
	"slices"
	"strconv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("API_KEY")

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	pg, err := db.ConnectToPostgreSQL(dbHost, dbUser, dbPassword, dbName, dbPort)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	stores, err := db.FetchStores(pg)
	if err != nil {
		log.Fatal("Failed to fetch stores:", err)
	}

	var kvs []models.RowKv
	// do weather forecasting for each store
	for _, s := range stores {
		// convert geo-coordinates into floats
		latitude, _ := strconv.ParseFloat(s.Latitude, 32)
		longitude, _ := strconv.ParseFloat(s.Longitude, 32)

		// get weather forecasts as json
		currWeather, err := wt.GetCurrentWeather(latitude, longitude, apiKey)
		forecast, err := wt.GetFiveDaysForecast(latitude, longitude, apiKey)
		if err != nil {
			log.Fatal("Failed to fetch weather forecast: ", err)
		}
		forecast = append(forecast, currWeather)

		// Marshal the forecastData slice into JSON with proper indentation
		formattedData, err := json.MarshalIndent(forecast, "", "    ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		// Write formatted JSON data to a file
		//if err := os.WriteFile("formatted_forecast_data.json", formattedData, 0644); err != nil {
		//	fmt.Println("Error writing file:", err)
		//	return
		//}

		// store result in KV store in NATS
		// K: id_store, V: json of weather data
		kvs = append(kvs, models.RowKv{
			Key:   int(s.ID),
			Value: formattedData,
		})
	}
	err = updateKvStore(kvs)
	if err != nil {
		log.Fatal("Failed to update KV Store: ", err)
	}
}

func updateKvStore(kvSlice []models.RowKv) error {
	logger := logger.GetInstance()

	// JetStream is the built-in NATS persistence system. nats.go provides a built-in API
	// enabling both managing JetStream assets as well as publishing/consuming persistent
	// messages.

	// Jetstream context support was added in Go 1.7
	// In the `jetstream` package, almost all API calls rely on `context.Context` for timeout/cancellation handling
	//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel()

	nc, err := nats.Connect(`nats://nats.cyreenapps.de:4222`, nats.UserCredentials("msuser.creds"))
	if err != nil {
		logger.Printf("ERROR 1: %v\n", err)
		return err
	}
	defer nc.Close()

	// Create a JetStream management interface
	js, err := nc.JetStream()
	if err != nil {
		logger.Printf("ERROR 2: %v\n", err)
		return err
	}

	var kv nats.KeyValue
	// Try to open an existing KV store
	kv, err = js.KeyValue("weather")
	if err != nil {
		// Unable to open an existing store. Create the KV store.
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "weather", History: 1})
		if err != nil {
			logger.Printf("ERROR 3: %v\n", err)
			return err
		}
	}

	var keysToRemove []string

	// Get all the keys in the store
	keyLister, err := kv.ListKeys()

	// Iterate through the keys
	for mykey := range keyLister.Keys() {
		logger.Printf("Key: %s\n", mykey)

		// Remove any key that's not an integer
		iVal, err := strconv.Atoi(mykey)
		if err != nil {
			// val is not an integer. Remove it.
			keysToRemove = append(keysToRemove, mykey)
			continue
		}

		// Find a matching key in our slice
		idx := slices.IndexFunc(kvSlice, func(c models.RowKv) bool { return c.Key == iVal })
		if idx == -1 {
			// the key was not found in our slice. Mark it for removal from the KV store.
			keysToRemove = append(keysToRemove, mykey)
			continue
		}

		// Update the kv store if the value is different
		kvEntry, err := kv.Get(mykey)
		if err != nil {
			logger.Printf("ERROR retrieving key value of key: %s Error: %v\n", mykey, err)
			return err
		}
		if string(kvEntry.Value()) != string(kvSlice[idx].Value) {
			_, err := kv.Update(mykey, kvSlice[idx].Value, kvEntry.Revision())
			if err != nil {
				logger.Printf("ERROR updating kv store key %s Error: %v\n", mykey, err)
				return err
			}
		}
	}

	// Remove any keys marked for removal
	for _, mykey := range keysToRemove {
		err = kv.Purge(mykey)
		if err != nil {
			logger.Printf("ERROR unable to delete key %s from KV store: %v", mykey, err)
		}
	}

	// NOTE: Jetstream count does not reduce after deletion. Therefore, we must check each key
	// Add any new keys to the kv store
	for _, thisElement := range kvSlice {
		// Does it exist in the kvStore?
		_, err := kv.Get(strconv.Itoa(thisElement.Key))
		if err != nil {
			// expected error: 'nats: key not found'
			// The key does not exist.  Add it.
			_, err = kv.Put(strconv.Itoa(thisElement.Key), thisElement.Value)
			if err != nil {
				logger.Printf("ERROR unable to add key %d %v\n", thisElement.Key, err)
			}
		}
	}

	return nil
}
