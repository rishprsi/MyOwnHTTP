package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Couldn't read the required file")
		os.Exit(1)
	}
	strChan := getLinesChannel(file)
	for str := range strChan {
		fmt.Printf("read: %s\n", str)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		fmt.Println("Entering the read loop")
		currStr := ""
		defer f.Close()
		for {
			data := make([]byte, 8)
			_, err := f.Read(data)
			if err != nil {
				if currStr != "" {
					lines <- currStr
				}
				if errors.Is(err, io.EOF) {
					close(lines)
					break
				} else {
					fmt.Printf("Error: %s\n", err)
				}
			}
			dataStr := string(data)
			if strings.Contains(dataStr, "\n") {
				splitStrings := strings.Split(dataStr, "\n")
				length := len(splitStrings)
				for i := 0; i < length-1; i++ {
					currStr += splitStrings[i]
					lines <- currStr
					currStr = ""
				}
				currStr += splitStrings[length-1]
			} else {
				currStr += dataStr
			}
		}
	}()
	return lines
}
