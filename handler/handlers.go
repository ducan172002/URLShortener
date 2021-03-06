package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"url-shortener/base62"
	"url-shortener/shorten"
	"url-shortener/store"

	"github.com/gorilla/mux"
)

var prefixLink string = "http://localhost:8080/"

type UrlCreationRequest struct {
	LongUrl string `json:"long_url"`
}

//Home Page
func Home(w http.ResponseWriter, _ *http.Request) {
	sendResponse(w, http.StatusOK, map[string]string{"message": "Welcome to URL shortener"})
}

func Set(id uint64, long string, short string, clicks uint, create time.Time, update time.Time) shorten.URLEntry {
	return shorten.URLEntry{
		Id:          id,
		OriginalURL: long,
		ShortenURL:  short,
		Clicks:      clicks,
		CreateAt:    create,
		UpdateAt:    update,
	}
}

//Create short link
func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var myurl UrlCreationRequest
	var urlshortener shorten.URLEntry

	err := json.NewDecoder(r.Body).Decode(&myurl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidURL(myurl.LongUrl) {
		respondWithError(w, http.StatusBadRequest, "Invalid url")
		return
	}

	if store.CheckURLinDB(myurl.LongUrl) {
		urlshortener = store.GetURLEntry(myurl.LongUrl)
		sendResponse(w, http.StatusOK, map[string]string{"response": urlshortener.ShortenURL, "status": "already in the database"})
	} else {
		shorurl := shorten.GenerateShortLink()
		//returns the current local time.
		timeNow := time.Now().UTC()

		urlshortener = Set(base62.Decode(shorurl), myurl.LongUrl, prefixLink+shorurl, 0, timeNow, timeNow)
		store.SaveURL(urlshortener)
		sendResponse(w, http.StatusOK, map[string]string{"response": urlshortener.ShortenURL, "status": "succescful"})
	}

}
func LinkCounter(entry shorten.URLEntry) {
	entry.Clicks++
	err := store.UpdateCounterLink(entry)
	if !err {
		fmt.Println("Update failed")
	}
	fmt.Println("Update successful")
}

// Redirect link
func HandleShortUrlRedirect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortPath := params["urlshorten"]
	fmt.Println(shortPath)

	var urlCreationRequest UrlCreationRequest

	urlCreationRequest.LongUrl = store.GetLongURL(shortPath)
	if urlCreationRequest.LongUrl == "" {
		respondWithError(w, http.StatusNotFound, "Not found")
		return
	}
	urlEntry := store.GetURLEntry(urlCreationRequest.LongUrl)
	LinkCounter(urlEntry)
	http.Redirect(w, r, urlCreationRequest.LongUrl, http.StatusSeeOther)
}

// Get info url entry
func GetURLEntry(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortPath := params["urlshorten"]
	var urlCreationRequest UrlCreationRequest

	urlCreationRequest.LongUrl = store.GetLongURL(shortPath)
	if urlCreationRequest.LongUrl == "" {
		respondWithError(w, http.StatusNotFound, "Not found")
		return
	}
	urlEntry := store.GetURLEntry(urlCreationRequest.LongUrl)
	fmt.Printf("Found")
	sendResponse(w, http.StatusOK, urlEntry)
}

// Delete short link
func DeleteShortUrl(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortPath := params["urlshorten"]

	key := base62.Decode(shortPath)
	check := store.DeleteShortURL(key)
	if !check {
		sendResponse(w, http.StatusBadRequest, map[string]string{"message": "delete failed"})
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "delete successful"})
}

//Update a new long url for shor url
func UpdateUrl(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortUrl := params["urlshorten"]

	var urlCreationRequest UrlCreationRequest
	var updateUrlEntry shorten.URLEntry

	err := json.NewDecoder(r.Body).Decode(&urlCreationRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := base62.Decode(shortUrl)
	longurl := store.GetLongURL(shortUrl)
	urlEntry := store.GetURLEntry(longurl)
	timeNow := time.Now().UTC()

	updateUrlEntry = Set(id, urlCreationRequest.LongUrl, prefixLink+shortUrl, 0, urlEntry.CreateAt, timeNow)
	check := store.UpdateURL(updateUrlEntry)

	if !check {
		sendResponse(w, http.StatusBadRequest, map[string]string{"message": "update failed"})
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "update successful"})
}

// Check url
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

// respond
func respondWithError(w http.ResponseWriter, code int, message string) {
	sendResponse(w, code, map[string]string{"error": message})
}
func sendResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
