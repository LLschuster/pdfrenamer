package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var keywords = []string{"Subject:", "Contract:"}

func reportTime(start time.Time) {
	elapsed := time.Since(start)
	fmt.Println("Execution time: ", elapsed)
}

func main() {
	start := time.Now()
	defer reportTime(start)

	fmt.Println("Start pdf renamer")
	fileNames := make([]string, 0, 130)

	dirEntry, err := os.ReadDir("pdfs")
	if err != nil {
		fmt.Printf("Error reading directory %v/n", err)
		return
	}

	for _, entry := range dirEntry {
		fileNames = append(fileNames, entry.Name())
	}

	syncGroup := sync.WaitGroup{}
	syncGroup.Add(len(fileNames))

	for _, name := range fileNames {
		go func(name string) {
			filePath := fmt.Sprintf("pdfs/%s", name)
			exec.Command("pdftotext", filePath).Run()
			syncGroup.Done()
		}(name)
	}

	syncGroup.Wait()

	syncGroup.Add(len(fileNames))
	for _, name := range fileNames {
		go func(name string) {
			defer syncGroup.Done()
			originalPath := fmt.Sprintf("pdfs/%s", name)
			filePath := fmt.Sprintf("pdfs/%s", strings.Replace(name, ".pdf", ".txt", 1))
			fmt.Println("File path: ", filePath)
			f, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Error opening file %v/n", err)
				return
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)

			scanner.Split(bufio.ScanLines)

			var newNameParts = make([]string, 0, 3)
			for scanner.Scan() {
				line := scanner.Text()
				for _, keyword := range keywords {
					if strings.Contains(line, keyword) {
						if keyword == "Betreff:" {
							newNameParts = append(newNameParts, strings.Replace(line, keyword, "", 1))
						}

						if keyword == "Vertrag:" {
							lineWithoutKeyword := strings.Replace(line, keyword, "", 1)
							addressPartOfLine := strings.Split(lineWithoutKeyword, ":")[1]
							addressParts := strings.Split(addressPartOfLine, ",")
							newNameParts = append(newNameParts, addressParts[0], addressParts[1])
						}
					}
				}
			}

			if scanner.Err() != nil {
				fmt.Printf("Error scanning %v/n", scanner.Err())
				return
			}

			newName := strings.ReplaceAll(strings.Trim(strings.Join(newNameParts, " - "), " "), "/", "|")

			err = os.Rename(originalPath, fmt.Sprintf("%s.pdf", newName))
			if err != nil {
				fmt.Printf("Error renaming file %v/n", err)
				return
			}
		}(name)
	}

	syncGroup.Wait()
}
