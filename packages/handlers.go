package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// item struct
type Item struct {
	TaskNum int    `json:"id"`
	Task    string `json:"task"`
	Status  bool   `json:"status"`
}

func OpenConnection() (*sql.DB, string) {
	// connecting to DB (in production)
	// retrieve the url
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("Error loading env variables.")
	}
	// connect to the db
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}

	// add email to users table if not present
	email := GetEmail()
	addEmail := `INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING;`
	_, err = db.Exec(addEmail, email)
	if err != nil {
		panic(err)
	}

	// get user_id
	var userId string
	getUser := `SELECT user_id FROM users WHERE email = $1;`
	err = db.QueryRow(getUser, email).Scan(&userId)
	if err != nil {
		panic(err)
	}

	return db, userId
}

// get and parse the (modified) token for email
func GetEmail() string {
	key, ok := os.LookupEnv("NAMESPACE_DOMAIN")
	if !ok {
		log.Fatal("Error loading env variables (namespace domain)")
	}
	_, token := Middleware()
	email := token[key].(string)
	return email
}

// get complete list
var GetList = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// set header to json content, otherwise data appear as plain text
	w.Header().Set("Content-Type", "application/json")

	db, userId := OpenConnection()

	rows, err := db.Query("SELECT id, task, status FROM tasks JOIN users ON tasks.user_uuid = users.user_id WHERE user_id = $1;", userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

	// Initializing slice like this and not "var items []Item" because aforementioned method returns null when empty, while used method returns empty slice
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.TaskNum, &item.Task, &item.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			panic(err)
		}
		items = append(items, item)
	}

	// output with indentation
	// convert items into byte stream
	itemBytes, _ := json.MarshalIndent(items, "", "\t")
	// write to w
	_, err = w.Write(itemBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)

	// output without indentation
	// NewEncoder: WHERE should the encoder write to
	// Encode: encode WHAT
	// _ = json.NewEncoder(w).Encode(items)
})

// add task
var AddTask = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// decode the requested data to 'newTask'
	var newTask Item

	// NewDecoder: Decode FROM WHERE
	// Decode: WHERE TO STORE the decoded data
	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	db, userId := OpenConnection()
	defer db.Close()

	sqlStatement := `INSERT INTO tasks (task, status, user_uuid) VALUES ($1, $2, $3) RETURNING id, task, status;`

	// retrieve the task after creation from the database and store its details in 'updatedTask' (updatedTask will have the correct id regardless of what was input, and auto-assigned false if no status was given. false status is also given to newTask, but newTask has an id 0 if not specified)
	var updatedTask Item
	err = db.QueryRow(sqlStatement, newTask.Task, newTask.Status, userId).Scan(&updatedTask.TaskNum, &updatedTask.Task, &updatedTask.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)

	// gives the new task as the output
	_ = json.NewEncoder(w).Encode(updatedTask)
})

// delete task
var DeleteTask = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the number from the request url
	vars := mux.Vars(r)
	number, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	db, userId := OpenConnection()
	sqlStatement := `DELETE FROM tasks WHERE id = $1 AND user_uuid = $2;`

	res, err := db.Exec(sqlStatement, number, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	// // verifying if row was deleted
	_, err = res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}
	// fmt.Println(count)

	// to get the remaining tasks, same as the GET function
	rows, err := db.Query("SELECT id, task, status FROM tasks JOIN users ON tasks.user_uuid = users.user_id WHERE user_id = $1;", userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}
	defer rows.Close()
	defer db.Close()

	// var items []Item
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.TaskNum, &item.Task, &item.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			panic(err)
		}
		items = append(items, item)
	}

	// output with indentation
	// convert items into byte stream
	itemBytes, _ := json.MarshalIndent(items, "", "\t")

	w.WriteHeader(http.StatusOK)

	// write to w
	_, err = w.Write(itemBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		panic(err)
	}
})

// edit task
var EditTask = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the number from the request url
	vars := mux.Vars(r)
	number, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	sqlStatement := `UPDATE tasks SET task = $2 WHERE id = $1 AND user_uuid = $3 RETURNING id, task, status;`

	// decode the requested data to 'newTask'
	var newTask Item

	// NewDecoder: Decode FROM WHERE
	// Decode: WHERE TO STORE the decoded data
	err = json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	db, userId := OpenConnection()
	defer db.Close()

	// retrieve the task after creation from the database and store its details in 'updatedTask'
	var updatedTask Item
	err = db.QueryRow(sqlStatement, number, newTask.Task, userId).Scan(&updatedTask.TaskNum, &updatedTask.Task, &updatedTask.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)

	// gives the new task as the output
	_ = json.NewEncoder(w).Encode(updatedTask)
})

// change task status
var DoneTask = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the number from the request url
	vars := mux.Vars(r)
	number, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	// store current status of the task from database
	var currStatus bool

	// store updated task
	var updatedTask Item

	sqlStatement1 := `SELECT status FROM tasks WHERE id = $1 AND user_uuid = $2;`
	sqlStatement2 := `UPDATE tasks SET status = $2 WHERE id = $1 AND user_uuid = $3 RETURNING id, task, status;`

	db, userId := OpenConnection()
	defer db.Close()

	// getting current status of the task
	err = db.QueryRow(sqlStatement1, number, userId).Scan(&currStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	// changing the status of the task
	err = db.QueryRow(sqlStatement2, number, !currStatus, userId).Scan(&updatedTask.TaskNum, &updatedTask.Task, &updatedTask.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}
	w.WriteHeader(http.StatusOK)

	// gives the new task as the output
	_ = json.NewEncoder(w).Encode(updatedTask)
})
