package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type server struct{}

// Define a mutex to synchronize access to the cache map
var cacheMutex sync.Mutex

// getContent fetches data from the Overpass API based on the specified parameters.
func getContent(radius string, LAT float64, LNG float64, keyInput string, valueInput string) ([]byte, error) {
	// Construct the Overpass API query URL using the input parameters
	queryURL := fmt.Sprintf("https://overpass-api.de/api/interpreter?data=(node(around:%s,%v,%v)[%s='%s'][name];>;);out;", radius, LAT, LNG, keyInput, valueInput)

	// Send a GET request to the Overpass API with the constructed URL
	resp, err := http.Get(queryURL)
	if err != nil {
		// Log the error if the GET request failed
		log.Printf("Error getting content: %v", err)
		return nil, err
	}

	// Ensure that the response body is closed after the function returns
	defer resp.Body.Close()

	// If the response status code is not OK, return an error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	// Read the response body into a byte array
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Log the error if reading the response body failed
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	return data, nil
}

// ServeHTTP handles HTTP requests and writes an XML response with nearby places information
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set the response header to specify the content type
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	// Get the query parameters from the request URL
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	radius := r.URL.Query().Get("radius")

	// Parse the radius parameter into an integer
	radiusInt, err := strconv.ParseInt(radius, 10, 64)
	if err != nil {
		// Handle any errors that occur during parsing
		log.Printf("Error parsing radius parameter: %v", err)
	}

	// Get nearby places information based on the query parameters
	places := getNearbyPlaces(key, value, radiusInt)

	// Construct an XML response with the query parameters and nearby places information
	xmlPrint :=
		"<company>Shahsuvarov corp.</company>" + "\n" +
			"\t<product>Geoloc</product>" + "\n" +
			"\t<key>" + key + "</key>\n" +
			"\t<value>" + value + "</value>\n" +
			"\t<radius>" + radius + "</radius>\n" +
			"\t<places>" + places + "</places>"

	// Write the XML response to the response writer
	data, err := w.Write([]byte(xmlPrint))
	if data != 0 {
		// Log the number of bytes written to the response writer
		log.Printf("Number of bytes written to response writer: %v", data)
	}
	if err == nil {
		// Log any errors that occur during writing to the response writer
		log.Printf("Error writing XML response to response writer: %v", err)
	}
}

// getLatLonJSON sends a POST request to the Google Geocoding API to obtain the
// latitude and longitude as JSON.
func getLatLonJSON() string {
	// Define the URL for the Google Geocoding API.
	url := "https://www.googleapis.com/geolocation/v1/geolocate?key=API_KEY"

	// Define an empty JSON string and create a new HTTP request with it.
	var jsonStr = []byte(``)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	// Set the necessary headers for the request.
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Vary", "Origin")
	req.Header.Set("Vary", "X-Origin")
	req.Header.Set("Vary", "Referer")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Date", "Wed, 01 Jul 2020 14:21:50 GMT")
	req.Header.Set("Server", "scaffolding on HTTPServer2")
	req.Header.Set("Cache-Control", "private")
	req.Header.Set("X-XSS-Protection", "0")
	req.Header.Set("X-Frame-Options", "SAMEORIGIN")
	req.Header.Set("X-Content-Type-Options", "nosniff")
	req.Header.Set("Alt-Svc", "h3-27="+":443"+"; ma=2592000,h3-25="+":443"+"; ma=2592000,h3-T050="+":443"+"; ma=2592000,h3-Q050="+":443"+"; ma=2592000,h3-Q046="+":443"+"; ma=2592000,h3-Q043="+":443"+"; ma=2592000,quic="+":443"+"; ma=2592000; v="+"46,43"+"")
	req.Header.Set("Transfer-Encoding", "chunked")

	// Check if there was an error creating the request and log it if so.
	if err != nil {
		log.Printf("error creating HTTP request: %v", err)
	}

	// Send the HTTP request and check for errors.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending HTTP request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Read the response body and convert it to a string.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
		return ""
	}

	return string(body)
}

// UserGeolocation Represents user geolocation object with Latitude and Longitute attributes. Is used for parsing lat and lng.
type UserGeolocation struct {
	Accuracy int64 `json:"accuracy"`
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
}

// getLatLon parses the latitude and longitude from structured JSON.
func getLatLon(locJSON string) (float64, float64) {
	// Initialize a UserGeolocation object.
	var loc UserGeolocation

	// Unmarshal the JSON into the UserGeolocation object.
	err := json.Unmarshal([]byte(locJSON), &loc)
	if err != nil {
		// Handle any errors that occurred during the unmarshaling process.
		log.Printf("Error in getting lat and lng: %v", err)
	}

	// Return the latitude and longitude.
	return loc.Location.Lat, loc.Location.Lng
}

// getNearbyPlaces is a function that takes in user's key and value input as well as a radius, gets the user's latitude and longitude using
// the Google Geocoding API, and uses these values to make a request to the Google Places API to get nearby places. It then returns the
// response data as a string.
func getNearbyPlaces(key string, value string, radius int64) string {
	// Check if the data is already cached
	cacheKey := fmt.Sprintf("%s_%s_%d", key, value, radius)
	cacheMutex.Lock()
	if data, ok := cache[cacheKey]; ok {
		cacheMutex.Unlock()
		return data
	}
	cacheMutex.Unlock()

	// If the data is not cached, fetch it from the Overpass API
	LAT, LNG := getLatLon()
	data, err := getContent(strconv.FormatInt(radius, 10), LAT, LNG, key, value)
	if err != nil {
		log.Printf("Error getting content: %v", err)
		return ""
	}

	// Convert the response data to a string and cache it
	places := string(data)
	cacheMutex.Lock()
	cache[cacheKey] = places
	cacheMutex.Unlock()

	return places
}

// main function runs the HTTP server to listen on incoming requests on port 8081
func main() {

	s := &server{} // TO RUN THE SERVER TO HAVE HTTP REQUESTS FROM OUTSIDE (E.G POSTMAN)
	http.Handle("/location-api/", s)
	log.Fatal(http.ListenAndServe(":8081", nil))

}
