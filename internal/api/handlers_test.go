package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"weather-app/internal/core"
	"weather-app/internal/platform/weatherprovider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWeatherProvider struct {
	mock.Mock
}

func (m *MockWeatherProvider) FetchWeather(city string) (*core.Weather, error) {
	args := m.Called(city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Weather), args.Error(1)
}

type MockSubscriptionService struct {
	mock.Mock
}

func TestWeatherHandler_GetWeather(t *testing.T) {
	tests := []struct {
		name                string
		cityQueryParam      string
		mockProviderWeather *core.Weather
		mockProviderError   error
		expectedStatusCode  int
		expectedBody        string
	}{
		{
			name:           "success",
			cityQueryParam: "London",
			mockProviderWeather: &core.Weather{
				Temperature: 15.5,
				Humidity:    60.0,
				Description: "Cloudy",
			},
			mockProviderError:  nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"temperature":15.5,"humidity":60,"description":"Cloudy"}`,
		},
		{
			name:               "missing city query parameter",
			cityQueryParam:     "",
			mockProviderError:  nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "city query parameter is required"}`,
		},
		{
			name:                "city not found error from provider",
			cityQueryParam:      "UnknownCity",
			mockProviderWeather: nil,
			mockProviderError:   weatherprovider.ErrCityNotFound,
			expectedStatusCode:  http.StatusNotFound,
			expectedBody:        `{"error": "City not found"}`,
		},
		{
			name:                "api request error from provider",
			cityQueryParam:      "ValidCityButAPIError",
			mockProviderWeather: nil,
			mockProviderError:   weatherprovider.ErrAPIRequest,
			expectedStatusCode:  http.StatusInternalServerError,
			expectedBody:        `{"error": "Failed to fetch weather data from provider"}`,
		},
		{
			name:                "generic error from provider",
			cityQueryParam:      "ValidCityButGenericError",
			mockProviderWeather: nil,
			mockProviderError:   errors.New("some other provider error"),
			expectedStatusCode:  http.StatusInternalServerError,
			expectedBody:        `{"error": "An unexpected error occurred"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockProvider := new(MockWeatherProvider)
			weatherHandler := NewWeatherHandler(mockProvider)

			reqPath := "/weather"
			if tc.cityQueryParam != "" {
				reqPath = fmt.Sprintf("/weather?city=%s", url.QueryEscape(tc.cityQueryParam))
			}
			req := httptest.NewRequest(http.MethodGet, reqPath, nil)
			rr := httptest.NewRecorder()

			if tc.cityQueryParam != "" {
				mockProvider.On("FetchWeather", tc.cityQueryParam).Return(tc.mockProviderWeather, tc.mockProviderError).Once()
			}

			http.HandlerFunc(weatherHandler.GetWeather).ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code, "status code mismatch")

			responseBody := strings.TrimSpace(rr.Body.String())
			assert.JSONEq(t, tc.expectedBody, responseBody, "response body mismatch")

			mockProvider.AssertExpectations(t)
		})
	}
}
