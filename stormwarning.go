package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ridge/must/v2"
)

type customTime time.Time

func (ct *customTime) UnmarshalJSON(input []byte) error {
	unix, err := strconv.ParseInt(string(input), 10, 64)
	if err != nil {
		return err
	}

	*ct = customTime(time.Unix(unix, 0))
	return nil
}

type prediction struct {
	Date customTime `json:"dt"`
	Wind struct {
		Speed float64 `json:"speed"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
}

func formatWinds(predictions []prediction, loc *time.Location) string {
	var out string
	for _, prediction := range predictions {
		out += fmt.Sprintf("%s: wind %.1f, gusts %.1f m/s\n", time.Time(prediction.Date).Format("2006-01-02 15:04"), prediction.Wind.Speed, prediction.Wind.Gust)
	}
	return out
}

func main() {
	locationStr := flag.String("location", "Europe/Malta", "")
	cityID := flag.String("city-id", "2562305", "")
	from := flag.String("from-email", "winds@dottedmag.net", "")

	flag.Parse()

	if locationStr == nil || cityID == nil || from == nil {
		flag.Usage()
		os.Exit(2)
	}

	apiKey := os.Getenv("OWM_API_KEY")
	to := strings.Split(strings.TrimSpace(os.Getenv("TO")), ",")
	location := must.OK1(time.LoadLocation(*locationStr))

	weather, err := fetchWeather(apiKey, *cityID)
	if err != nil {
		panic(err)
	}

	if strong := strongWindsTomorrow(weather, location); strong != nil {
		for _, to := range to {
			msg := formatMessage(*from, to, strong, location)

			if err := sendMessage(must.OK1(mail.ParseAddress(*from)), []byte(msg)); err != nil {
				panic(err)
			}
		}
	}
}

func sameDate(a, b time.Time, loc *time.Location) bool {
	a = a.In(loc)
	b = b.In(loc)

	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func startOfTomorrow(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
}

func strongWindsTomorrow(predictions []prediction, loc *time.Location) []prediction {
	var out []prediction

	tomorrow := startOfTomorrow(loc)
	for _, prediction := range predictions {
		if !sameDate(tomorrow, time.Time(prediction.Date), loc) {
			continue
		}
		if prediction.Wind.Speed > 10 || prediction.Wind.Gust > 15 {
			out = append(out, prediction)
		}
	}
	return out
}

func fetchWeather(apiKey string, cityID string) ([]prediction, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?id=%s&appid=%s", cityID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		List []prediction `json:"list"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.List, nil
}

func formatMessage(from string, to string, strong []prediction, loc *time.Location) string {
	return "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Strong winds tomorrow, remove awnings!\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		formatWinds(strong, loc)
}

func sendMessage(fromAddr *mail.Address, body []byte) error {
	cmd := exec.Command("/usr/sbin/sendmail", "-t", "-oi",
		"-f", fromAddr.Address, "-F", fromAddr.Name)
	cmd.Stdin = bytes.NewReader(body)

	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}
