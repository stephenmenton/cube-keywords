package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tidwall/gjson"
)

const (
	cobraDownloadBase = "https://cubecobra.com/cube/download/plaintext/"
	scryfallNamedURL  = "https://api.scryfall.com/cards/named"
)

func parseCube(cubeID string) (map[string]int, error) {
	cards, err := getList(cubeID)
	if err != nil {
		return map[string]int{}, err
	}

	freq := make(map[string]int)

	// for each card
	for _, v := range cards {
		if v != "" {
			// get details
			time.Sleep(20 * time.Millisecond) // don't kill scryfall
			k, l, _ := getCardDetails(v)      // keywords, layout
			if l != "normal" {
				freq[strings.ToLower(l)]++
			}
			for _, kv := range k {
				freq[strings.ToLower(kv)]++
			}
		}
	}

	return freq, nil
}

func getList(cubeID string) ([]string, error) {
	u, _ := url.Parse(cobraDownloadBase)
	u, _ = u.Parse(cubeID)

	// get the list
	resp, err := http.Get(u.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving cube [%s]", cubeID)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// bad list?
	badCube, _ := regexp.Match(`<html`, body)
	if badCube == true {
		return nil, fmt.Errorf("invalid cube [%s]", cubeID)
	}
	cards := strings.Split(string(body), "\n")
	// strip last one, empty
	if cards[len(cards)-1] == "" {
		cards = cards[:len(cards)-1]
	}

	return cards, nil
}

// returns slice of keywords, layout, err
func getCardDetails(name string) ([]string, string, error) {
	u, _ := url.Parse(scryfallNamedURL)
	uv := url.Values{}
	uv.Add("exact", name)
	u.RawQuery = uv.Encode()

	// get the card
	resp, err := http.Get(u.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("error retrieving card [%s]", name)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	gk := gjson.GetBytes(body, "keywords")
	var keywords []string
	for _, v := range gk.Array() {
		keywords = append(keywords, v.Str)
	}
	gl := gjson.GetBytes(body, "layout")
	return keywords, gl.Str, nil
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	pc, found := request.PathParameters["cube"]
	if found {
		// path parameters are typically URL encoded so to get the value
		cubeID, err := url.QueryUnescape(pc)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		freq, err := parseCube(cubeID)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		var body string
		var keys []string
		for k := range freq {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			body += fmt.Sprintf("\"%s\" %d\n", k, freq[k])
		}
		return events.APIGatewayProxyResponse{Body: body, StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html; charset=UTF-8"}}, nil
	} else {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("unable to determine cubeID")
	}
}

func main() {
	lambda.Start(Handler)
}
