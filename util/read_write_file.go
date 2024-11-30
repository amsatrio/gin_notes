package util

import (
	"encoding/csv"
	"log"
	"os"
)

func WriteToFile(folderName string, fileName string, content string, overwrite bool) {
	// create a dir
	err := os.MkdirAll(folderName, os.ModePerm)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	filePath := folderName + "/" + fileName

	// create or append to file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("create or append to file error")
		log.Fatal(err)
	}

	if overwrite {
		f, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Println("create file error")
			log.Fatal(err)
		}
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	if _, err = f.WriteString(content); err != nil {
		log.Fatal(err)
	}
}

func ReadFromFile(folderName string, fileName string) string {
	filePath := folderName + "/" + fileName
	content, err := os.ReadFile(filePath)

	if err != nil {
		log.Fatal(err)
	}

	return string(content)
}

func ReadCSV(path string, separator int32) [][]string {
	// open file
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comma = separator
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func RemoveAll(folderName string) error {
	err := os.RemoveAll(folderName)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
