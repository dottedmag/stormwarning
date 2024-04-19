package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ridge/must/v2"
)

func TestStuff(t *testing.T) {
	contents := must.OK1(os.ReadFile("foo.txt"))
	fmt.Println(string(contents))

	var response struct {
		List []prediction `json:"list"`
	}
	err := json.Unmarshal(contents, &response)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", response)
	malta := must.OK1(time.LoadLocation("Europe/Malta"))
	strong := strongWindsTomorrow(response.List, malta)
	fmt.Println(formatWinds(strong, malta))
}
