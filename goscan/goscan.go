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
	"strings"
	"time"
)

func printBanner(printHelp bool){
	banner := `
	 ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
	██╔════╝ ██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
	██║  ███╗██║   ██║███████╗██║     ███████║██╔██╗ ██║
	██║   ██║██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
	╚██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
	 ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝`

	fmt.Printf("\x1b[38;5;1m %s \x1b[0m \n",banner)
	if printHelp{
		fmt.Println("\t ⚡️ blazing fast directory scanner ⚡️ v1.0.1")
		fmt.Println("\t Made by https://github.com/isa-programmer")
		fmt.Println("\t Usage:")
		fmt.Println("\t\t goscan wordlist/wordlist.txt https://example.com/")
		fmt.Println("\t\t goscan wordlist/wordlist.txt https://example.com/ --no-warning # If you want ignore errors")
	}
}


func getUrlsFromFile(path string) ([]string, error) {
	var url_list []string
	var line string
	file, err := os.Open(path)
	if err != nil {
		return url_list, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = string(scanner.Text())
		if len(line) > 0 && line[0] != '#' {
			url_list = append(url_list, line)
		}
	}
	return url_list, nil
}

func isValidUrl(url string, maxAttemps int) (int, error) { 
	var attemps int = 0
	resp, err := http.Get(url)
	if err != nil {
		if attemps < maxAttemps{
			attemps++
			return isValidUrl(url,maxAttemps)
		} else {
		return 0, err
		}
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
	var color string
	var warning bool = true
	if len(os.Args) < 4 {
		if len(os.Args) < 3 {
			printBanner(true)
			return
		} 
	} else {

			if os.Args[3] == "--no-warning"{
				warning = false
			}
		}

	path = os.Args[1]
	domain = os.Args[2]
	url_list, err := getUrlsFromFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	var wg sync.WaitGroup
	printBanner(false)
	start := time.Now()
	for _, url := range url_list {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			statusCode, err = isValidUrl(domain + url,3)
			if err != nil && warning{
				fmt.Println(err)
			}			
			if statusCode != 0 {
				space := strings.Repeat(" ",35-len(url))
				if statusCode >= 400{
					color = "\x1b[38;5;1m"
				} else {
					color = "\x1b[38;5;2m"
				}
				fmt.Printf("%s[+]\x1b[0m %s -> %s [%d] \n",color, url,space, statusCode)
				succes++

			} else {
				failed++
			}

		}(url)
	}

	wg.Wait()
	stop := time.Now()
	duration := stop.Sub(start)
	fmt.Println("\x1b[38;5;2mSucces:\x1b[0m", succes)
	fmt.Println("\x1b[38;5;1mFailed:\x1b[0m", failed)
	fmt.Printf("Elapsed time: %v\n",duration)

}
