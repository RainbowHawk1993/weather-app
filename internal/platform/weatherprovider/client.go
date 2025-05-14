package weatherprovider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"weather-app/internal/core"
)

const weatherAPIBaseURL = "http://api.weatherapi.com/v1/current.json"

type WeatherAPIResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

var (
	ErrCityNotFound = fmt.Errorf("city not found")
	ErrAPIRequest   = fmt.Errorf("weather API request failed")
)

func (c *Client) FetchWeather(city string) (*core.Weather, error) {
	params := url.Values{}
	params.Add("key", c.apiKey)
	params.Add("q", city)
	params.Add("aqi", "no")

	reqURL := fmt.Sprintf("%s?%s", weatherAPIBaseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPIRequest, err)
	}
	defer resp.Body.Close()

	var apiResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response: %v", ErrAPIRequest, err)
	}

	if apiResp.Error != nil {
		// WeatherAPI.com error codes: https://www.weatherapi.com/docs/
		if apiResp.Error.Code == 1006 {
			return nil, ErrCityNotFound
		}
		return nil, fmt.Errorf("%w: %s (code: %d)", ErrAPIRequest, apiResp.Error.Message, apiResp.Error.Code)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: received status %d", ErrAPIRequest, resp.StatusCode)
	}

	weather := &core.Weather{
		Temperature: apiResp.Current.TempC,
		Humidity:    float64(apiResp.Current.Humidity),
		Description: apiResp.Current.Condition.Text,
	}

	return weather, nil
}
