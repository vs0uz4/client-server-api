package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var programStartTime = time.Now()

type CpuStats struct {
	Cores       int       `json:"cores"`
	UsedPercent []float64 `json:"usedPercent"`
}

type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	Free        uint64  `json:"free"`
	Active      uint64  `json:"active"`
	Inactive    uint64  `json:"inactive"`
}

type ExchangeRate struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type CurrencyExchangeRate map[string]ExchangeRate

func getCPUStats() (*CpuStats, error) {
	usedPercent, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	cores, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	stats := &CpuStats{
		Cores:       cores,
		UsedPercent: usedPercent,
	}

	return stats, nil
}

func getMemoryStats() (*MemoryStats, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	stats := &MemoryStats{
		Total:       vmStat.Total,
		Available:   vmStat.Available,
		Used:        vmStat.Used,
		UsedPercent: vmStat.UsedPercent,
		Free:        vmStat.Free,
		Active:      vmStat.Active,
		Inactive:    vmStat.Inactive,
	}

	return stats, nil
}

func getExchangeRate() (*CurrencyExchangeRate, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var exchangeRate CurrencyExchangeRate
	err = json.Unmarshal(body, &exchangeRate)
	if err != nil {
		return nil, err
	}

	return &exchangeRate, nil
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusRequestTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "Request timeout"})
			log.Printf("Request timeout :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
		} else {
			log.Printf("Request canceled :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
		}
		return
	default:
		w.Header().Set("Accept", "application/health+json")
		w.Header().Set("Content-Type", "application/health+json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.WriteHeader(http.StatusOK)

		memStats, err := getMemoryStats()
		if err != nil {
			log.Printf("Error getting memory stats :: %v", err)
			return
		}

		cpuStats, err := getCPUStats()
		if err != nil {
			log.Printf("Error getting CPU stats :: %v", err)
			return
		}

		duration := time.Since(startTime)
		uptime := time.Since(programStartTime)

		res := map[string]interface{}{
			"status":   "pass",
			"duration": duration.String(),
			"uptime":   uptime.String(),
			"cpu": map[string]interface{}{
				"cores":        cpuStats.Cores,
				"percent_used": cpuStats.UsedPercent,
			},
			"memory": map[string]interface{}{
				"total":        memStats.Total,
				"available":    memStats.Available,
				"used":         memStats.Used,
				"free":         memStats.Free,
				"percent_used": memStats.UsedPercent,
			},
		}

		json.NewEncoder(w).Encode(res)
		log.Printf("Request success :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
	}
}

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusRequestTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "Request timeout"})
			log.Printf("Request timeout :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
		} else {
			log.Printf("Request canceled :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
		}
		return
	default:
		ExchangeRate, err := getExchangeRate()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error getting exchange rate"})
			log.Printf("Error getting exchange rate :: %v", err)
			return
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		res := map[string]interface{}{
			"Bid": (*ExchangeRate)["USDBRL"].Bid,
		}

		json.NewEncoder(w).Encode(res)
		log.Printf("Request success :: %s - [%s] - %s", r.Proto, r.URL.Path, r.RemoteAddr)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlerHealth)
	mux.HandleFunc("/cotacao", handlerQuotation)

	log.Println("Server listening on localhost:8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
