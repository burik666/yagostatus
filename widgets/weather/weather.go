package weather

import (
	"encoding/json"
	"fmt"
	"github.com/burik666/yagostatus/widgets/blank"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

// WidgetParams are widget parameters.
type WidgetParams struct {
	Interval uint
	Location string
	ApiKey   string `yaml:"api_key"`
	Units    string `yaml:"units" default:"metric"`
	Format   string `yaml:"format"`
}

// Widget implements a clock.
type Widget struct {
	blank.Widget
	params WidgetParams
}

type Coordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type Main struct {
	Temp        float32 `json:"temp"`
	FeelsLike   float32 `json:"feels_like"`
	TempMin     float32 `json:"temp_min"`
	TempMax     float32 `json:"temp_max"`
	Pressure    float32 `json:"pressure"`
	Humidity    float32 `json:"humidity"`
	SeaLevel    float32 `json:"sea_level"`
	GroundLevel float32 `json:"grnd_level"`
}

type Condition struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Wind struct {
	Speed  float32 `json:"speed"`
	Degree float32 `json:"deg"`
	Gust   float32 `json:"gust"`
}

type Clouds struct {
	All float32 `json:"all"`
}

type Hourly struct {
	OneHour    float32 `json:"1h"`
	ThreeHours float32 `json:"3h"`
}

type Rain struct {
	Hourly
}

type Snow struct {
	Hourly
}

type Sys struct {
	Type    int    `json:"type"`
	Id      int `json:"id"`
	Message float32 `json:"message"`
	Country string `json:"country"`
	Sunrise int `json:"sunrise"`
	Sunset  int `json:"sunset"`
}

type Result struct {
	Coord           Coordinates `json:"coord"`
	Weather         []Condition `json:"weather"`
	Base            string      `json:"base"`
	Main            Main        `json:"main"`
	Wind            Wind        `json:"wind"`
	Clouds          Clouds      `json:"clouds"`
	Rain            Rain        `json:"rain"`
	Snow            Snow        `json:"snow"`
	CalculationTime int         `json:"dt"`
	Sys             Sys         `json:"sys"`
	Timezone        int         `json:"timezone"`
	Id              int         `json:"id"`
	Name            string      `json:"name"`
	Cod             int         `json:"cod"`
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("weather", NewWeatherWidget, WidgetParams{
		Interval: 1,
		Location: "London, UK",
	})
}

// NewWeatherWidget returns a new ClockWidget.
func NewWeatherWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	w := &Widget{
		params: params.(WidgetParams),
	}

	return w, nil
}

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = w.getWeather()
	c <- res
}

func (w *Widget) getWeather() string {
	c := http.DefaultClient
	apiUrl, err := url.Parse("https://api.openweathermap.org/data/2.5/weather")
	if err != nil {
		return ""
	}
	query := apiUrl.Query()
	query.Add("q", w.params.Location)
	query.Add("APPID", w.params.ApiKey)
	query.Add("units", w.params.Units)
	apiUrl.RawQuery = query.Encode()
	finalUrl := apiUrl.String()
	req, err := http.NewRequest("GET", finalUrl, nil)
	if err != nil {
		return ""
	}

	res, err := c.Do(req)
	if err != nil {
		return ""
	}

	if res.StatusCode != http.StatusOK {
		return ""
	}

	decoder := json.NewDecoder(res.Body)

	var result Result
	err = decoder.Decode(&result)
	if err != nil {
		return err.Error()
	}
	return formatWater(result, w.params.Format)
}

func formatWater(result Result, format string) string {
	// Weather
	format = strings.Replace(format, "%wm", result.Weather[0].Main, -1)
	format = strings.Replace(format, "%wd", result.Weather[0].Description, -1)
	format = strings.Replace(format, "%we", weatherEmoji(result.Weather[0].Id), -1)

	// Main
	format = strings.Replace(format, "%mt", fmt.Sprintf("%.0f", result.Main.Temp), -1)
	format = strings.Replace(format, "%mf", fmt.Sprintf("%.2f", result.Main.FeelsLike), -1)
	format = strings.Replace(format, "%mp", fmt.Sprintf("%.2f", result.Main.Pressure), -1)
	format = strings.Replace(format, "%ma", fmt.Sprintf("%.2f", result.Main.TempMin), -1)
	format = strings.Replace(format, "%mb", fmt.Sprintf("%.2f", result.Main.TempMax), -1)
	format = strings.Replace(format, "%mc", fmt.Sprintf("%.2f", result.Main.SeaLevel), -1)
	format = strings.Replace(format, "%md", fmt.Sprintf("%.2f", result.Main.GroundLevel), -1)

	// Wind 
	format = strings.Replace(format, "%es", fmt.Sprintf("%.2f", result.Wind.Speed), -1)
	format = strings.Replace(format, "%ed", fmt.Sprintf("%.2f", result.Wind.Degree), -1)
	format = strings.Replace(format, "%eg", fmt.Sprintf("%.2f", result.Wind.Gust), -1)

	// Clouds
	format = strings.Replace(format, "%c", fmt.Sprintf("%.2f", result.Clouds.All), -1)

	// Rain
	format = strings.Replace(format, "%r1", fmt.Sprintf("%.2f", result.Rain.OneHour), -1)
	format = strings.Replace(format, "%r3", fmt.Sprintf("%.2f", result.Rain.ThreeHours), -1)

	// Snow
	format = strings.Replace(format, "%s1", fmt.Sprintf("%.2f", result.Snow.OneHour), -1)
	format = strings.Replace(format, "%s3", fmt.Sprintf("%.2f", result.Snow.ThreeHours), -1)

	// Name
	format = strings.Replace(format, "%ln", fmt.Sprintf("%s", result.Name), -1)

	return format
}

func weatherEmoji(icon int) string {
	// https://openweathermap.org/weather-conditions
	if icon == 800 {
		// Clear
		return "🌞"
	}

	if icon >= 200 && icon < 300 {
		return "⛈️"
	}

	if icon >= 300 && icon < 400 {
		// Drizzle
		return "⛆"
	}

	if icon >= 500 && icon < 600 {
		// Rain
		return "🌧️"
	}

	if icon >= 600 && icon < 700 {
		// Snow
		return "❄️"
	}

	if icon >= 700 && icon < 800 {
		//
		return "❓"
	}

	if icon >= 800 && icon < 900 {
		// Clouds
		return "☁️"
	}

	return "❓"
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	w.loop(c)
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for range ticker.C {
		w.loop(c)
	}

	return nil
}
