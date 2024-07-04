package main

import (
	"context"
	"encoding/json"
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

		duration := time.Since(startTime)
		uptime := time.Since(programStartTime)

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

		res := map[string]interface{}{
			"status":   "ok",
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

func main() {
	http.HandleFunc("/health", handlerHealth)
	log.Println("Server listening on localhost:8080")
	http.ListenAndServe(":8080", nil)
}
