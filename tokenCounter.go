package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Result struct {
	url        string
	token      string
	tokenCount int
}

func main() {
	const (
		goRoutineLimit = 5
		token          = "Go"
	)

	var (
		inputChanel  = make(chan string)
		urlsChanel   = make(chan string, goRoutineLimit-1)
		resultChanel = make(chan Result)
	)

	go ParseStdin(inputChanel)
	go ValidateUrls(inputChanel, urlsChanel)
	go CountTokenFromUrls(token, goRoutineLimit, urlsChanel, resultChanel)
	PrintResult(resultChanel)

}

func ParseStdin(inputChanel chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputString := scanner.Text()
		inputChanel <- inputString
	}
	close(inputChanel)
}

func ValidateUrls(inputChanel chan string, urlsChanel chan string) {
	for input := range inputChanel {
		url_, err := url.ParseRequestURI(input)
		if err != nil {
			fmt.Println("Строка=", input, ", ошибка=", err)
			continue
		}
		urlsChanel <- url_.String()
	}
	close(urlsChanel)
}

func CountTokenFromUrls(token string, goRoutineLimit int, urlsChanel chan string, resultChanel chan Result) {
	var (
		isDoneChanel = make(chan bool)
		workers      = 0
		limit        = make(chan int, goRoutineLimit)
	)

	for url_ := range urlsChanel {
		limit <- 1
		workers++
		go func(url string) {
			responseBody := sendRequest(url)
			if responseBody == "" {
				resultChanel <- Result{url, token, 0}
			}
			count := strings.Count(responseBody, token)
			resultChanel <- Result{url, token, count}
			<-limit
			isDoneChanel <- true
		}(url_)
	}

	// Дожидаюсь завершения всех заданий
	for done := range isDoneChanel {
		if !done {
			fmt.Println("worker error")
		}
		workers--
		if workers == 0 {
			break
		}
	}
	close(isDoneChanel)
	close(resultChanel)
}

func PrintResult(resultChanel chan Result) {
	sum := 0
	for result := range resultChanel {
		fmt.Println("Count for", result.url, ":", result.tokenCount)
		sum += result.tokenCount
	}
	fmt.Println("Total:", sum)
}

func sendRequest(url string) string {
	response, err := http.Get(url)
	if response != nil {
		if err != nil {
			fmt.Println("Ошибка.", response.Status, "name=", url)
			return ""
		}

		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			return string(body)
		}
	}
	return ""
}
