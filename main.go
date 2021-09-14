package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/tidwall/gjson"
)

const (
	cobraDownloadBase = "https://cubecobra.com/cube/download/plaintext/"
)

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cubeID, found := request.PathParameters["cube"]
	if found {
		// download cube list
		u, _ := url.Parse(cobraDownloadBase)
		u, _ = u.Parse(cubeID)

		// get the list
		resp, err := http.Get(u.String())
		if err != nil || resp.StatusCode != http.StatusOK {
			return events.APIGatewayProxyResponse{}, fmt.Errorf("error retrieving cube [%s]", cubeID)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		// failure to read
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		// bad list?
		badCube, _ := regexp.Match(`<html`, body)
		if badCube {
			return events.APIGatewayProxyResponse{}, fmt.Errorf("invalid cube [%s]", cubeID)
		}
		cards := strings.Split(string(body), "\r\n")
		// strip last one, empty
		if cards[len(cards)-1] == "" {
			cards = cards[:len(cards)-1]
		}
		// now have cards populated with cube list

		// download oracle
		bucket := "cube-keywords"
		item := "oracle-cards.json"

		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String("us-west-2")},
		)

		file, err := os.Create("/tmp/" + item)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		defer file.Close()

		downloader := s3manager.NewDownloader(sess)

		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(item),
			})
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		oracleBytes, err := ioutil.ReadFile("/tmp/" + item)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		oracle := string(oracleBytes)
		if oracle != "" {
		}

		// find keywords
		// frequency of keyword/layout
		freq := make(map[string]int)

		// for each card...
		for _, v := range cards {
			// do nothing
			if v != "" {
			}
			// get keywords
			keywords := gjson.Get(oracle, fmt.Sprintf(`#(name="%s").keywords`, v))
			if keywords.Exists() {
				// return events.APIGatewayProxyResponse{Body: fmt.Sprintf("%+v", keywords), StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html; charset=UTF-8"}}, nil
				keywords.ForEach(func(key, value gjson.Result) bool {
					freq[strings.ToLower(value.String())]++
					return true // keep iterating
				})
			}
		}

		var returnBody string
		var keys []string
		for k := range freq {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			returnBody += fmt.Sprintf("\"%s\" %d\n", k, freq[k])
		}

		return events.APIGatewayProxyResponse{Body: returnBody, StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html; charset=UTF-8"}}, nil

	} else {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("unable to determine cubeID")
	}
}

func main() {
	lambda.Start(Handler)
}
