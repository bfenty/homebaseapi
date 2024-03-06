package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type shift []struct {
	ID         int    `json:"id"`
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Payroll_id string `json:"payroll_id"`
	Shift_id   int    `json:"shift_id"`
	Role       string `json:"role"`
	Labor      struct {
		Paid_hours      float64 `json:"paid_hours"`
		Scheduled_hours float64 `json:"scheduled_hours"`
		Wage_rate       float64 `json:"wage_rate"`
	} `json:"labor"`
	Clock_in string `json:"clock_in"`
}

type shiftdetail struct {
	ID         int
	First_name string
	Last_name  string
	Payroll_id string
	Shift_id   int
	Role       string
	Labor      struct {
		Paid_hours      float64
		Scheduled_hours float64
	}
	Clock_in time.Time
}

func shiftinsert(shifts shift) {
	username := os.Getenv("USER")
	password := os.Getenv("PASS")
	server := os.Getenv("SERVER")
	port := os.Getenv("PORT")
	connectstring := username + ":" + password + "@tcp(" + server + ":" + port + ")/orders"

	db, err := sql.Open("mysql", connectstring)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	for _, s := range shifts {
		if s.Payroll_id != "" {
			parsedTime, err := time.Parse(time.RFC3339, s.Clock_in)
			if err != nil {
				log.Printf("Failed to parse time [%s] for shift [%d]: %v", s.Clock_in, s.Shift_id, err)
				continue // Skip this entry or handle the error as appropriate
			}

			// Convert to UTC and format as MySQL datetime string
			clockInUTC := parsedTime.UTC().Format("2006-01-02 15:04:05")

			query := "REPLACE INTO shifts (shift_id, payroll_id, role, clock_in, paid_hours, scheduled_hours, wage_rate) VALUES (?, ?, ?, ?, ?, ?, ?)"
			_, err = db.Exec(query, s.Shift_id, s.Payroll_id, s.Role, clockInUTC, s.Labor.Paid_hours, s.Labor.Scheduled_hours, s.Labor.Wage_rate)
			if err != nil {
				log.Printf("Failed to insert shift [%d]: %v", s.Shift_id, err)
			} else {
				log.Printf("Successfully inserted shift [%d] with clock-in time %s", s.Shift_id, clockInUTC)
			}
		}
	}
}

func main() {
	log.Println("Starting process to insert shifts...")
	location := os.Getenv("LOCATION")
	limit := "100"
	end_date := time.Now()
	start_date := end_date.AddDate(0, 0, -1)
	API_key := os.Getenv("API_KEY")

	url := "https://app.joinhomebase.com/api/public/locations/" + location + "/timecards?page=1&per_page=" + limit + "&start_date=" + start_date.Format("2006-1-2") + "&end_date=" + end_date.Format("2006-1-2") + "&date_filter=clock_in"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+API_key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Debug: Log the raw JSON string
	log.Printf("Raw JSON response: %s", string(body))

	var shifts shift
	err = json.Unmarshal(body, &shifts)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	shiftinsert(shifts)
	log.Println("Finished inserting shifts.")
}
