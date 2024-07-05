package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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

type RowsAffected struct {
	RowsQuantity int64  `json:"rowsQuantity"`
	RowsString   string `json:"rowsString"`
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

func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./data/quotations.db")
	if err != nil {
		log.Printf("Error connecting to database :: %v", err)
		return nil, err
	}
	return db, nil
}

func createExchangeRateTable(db *sql.DB) error {
	ddlStmt := `    
        CREATE TABLE IF NOT EXISTS exchange_rate (
            id TEXT PRIMARY KEY,
            code TEXT,
            codein TEXT,
            name TEXT,
            high TEXT,
            low TEXT,
            varBid TEXT,
            pctChange TEXT,
            bid TEXT,
            ask TEXT,
            timestamp TEXT,
            create_date TEXT
        );
    `
	_, err := db.Exec(ddlStmt)
	if err != nil {
		log.Printf("Error creating exchange-rate table :: %v - %s", err, ddlStmt)
		return err
	}
	return nil
}

func saveExchangeRate(currencyExchangeRate *CurrencyExchangeRate) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db, err := dbConnect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = createExchangeRateTable(db)
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error when starting transaction :: %v", err)
		return nil, err
	}

	dmlStmt, err := tx.Prepare(`INSERT INTO exchange_rate (id, code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		log.Printf("Error when generating prepare statement :: %v - %v", err, dmlStmt)
		return nil, err
	}
	defer dmlStmt.Close()

	uuidValue, err := uuid.NewV6()
	if err != nil {
		log.Printf("Error generating UUID :: %v", err)
		return nil, err
	}

	result, err := dmlStmt.ExecContext(
		ctx,
		uuidValue.String(),
		(*currencyExchangeRate)["USDBRL"].Code,
		(*currencyExchangeRate)["USDBRL"].CodeIn,
		(*currencyExchangeRate)["USDBRL"].Name,
		(*currencyExchangeRate)["USDBRL"].High,
		(*currencyExchangeRate)["USDBRL"].Low,
		(*currencyExchangeRate)["USDBRL"].VarBid,
		(*currencyExchangeRate)["USDBRL"].PctChange,
		(*currencyExchangeRate)["USDBRL"].Bid,
		(*currencyExchangeRate)["USDBRL"].Ask,
		(*currencyExchangeRate)["USDBRL"].Timestamp,
		(*currencyExchangeRate)["USDBRL"].CreateDate,
	)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			log.Printf("Error when rolling back transaction :: %v", err)
			return nil, err
		}

		log.Printf("Error when saving exchange-rate data :: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error when committing transaction :: %v", err)
		return nil, err
	}

	return result, nil
}

func getRowsAffected(result sql.Result) (*RowsAffected, error) {
	rowsAffectedQty, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	var rowsString string
	if rowsAffectedQty > 1 {
		rowsString = "rows"
	} else {
		rowsString = "row"
	}

	var rowsAffected RowsAffected
	rowsAffected.RowsQuantity = rowsAffectedQty
	rowsAffected.RowsString = rowsString

	return &rowsAffected, nil
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
			json.NewEncoder(w).Encode(map[string]string{"error": "Error getting exchange-rate"})
			log.Printf("Error getting exchange rate :: %v", err)
			return
		}

		result, err := saveExchangeRate(ExchangeRate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error saving exchange-rate"})
			log.Printf("Error saving exchange rate :: %v", err)
			return
		}

		rowsAffected, err := getRowsAffected(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error getting rows affected"})
			log.Printf("Error getting rows affected :: %v", err)
			return
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		res := map[string]interface{}{
			"Bid": (*ExchangeRate)["USDBRL"].Bid,
		}
		json.NewEncoder(w).Encode(res)
		log.Printf("Request processed, (%d) %s affected :: %s - [%s] - %s", rowsAffected.RowsQuantity, rowsAffected.RowsString, r.Proto, r.URL.Path, r.RemoteAddr)
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
