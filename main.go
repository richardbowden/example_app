package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"github.com/hashicorp/go-multierror"

	_ "github.com/lib/pq"
	"gopkg.in/matryer/try.v1"
	"github.com/lib/pq"
	"net"
)

var db *sql.DB

type Employee struct {
	ID   int
	Name string
	City string
}

type DBConfiger interface {
	ConnectionString() string
	DBDriver() string
}

type PGDBConfig struct {
	User   string
	Passwd string
	DBName string
	Host   string
	Port   string
}

func (c *PGDBConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Passwd, c.Host, c.Port, c.DBName)
}

func (c PGDBConfig) DBDriver() string {
	return "postgres"
}

func dbConn(config DBConfiger) (*sql.DB, error) {
	conStr := config.ConnectionString()
	db, err := sql.Open(config.DBDriver(), conStr)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

var tmpl = template.Must(template.ParseGlob("form/*.gohtml"))

func Index(w http.ResponseWriter, r *http.Request) {
	selDB, err := db.Query("SELECT * FROM Employee ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	res := []Employee{}
	for selDB.Next() {
		var id int
		var name, city string
		err = selDB.Scan(&id, &name, &city)
		if err != nil {
			panic(err.Error())
		}
		emp.ID = id
		emp.Name = name
		emp.City = city
		res = append(res, emp)
	}

	tmpl.ExecuteTemplate(w, "Index", res)
}

func Show(w http.ResponseWriter, r *http.Request) {
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM Employee WHERE id=$1", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	for selDB.Next() {
		var id int
		var name, city string
		err = selDB.Scan(&id, &name, &city)
		if err != nil {
			panic(err.Error())
		}
		emp.ID = id
		emp.Name = name
		emp.City = city
	}
	tmpl.ExecuteTemplate(w, "Show", emp)
}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM Employee WHERE id=$1", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	for selDB.Next() {
		var id int
		var name, city string
		err = selDB.Scan(&id, &name, &city)
		if err != nil {
			panic(err.Error())
		}
		emp.ID = id
		emp.Name = name
		emp.City = city
	}
	tmpl.ExecuteTemplate(w, "Edit", emp)
}

func Insert(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		name := r.FormValue("name")
		city := r.FormValue("city")
		insForm, err := db.Prepare("INSERT INTO Employee(name, city) VALUES($1,$2)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, city)
		log.Println("INSERT: Name: " + name + " | City: " + city)
	}
	http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		name := r.FormValue("name")
		city := r.FormValue("city")
		id := r.FormValue("uid")
		insForm, err := db.Prepare("UPDATE Employee SET name=$1, city=$2 WHERE id=$3")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, city, id)
		log.Println("UPDATE: Name: " + name + " | City: " + city)
	}
	http.Redirect(w, r, "/", 301)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	emp := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM Employee WHERE id=$1")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(emp)
	log.Println("DELETE")

	http.Redirect(w, r, "/", 301)
}

func getRuntimeConfigFromEnv(config *PGDBConfig) (error) {
	// this is designed to allow settings that have all ready been populated in config to be overridden
	// if set in the environment

	var err *multierror.Error
	baseErr := "missing env var: %s"

	db_host := os.Getenv("EXAMPLE_APP_DB_HOST")
	db_port := os.Getenv("EXAMPLE_APP_DB_PORT")
	db_name := os.Getenv("EXAMPLE_APP_DB_NAME")
	db_user := os.Getenv("EXAMPLE_APP_DB_USER")
	db_passwd := os.Getenv("EXAMPLE_APP_DB_PASSWD")

	if db_host == "" {
		err = multierror.Append(err, fmt.Errorf(baseErr, "EXAMPLE_APP_DB_HOST"))
	} else {
		config.Host = db_host
	}

	if db_port == "" {
		err = multierror.Append(err, fmt.Errorf(baseErr, "EXAMPLE_APP_DB_PORT"))
	} else {
		config.Port = db_port
	}

	if db_name == "" {
		err = multierror.Append(err, fmt.Errorf(baseErr, "EXAMPLE_APP_DB_NAME"))
	} else {
		config.DBName = db_name
	}

	if db_user == "" {
		err = multierror.Append(err, fmt.Errorf(baseErr, "EXAMPLE_APP_DB_USER"))
	} else {
		config.User = db_user
	}

	if db_passwd == "" {
		err = multierror.Append(err, fmt.Errorf(baseErr, " EXAMPLE_APP_DB_PASSWD"))
	} else {
		config.Passwd = db_passwd
	}

	return err.ErrorOrNil()
}

func main() {

	var conConfig PGDBConfig

	err := getRuntimeConfigFromEnv(&conConfig)

	if err != nil {
		fmt.Println("errors detected that prevents the server from starting")
		if merr, ok := err.(*multierror.Error); ok {
			fmt.Println(merr)
		}
		os.Exit(1)
	}

	connectionAttempts := 1000
	try.MaxRetries = connectionAttempts
	err = try.Do(func(attempt int) (bool, error) {
		fmt.Printf("attempting to connect to db %s@%s\n", conConfig.User, conConfig.Host)
		db, err = dbConn(&conConfig)
		if err != nil {
			if cerr, ok := err.(*net.OpError); ok {
				fmt.Printf("error connecting to db: %s, retrying in 5 seconds, try: %v out of %v\n", cerr, attempt, connectionAttempts)
				time.Sleep(5 * time.Second) // wait a minute
			}else if perr, ok := err.(*pq.Error); ok {
				// agg, this is dam annoying, the pq lib does not list all the error codes as
				// consts, so i have to look up the error code, oh look, 3D000 means no database.
				if perr.Code == "3D000"{
					fmt.Printf("database %s does not exist, this needs to be created!\n", conConfig.DBName)
					os.Exit(1)
				}else{
					fmt.Printf("we have an error that we cannot recover from, here is the error: %s, we are now going to quit!\n", err)
					os.Exit(1)
				}
			}else {
				fmt.Printf("we have an error that we cannot recover from, here is the error: %s, we are now going to quit!\n", err)
				os.Exit(1)
			}
		}
		return attempt < connectionAttempts, err
	})

	create_tables()

	if err != nil {
		fmt.Println("we have a problem, we failed to connect to the database, see errors that occured before this message")
		os.Exit(1)
	}
	// create a global db connection, this is thread safe, if uses an internal connection pool so we are resuing connections to the db
	fmt.Println("database connection successful")
	defer db.Close()

	log.Println("Server started on: http://localhost:8081")
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8081", nil)

}

func create_tables() {
	var t = `CREATE SEQUENCE employee_seq;

CREATE TABLE IF NOT EXISTS employee (
  id int check (id > 0) NOT NULL DEFAULT NEXTVAL ('employee_seq'),
  name varchar(30) NOT NULL,
  city varchar(30) NOT NULL,
  PRIMARY KEY (id)
)  ;

ALTER SEQUENCE employee_seq RESTART WITH 1;`

	_, err := db.Exec(t)

	if perr, ok := err.(*pq.Error); ok {
		if perr.Code == "42P07"{
			fmt.Println("table exists, moving on, nothing to see here!")
			return
		}else{
			fmt.Printf("odd error pg when trying to create table - failed to create tables, see source for this error=%s\n", err)
		}

	}

	if err != nil {
		fmt.Printf("odd error when trying to create table - failed to create tables, see source for this error=%s\n", err)
		os.Exit(1)
	}

	fmt.Println("tables created!")
}
