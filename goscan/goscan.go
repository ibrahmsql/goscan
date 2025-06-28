// Name: goscan
// Description: A cross-platform directory scanner written in Golang
// Author: isa-programmer
// Repository: https://github.com/isa-programmer/goscan
// LICENSE: MIT

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
)

func getUrlsFromFile(path string) ([]string, error) {
	var url_list []string
	file, err := os.Open(path)
	if err != nil {
		return url_list, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url_list = append(url_list, scanner.Text())
	}
	return url_list, nil
}

func isValidUrl(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 && resp.StatusCode < 500 {
		return resp.StatusCode, nil
	} else {
		return 0, nil
	}
}

func main() {
	var succes int = 0
	var failed int = 0
	var statusCode int
	var path string
	var domain string
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./gscan wordlist.txt https://example.com")
		return
	}

	path = os.Args[1]
	domain = os.Args[2]
	url_list, err := getUrlsFromFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	var wg sync.WaitGroup

	for _, url := range url_list {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			statusCode, err = isValidUrl(domain + url)
			if err != nil {
				fmt.Println(err)
			}

			if statusCode != 0 {
				fmt.Printf("[+] %s -> [%d] \n", url, statusCode)
				succes++

			} else {
				failed++
			}

		}(url)
	}

	wg.Wait()
	fmt.Println("Succes:", succes)
	fmt.Println("Failed:", failed)

}
