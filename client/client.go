package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Result map[string]string

type FileInfo struct {
	Name string
	Size int64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Error creating request :: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending request :: %v", err)
	}
	defer res.Body.Close()

	log.Printf("Sending request :: %s [%s] - %s%s", req.Proto, req.Method, req.Host, req.URL.Path)

	file, err := os.Create("logs/cotacao.txt")
	if err != nil {
		log.Fatalf("Error creating log file :: %v", err)
	}
	defer file.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body :: %v", err)
	}

	var result Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error parsing JSON-encoded data :: %v", err)
	}

	resultString := fmt.Sprintf("Dólar: %s", result["bid"])
	_, err = file.WriteString(resultString)
	if err != nil {
		log.Fatalf("Error writing in log file :: %v", err)
	}

	io.Copy(os.Stdout, res.Body)

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Error getting file info :: %v", err)
	}
	log.Printf("Log file created :: %s - (%d bytes)", fileInfo.Name(), fileInfo.Size())

}
