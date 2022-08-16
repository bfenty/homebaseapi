package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"encoding/json"
	// "strconv"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	// "github.com/golang-module/carbon/v2"
)

//Define the shift range struct
type shift []struct {
	ID								int			`json:"id"`
	First_name				string	`json:"first_name"`
	Last_name					string	`json:"last_name"`
	Payroll_id				string	`json:"payroll_id"`
	Shift_id					int			`json:"shift_id"`
	Role							string	`json:"role"`
	Labor struct {
		Paid_hours			float64	`json:"paid_hours"`
		Scheduled_hours	float64	`json:"scheduled_hours"`
		}												`json:"labor"`
	Clock_in					string	`json:"clock_in"`
}

//Define the individual shift struct
type shiftdetail struct {
	ID								int
	First_name				string
	Last_name					string
	Payroll_id				string
	Shift_id					int
	Role							string
	Labor struct {
		Paid_hours			float64
		Scheduled_hours	float64
	}
	Clock_in					time.Time
}

func shiftinsert(shifts shift) {

username := os.Getenv("USER")
password := os.Getenv("PASS")
server := os.Getenv("SERVER")
port := os.Getenv("PORT")

//open connection to database
	connectstring := username+":"+password+"@tcp("+server+":"+port+")/orders"
	db, err := sql.Open("mysql",
	connectstring)
	if err != nil {
		fmt.Println("Message: ",err.Error())
	}

	//Test Connection
	pingErr := db.Ping()
	if pingErr != nil {
		fmt.Println("Message: ",err.Error())
	}

	for i := range shifts {
		var newquery string = "replace into shifts (shift_id,payroll_id,role,clock_in,paid_hours,scheduled_hours) VALUES (?,?,?,?,?,?)"
		fmt.Println(newquery,shifts[i].Shift_id,shifts[i].Payroll_id,shifts[i].Role,shifts[i].Clock_in,shifts[i].Labor.Paid_hours,shifts[i].Labor.Scheduled_hours)
		if shifts[i].Payroll_id != "" {
			rows, err := db.Query(newquery,shifts[i].Shift_id,shifts[i].Payroll_id,shifts[i].Role,shifts[i].Clock_in,shifts[i].Labor.Paid_hours,shifts[i].Labor.Scheduled_hours)
			if err != nil {
				fmt.Println("Message: ",err.Error())
			}
			err = rows.Err()
			if err != nil {
				fmt.Println("Message: ",err.Error())
			}
		}
	}
}

func main() {
	fmt.Println("Setting URL...")
	location := os.Getenv("location")
	limit := "100"
	end_date := time.Date(time.Now().Year(),time.Now().Month(),time.Now().Day(),0,0,0,0,time.Local)
	start_date := end_date.AddDate(0,0,-1)
	API_key := os.Getenv("API_key")

	fmt.Println("Finding starting order...")
	url := "https://app.joinhomebase.com/api/public/locations/"+location+"/timecards?page=1&per_page="+limit+"&start_date="+start_date.Format("2006-1-2")+"&end_date="+end_date.Format("2006-1-2")+"&date_filter=clock_in"

	fmt.Println("Creating Request...")
	req, _ := http.NewRequest("GET", url, nil)

	fmt.Println("Setting Headers...")
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+API_key)

	fmt.Println("Executing Request...")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println("Creating Shift Structure...")
	shifts:=shift{}

	fmt.Println("Loading JSON...")
	jsonErr := json.Unmarshal(body, &shifts)
	if jsonErr != nil {
		fmt.Println("Error:",jsonErr.Error())
		fmt.Println("Body:",string(body))
	}

	fmt.Println("Inserting Shifts to Database...")
	shiftinsert(shifts)

//Go through the range of shifts
var tempshift shiftdetail
fmt.Println("Unpacking JSON...")
	for i := range shifts {
		var err error
		fmt.Println(shifts[i].Payroll_id)
		tempshift.ID = shifts[i].ID
		tempshift.First_name = shifts[i].First_name
		tempshift.Last_name = shifts[i].Last_name
		tempshift.Payroll_id = shifts[i].Payroll_id
		tempshift.Shift_id = shifts[i].Shift_id
		tempshift.Role = shifts[i].Role
		tempshift.Labor.Paid_hours = shifts[i].Labor.Paid_hours
		tempshift.Labor.Scheduled_hours = shifts[i].Labor.Scheduled_hours
		tempshift.Clock_in,err = time.Parse(time.RFC3339,shifts[i].Clock_in)
		fmt.Println(tempshift.Clock_in.Month(),"-",tempshift.Clock_in.Day(),"-",tempshift.Clock_in.Year())
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
