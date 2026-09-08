package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

type shortUrl struct {
	Id          int    `json:"id"`
	OriginalUrl string `json:"orginalUrl"`
}
type Url struct {
	Url string `json:"shortUrl"`
}

var mapping = make(map[string]string)

func main() {
	initDB()
	intiRedis()
	http.HandleFunc("/shortenUrl", shortenUrl)
	http.HandleFunc("/getUrl/", getUrl)

	http.ListenAndServe(":8080", nil)
}

func shortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Please use the POST Method")
		return
	}

	var reqbody shortUrl

	w.Header().Set("content-type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		fmt.Fprintf(w, "Not able to decode the request body")
		return
	}

	longUrl := reqbody.OriginalUrl

	short, err := generaterandomcode()
	if err != nil {
		fmt.Fprintf(w, "Not able to Generate the code")
		return
	}
	mapping[short] = longUrl

	_, err = db.Exec("Insert into urls(orginalUrl,shortUrl) values (?,?)", longUrl, short)
	if err != nil {
		fmt.Fprintf(w, "Not able to store into db")
		return
	}
	fmt.Println("Short code:", short)
	fmt.Println("Redirect URL:", mapping[short])
	json.NewEncoder(w).Encode(short)

}

func getUrl(w http.ResponseWriter, r *http.Request) {
	x := r.URL.Path
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Please use the GET Method")
		return
	}

	shortCode := strings.Split(x, "/getUrl/")[1]
	var originalUrl string

	//Cache HIT
	cached, err := rdb.Get(ctx, shortCode).Result()
	if err == nil {
		fmt.Println("CacheHit- Returning from the Redis")
		w.Header().Set("content-type", "application/json")
		http.Redirect(w, r, cached, http.StatusFound)
		return
	}

	//CacheMISS
	fmt.Println("CacheMiss - Querring the sql")
	err = db.QueryRow("Select orginalUrl from urls where shortUrl=?", shortCode).Scan(&originalUrl)
	if err != nil {
		fmt.Fprintf(w, "Not able to fetch from db")
		return
	}
	data, err := json.Marshal(originalUrl)
	if err == nil {
		rdb.Set(ctx, shortCode, string(data), 10*time.Minute)
		fmt.Println("Stored in Redis Cache")
	}
	http.Redirect(w, r, originalUrl, http.StatusFound)

}

func generaterandomcode() (string, error) {
	allowedChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var shortUrl []byte
	for i := 0; i < 6; i++ {
		randomIndex := rand.IntN(62)
		randomChar := allowedChars[randomIndex]
		shortUrl = append(shortUrl, randomChar)
	}
	shortcode := string(shortUrl)
	return shortcode, nil

}
