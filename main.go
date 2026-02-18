package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type TemperatureResponse struct {
	Monthly MonthlyData `json:"monthly"`
	Trend   TrendData   `json:"trend"`
}

func main() {
	monthly, trend, err := ReadTemperatureFile("data/uahncdc_lt_6.1.txt")
	if err != nil {
		log.Fatalf("load temperature data: %v", err)
	}

	response := TemperatureResponse{
		Monthly: monthly,
		Trend:   trend,
	}
	downloadHandler := NewDownloadHandler("downloads")

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/download", downloadHandler)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
