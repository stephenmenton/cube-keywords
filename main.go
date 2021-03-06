package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const (
	cobraDownloadBase = "https://cubecobra.com/cube/download/plaintext/"
	scryfallNamedURL  = "https://api.scryfall.com/cards/named"
)

// this is a quick and dirty POC utility
// this is not representative of my coding style or ability

func main() {
	// parse
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid syntax\n")
		os.Exit(2)
	}
	cubeID := os.Args[1]
	cards, err := getList(cubeID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	}

	freq := make(map[string]int)

	// for each card
	for _, v := range cards {
		if v != "" {
			// get details
			time.Sleep(100 * time.Millisecond) // don't kill scryfall
			k, l, _ := getCardDetails(v)       // keywords, layout
			if l != "normal" {
				freq[strings.ToLower(l)]++
			}
			for _, kv := range k {
				freq[strings.ToLower(kv)]++
			}
		}
	}

	// TODO: flexible output formatting, ordering, etc
	var keys []string
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("\"%s\" %d\n", k, freq[k])
	}
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
