package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type OutputCSV struct {
	Query          string
	OverLap        int
	OverLapPercent float32
	Spearman       float32
}

type AllAvgs struct {
	OverLap        float32
	OverLapPercent float32
	Spearman       float32
}

type RankTable struct {
	GoogleRank []uint
	YahooRank  []uint
	Di         []int
	DiSqr      []int
}

func main() {
	// SortTheData()
	// return
	// generateMyData()
	csvdata, avgs, err := compareData("Google_Result2.json", "yahooData.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// fmt.Println(csvdata)

	file, err := os.Create("hw1.csv")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{
		"Queries",
		"Number of Overlapping Results",
		"Percent Overlap",
		"Spearman Coefficient",
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, record := range csvdata {
		err = writer.Write([]string{
			record.Query,
			strconv.Itoa(record.OverLap),
			strconv.FormatFloat(float64(record.OverLapPercent), 'f', 2, 32),
			strconv.FormatFloat(float64(record.Spearman), 'f', 2, 32),
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	err = writer.Write([]string{
		"Averages",
		strconv.FormatFloat(float64(avgs.OverLap), 'f', 2, 64),
		strconv.FormatFloat(float64(avgs.OverLapPercent), 'f', 2, 64),
		strconv.FormatFloat(float64(avgs.Spearman), 'f', 2, 64),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

func compareData(fileName1 string, fileName2 string) ([]OutputCSV, AllAvgs, error) {
	file1, err := os.Open(fileName1)
	if err != nil {
		fmt.Println("error while opening a queries file: ", err.Error())
		return nil, AllAvgs{}, err
	}
	defer file1.Close()

	file2, err := os.Open(fileName2)
	if err != nil {
		fmt.Println("error while opening a queries file: ", err.Error())
		return nil, AllAvgs{}, err
	}
	defer file2.Close()

	var file1Data map[string][]string
	var file2Data map[string][]string

	readBytes1, err := io.ReadAll(file1)
	if err != nil {
		fmt.Println("error while reading the data from the file: ", err.Error())
		return nil, AllAvgs{}, err
	}

	readBytes2, err := io.ReadAll(file2)
	if err != nil {
		fmt.Println("error while reading the data from the file: ", err.Error())
		return nil, AllAvgs{}, err
	}

	err = json.Unmarshal(readBytes1, &file1Data)
	if err != nil {
		return nil, AllAvgs{}, err
	}

	err = json.Unmarshal(readBytes2, &file2Data)
	if err != nil {
		return nil, AllAvgs{}, err
	}

	file3, err := os.Open("100QueriesSet2.txt")
	if err != nil {
		return nil, AllAvgs{}, err
	}

	readKeys, err := io.ReadAll(file3)
	if err != nil {
		return nil, AllAvgs{}, err
	}

	var finalResult []OutputCSV
	keys := strings.Split(string(readKeys), "\n")
	for i, key := range keys {
		key = strings.TrimSpace(key)
		var result OutputCSV
		result.Query = "Query " + strconv.Itoa(i+1)
		if value1Data, ok := file1Data[key]; ok {
			if value2Data, ok := file2Data[key]; ok {
				value1map := make(map[string]int)
				value2map := make(map[string]int)
				for v1index, v1 := range value1Data {
					v1 = strings.TrimSpace(v1)
					value1map[v1] = v1index + 1
				}

				for v2index, v2 := range value2Data {
					v2 = strings.TrimSpace(v2)
					value2map[v2] = v2index + 1
				}

				var rankTable RankTable
				for cmpV1, R1 := range value1map {
					if R2, ok := value2map[cmpV1]; ok {
						rankTable.GoogleRank = append(rankTable.GoogleRank, uint(R1))
						rankTable.YahooRank = append(rankTable.YahooRank, uint(R2))
						rankDif := R1 - R2
						rankTable.Di = append(rankTable.Di, rankDif)
						rankTable.DiSqr = append(rankTable.DiSqr, rankDif*rankDif)
					}
				}

				sumDiSqr := 0
				for _, sumVal := range rankTable.DiSqr {
					sumDiSqr += sumVal
				}

				n := len(rankTable.GoogleRank)
				scc := float32(0)
				// if n != 0 {
				// 	scc = 1 - ((6) * float32(sumDiSqr) / float32(n*((n*n)-1)))
				// }

				if n > 1 {
					denominator := float32(n * ((n * n) - 1))
					if denominator != 0 {
						scc = 1 - (6 * float32(sumDiSqr) / denominator)
					} else {
						scc = float32(0)
					}
				} else if n == 1 {
					if len(rankTable.GoogleRank) == 1 && len(rankTable.YahooRank) == 1 && rankTable.GoogleRank[0] == rankTable.YahooRank[0] {
						scc = 1
					} else {
						scc = 0
					}
				} else {
					scc = float32(0)
				}
				result.OverLap = n
				result.OverLapPercent = (float32(n) / 10) * 100
				result.Spearman = scc
			} else {
				return nil, AllAvgs{}, fmt.Errorf("error key not found in file2Data")
			}
		} else {
			return nil, AllAvgs{}, fmt.Errorf("error key not found in file1Data")
		}
		finalResult = append(finalResult, result)
	}

	var totalOverLap, totalPercentOverlap, totalSpearman float32
	for _, r := range finalResult {
		totalOverLap += float32(r.OverLap)
		totalPercentOverlap += r.OverLapPercent
		totalSpearman += r.Spearman
	}
	count := float32(len(finalResult))
	avg := AllAvgs{
		OverLap:        totalOverLap / count,
		OverLapPercent: totalPercentOverlap / count,
		Spearman:       totalSpearman / count,
	}

	return finalResult, avg, nil
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

func checkAndResolve(fileName string) (map[string]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	dataMap := make(map[string][]string)
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}

	defect := 0

	for i := range dataMap {
		count := 0
		for range dataMap[i] {
			count = count + 1
		}
		if count != 10 {
			fmt.Println(i)
			// time.Sleep(time.Second * 10)
			// res, err := getYahooSearchResults(strings.TrimSpace(i))
			// if err != nil {
			// 	err = writeResultsToFile(dataMap, "res2.json")
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	return nil, err
			// }

			// dataMap[i] = res

			defect = defect + 1
		}
	}

	// err = writeResultsToFile(dataMap, "res2.json")
	// if err != nil {
	// 	return nil, err
	// }
	fmt.Println("defect: ", defect)

	return nil, nil
}

func SortTheData() {
	// Step 1: Read the text file containing the keys
	textFilePath := "hw1_set2_queries.txt"
	textFileData, err := os.ReadFile(textFilePath)
	if err != nil {
		log.Fatalf("Failed to read text file: %v", err)
	}
	keys := strings.Split(string(textFileData), "\n")
	for i, key := range keys {
		keys[i] = strings.TrimSpace(key)
	}

	// Step 2: Read and parse the JSON data
	jsonFilePath := "yahooData.json"
	jsonFileData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var data map[string][]string
	err = json.Unmarshal(jsonFileData, &data)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Step 3: Sort the data based on the keys from the text file
	// Use a slice of maps to maintain order
	var sortedData []map[string][]string
	for _, key := range keys {
		if val, exists := data[key]; exists {
			sortedData = append(sortedData, map[string][]string{key: val})
		}
	}

	// Step 4: Marshal the sorted data back into JSON and print/save
	sortedJSON, err := json.MarshalIndent(sortedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal sorted JSON: %v", err)
	}

	// Optional: Save the sorted JSON back to a file
	err = os.WriteFile("sorted_qyery.json", sortedJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write sorted JSON to file: %v", err)
	}
}
