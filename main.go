package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"log"
	"time"

	
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	_ "github.com/lib/pq"

	"google.golang.org/appengine"
)

type Photo struct {
	PhotoId    int
	PhotoUrl   string   
	AllVotes, UserVotes   int
}

type Vote struct {
	VoteChoice int `json:"vote"`
	PhotoId    int `json:"photo"`
	UserId     int  
}

type User struct {
	UserId, UserSession   int
}


func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


func openDbConnection() *sql.DB {
	dbinfo := os.Getenv("POSTGRES_CONNECTION")
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	return db
}


func createAnonUser(w http.ResponseWriter, r *http.Request) *http.Cookie {
	db := openDbConnection()
	defer db.Close()
	cookie, err := r.Cookie("session-id")
	if err != nil {
		id, _ := uuid.NewV4()
		cookie = &http.Cookie{
			Name:"session-id",
			Value: id.String(),
		}
		sqlStatement := `
		INSERT INTO users(session_id) 
		VALUES($1)`
		_, err = db.Exec(sqlStatement, cookie.Value)
		if err != nil {
			panic(err)
		}
		http.SetCookie(w, cookie)
	}
	return cookie
}


func getUserId(w http.ResponseWriter, r *http.Request) int {
	var currentUser User
	session_id, err := r.Cookie("session-id")
	if err != nil {
		session_id = createAnonUser(w, r)
	}
    db := openDbConnection()
	defer db.Close()
	sqlStatement := `SELECT id from users where session_id = $1`
	err = db.QueryRow(sqlStatement, session_id.Value).Scan(&currentUser.UserId)
	
	return currentUser.UserId
}



// Get votes of all photos and votes for current user
func getAllPhotos(w http.ResponseWriter, r *http.Request) {
	currentUserId := getUserId(w, r)
	db := openDbConnection()
	defer db.Close()
	
	query := `
	select photos.id, photos.url, COUNT(votes), 
	COUNT(votes) filter (where users.id = $1 and votes.choice = 1)
	from photos 
	LEFT JOIN votes on votes.photo_id = photos.id
	LEFT JOIN users on users.id = votes.user_id
	GROUP BY photos.id, photos.url
	ORDER BY count(votes) desc
	`
	rows, err := db.Query(query, currentUserId)
	checkErr(err)
	var photos []Photo
	
	for rows.Next() {
		var photo Photo
		err := rows.Scan(&photo.PhotoId, &photo.PhotoUrl, &photo.AllVotes, &photo.UserVotes)
		photos = append(photos, photo)
		checkErr(err)
	}
	json, err := json.Marshal(photos)
	checkErr(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}


// Get photos user has not voted on
func getPhotoStack(w http.ResponseWriter, r *http.Request) {
	currentUserId := getUserId(w, r)
	db := openDbConnection()
	defer db.Close()
	query := `
	SELECT photos.id, photos.url
	FROM photos
	LEFT JOIN votes ON votes.photo_id = photos.id AND votes.user_id = $1
	WHERE votes.user_id is null
	`
	rows, err := db.Query(query, currentUserId)
	checkErr(err)
	var photos []Photo
	
	for rows.Next() {
		var photo Photo
		err := rows.Scan(&photo.PhotoId, &photo.PhotoUrl)
		photos = append(photos, photo)
		checkErr(err)
	}
	json, err := json.Marshal(photos)
	checkErr(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}


func submitVote(w http.ResponseWriter, r *http.Request) {
	currentUserId := getUserId(w, r)
	var vote Vote

	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		db := openDbConnection()
		defer db.Close()
		err := decoder.Decode(&vote)
		checkErr(err)

		sqlStatement := `
		INSERT INTO votes(choice, photo_id, user_id) 
		VALUES($1, $2, $3)`
		_, err = db.Exec(sqlStatement, 
					  vote.VoteChoice,
					  vote.PhotoId, 
				      currentUserId)
		checkErr(err)
	}
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/home.html")
	t.Execute(w, nil)
}


func voteHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/vote.html")
	t.Execute(w, nil)
}


func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/vote", voteHandler)
	r.HandleFunc("/photos", getAllPhotos)
	r.HandleFunc("/photos/stack", getPhotoStack)
	r.HandleFunc("/photos/vote", submitVote)
	r.PathPrefix("/public").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	log.Fatal(srv.ListenAndServe())
	appengine.Main()
}