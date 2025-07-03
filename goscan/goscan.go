// Name: goscan
// Description: A cross-platform directory scanner written in Golang
// Author: isa-programmer
// Repository: https://github.com/isa-programmer/goscan
// LICENSE: MIT
// Version: v1.1.0
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
	"io"
	"github.com/charmbracelet/lipgloss"
)

type Response struct {
	Url string
	StatusCode int
	BodyLength int
	IsValid bool
	IsFailed bool
}

type Responses struct{
	Is2XX []Response
	Is4XX []Response
}


func resultsBox(results string){
	style := lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				Bold(true).
				Width(50)
	fmt.Println(style.Render(results))
}

func urlsBox(responses []Response,color string){
	var output string 
	var tempString string
	// green := lipgloss.Color("#74fc05")
	// red := lipgloss.Color("#fc0509")

	style := lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(color)).
			Width(110)
	for index,resp := range responses {
		tempString = fmt.Sprintf("%d. %s [CODE:%d](LEN:%d)\n",
						index+1,
						resp.Url,
						resp.StatusCode,
						resp.BodyLength)
		
		output = output + tempString
	}
	fmt.Println(style.Render(output))

}

func printBanner(){
	style := lipgloss.NewStyle().
				Bold(true).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("#66082a")).
				Foreground(lipgloss.Color("#7c000a")).
				Background(lipgloss.Color("#0f0f0f")).
				Width(60)
	banner := `
	 ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
	██╔════╝ ██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
	██║  ███╗██║   ██║███████╗██║     ███████║██╔██╗ ██║
	██║   ██║██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
	╚██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
	 ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝

	  ⚡️ Fast Web Directory Scanner. ⚡️ v1.1.0
	  	Made by https://github.com/isa-programmer
	  `

	fmt.Println(style.Render(banner))
}


func getUrlsFromFile(path string) ([]string, error) {
	var urlList []string // Creating a empty array...
	var line string
	file, err := os.Open(path) // Try to open file...
	if err != nil {
		return urlList, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = string(scanner.Text())
		if len(line) > 0 && line[0] != '#' { // Ignore empty lines and comments...
			urlList = append(urlList, line)
		}
	}
	return urlList, nil
}

func checkUrl(url string) (Response, error) { // removed maxAttemps for now...
	newResponse := Response{
		Url: url,
		StatusCode: 0,
		BodyLength: 0,
		IsValid: false,
		IsFailed: false,
	}
	resp, err := http.Get(url)
	if err != nil {
		newResponse.IsFailed = true
		return newResponse, err
	}

	defer resp.Body.Close()
	newResponse.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err == nil{
		newResponse.BodyLength = len(body)
	}

	if resp.StatusCode != 404 && resp.StatusCode < 500 {
		newResponse.IsValid = true
		return newResponse, nil
	} else {
		newResponse.IsValid = false
		return newResponse, nil
	}
}

func main() {
	var succes int = 0
	var failed int = 0
	var path string
	var domain string
	var warning bool = true
	var validUrls Responses

	if len(os.Args) < 4 {

		if len(os.Args) < 3 {
			printBanner()
			return
		} 

	} else {

		if os.Args[3] == "--no-warning"{
				warning = false
		}
	}

	path = os.Args[1]
	domain = os.Args[2]
	urlList, err := getUrlsFromFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	var wg sync.WaitGroup
	printBanner()
	start := time.Now()
	for _, url := range urlList {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := checkUrl(domain + url)
			if err != nil && warning{
				fmt.Println(err)
			}			
			if resp.IsValid {
				if resp.StatusCode >= 400{
					validUrls.Is4XX = append(validUrls.Is4XX,resp)
				} else {
					validUrls.Is2XX = append(validUrls.Is2XX,resp)
				}
				succes++

			} else {
				failed++
			}

		}(url)
	}

	wg.Wait()
	stop := time.Now()
	duration := stop.Sub(start)
	finalResults := fmt.Sprintf(`
	✅ Succes:%d
	❌ Failed:%d
	⏳ Elapsed Time:%v
	`, succes,failed,duration)
	urlsBox(validUrls.Is2XX,"#74fc05")
	urlsBox(validUrls.Is4XX,"#fc0509")
	resultsBox(finalResults)

}
