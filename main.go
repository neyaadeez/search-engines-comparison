package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// generateMyData()

}

func generateMyData() {
	file, err := os.Open("hw1_set2_queries.txt")
	if err != nil {
		fmt.Println("error while opening a queries file: ", err.Error())
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("error while reading the data from the file: ", err.Error())
		return
	}

	lines := strings.Split(string(data), "\n")
	resMap := make(map[string][]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		time.Sleep(time.Second * 2)
		res, err := getYahooSearchResults(line)
		if err != nil {
			fmt.Println("error while quering search: ", err.Error())
			return
		}

		resMap[line] = res
	}

	err = writeResultsToFile(resMap, "myResult1.json")
	if err != nil {
		fmt.Println("error while writing file: ", err.Error())
		return
	}

	fmt.Println("success")
}

func getYahooSearchResults(query string) ([]string, error) {
	var urls []string

	query = "https://search.yahoo.com/search?p=" + strings.ReplaceAll(query, " ", "+") + "&" + "n=10"

	resp, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, err
	}

	respData, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	respData.Find("h3.title a").Each(func(i int, s *goquery.Selection) {
		if i < 10 {
			url, exists := s.Attr("href")
			if exists {
				urls = append(urls, url)
			}
		}
	})

	return urls, nil
}

func writeResultsToFile(resMap map[string][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(resMap, "", " ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

// func checkAndResolve(fileName string) (map[string]string, error) {
// 	file, err := os.Open(fileName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	data, err := io.ReadAll(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dataMap := make(map[string][]string)
// 	err = json.Unmarshal(data, &dataMap)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defect := 0

// 	for i := range dataMap {
// 		count := 0
// 		for range dataMap[i] {
// 			count = count + 1
// 		}
// 		if count != 10 {
// 			fmt.Println(i)
// 			// time.Sleep(time.Second * 10)
// 			// res, err := getYahooSearchResults(strings.TrimSpace(i))
// 			// if err != nil {
// 			// 	err = writeResultsToFile(dataMap, "res2.json")
// 			// 	if err != nil {
// 			// 		return nil, err
// 			// 	}
// 			// 	return nil, err
// 			// }

// 			// dataMap[i] = res

// 			defect = defect + 1
// 		}
// 	}

// 	// err = writeResultsToFile(dataMap, "res2.json")
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	fmt.Println("defect: ", defect)

// 	return nil, nil
// }
