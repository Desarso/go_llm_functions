package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	ell "github.com/desarso/go_llm_functions/helpers"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	ell.API_KEY = os.Getenv("OPEN_ROUTER_API")
	ell.MODEL = "deepseek/deepseek-r1-distill-qwen-32b"
}

func main() {
	assistant := ell.LLM(
		func(prompt string) string {
			return prompt
		},
		"You are a helpful assistant that can get weather information. For coordinates, use San Francisco (37.7749, -122.4194) as an example.",
		ell.Options{
			Debug: true,
		},
		getWeather,
		sayHi,
	)

	fmt.Println(assistant("Say hi to the current user"))
}

var sayHi = ell.CreateTool(
	"sayHi",
	"A function to say hi to the user, only call this function if you know their real first name such as `John` it must be a real name, if you don't know it ask.",
	func(name string) string {
		return fmt.Sprintf("Hello there %s", name)
	},
)

var getWeather = ell.CreateTool(
	"getWeather",
	"A function to get the weather for a given location using latitude and longitude",
	func(lat, lon float64) string {
		url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true", lat, lon)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Sprintf("Error getting weather: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Sprintf("Error: received non-200 response code %d", resp.StatusCode)
		}

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return fmt.Sprintf("Error decoding weather response: %v", err)
		}

		current, ok := result["current_weather"].(map[string]interface{})
		if !ok {
			return "Error: unexpected weather response format"
		}

		temperatureC := current["temperature"].(float64)
		temperatureF := (temperatureC * 9 / 5) + 32
		weathercode := int(current["weathercode"].(float64))

		// Convert weathercode to description
		weather := getWeatherDescription(weathercode)

		return fmt.Sprintf("The weather at coordinates (%f, %f) is %s with temperature %.1f°F", lat, lon, weather, temperatureF)
	},
)

// Helper function to convert weather codes to descriptions
func getWeatherDescription(code int) string {
	codes := map[int]string{
		0:  "Clear sky",
		1:  "Mainly clear",
		2:  "Partly cloudy",
		3:  "Overcast",
		45: "Foggy",
		48: "Depositing rime fog",
		51: "Light drizzle",
		53: "Moderate drizzle",
		55: "Dense drizzle",
		61: "Slight rain",
		63: "Moderate rain",
		65: "Heavy rain",
		71: "Slight snow fall",
		73: "Moderate snow fall",
		75: "Heavy snow fall",
		77: "Snow grains",
		80: "Slight rain showers",
		81: "Moderate rain showers",
		82: "Violent rain showers",
		85: "Slight snow showers",
		86: "Heavy snow showers",
		95: "Thunderstorm",
		96: "Thunderstorm with slight hail",
		99: "Thunderstorm with heavy hail",
	}

	if desc, ok := codes[code]; ok {
		return desc
	}
	return "Unknown weather condition"
}
