package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type server struct{}

func getContent(radius string, LAT float64, LNG float64, keyInput string, valueInput string) ([]byte, error) {
	resp, err := http.Get("https://overpass-api.de/api/interpreter?data=(node(around:" + radius + "," + fmt.Sprintf("%v", LAT) + "," + fmt.Sprintf("%v", LNG) + ")[" + keyInput + "='" + valueInput + "'][name];>;);out;")
	
	if err != nil {
	log.Printf("error in getting the content is: %v", err)
	}


	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, errtwo := ioutil.ReadAll(resp.Body)
	
	if errtwo == nil {
	log.Printf("error is empty: %v", err)
	}
	
	return data, nil
}

// GET REQUEST
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml") // ADJUST YOUR RETURN TYPE HERE can be e.g application/json
	w.WriteHeader(http.StatusOK)

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	radius := r.URL.Query().Get("radius")

	// Radius is parsed here
	radiusInt, err := strconv.ParseInt(radius, 10, 64)
	if err != nil {
		// handle error
		log.Printf("error in serve http is: %v", err)
		
	}

	places := getNearbyPlaces(key, value, radiusInt)

	xmlPrint :=

		"<company>Shahsuvarov corp.</company>" + "\n" +
			"\t<product>GeoLoc.</product>" + "\n" +
			"\t<key>" + key + "</key>\n" +
			"\t<value>" + value + "</value>\n" +
			"\t<radius>" + radius + "</radius>\n" +
			"\t<places>" + places + "</places>"

	data,err := w.Write([]byte(xmlPrint))
	
	if data != 0 {
	log.Printf("Data is : %v", data)
	}
	
	if err == nil {
	log.Printf("No errors ocurred: %v", err)
	}

}

// For getting the latitude and longitude as JSON from GOOGLE GEOCODING API request
func getLatLonJSON() string {
	// POST REQUEST TO GOOGLE GEOCODING API WITH API KEY
	url := "https://www.googleapis.com/geolocation/v1/geolocate?key=YOUR_API_KEY"

	var jsonStr = []byte(``)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
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
	
	if err != nil {
	log.Printf("error in getting the lat and lng: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

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

// For parsing the latitude and longitude from strutured JSON
func getLatLon(locJSON string) (float64, float64) {

	var loc UserGeolocation
	err := json.Unmarshal([]byte(locJSON), &loc)
	
	if err != nil {
	log.Printf("err in getting lat and lng: %v", err)
	}
	
	return loc.Location.Lat, loc.Location.Lng
}

func getNearbyPlaces(keyInput, valueInput string, radius int64) string {

	var LAT, LNG = getLatLon(getLatLonJSON())
	fmt.Printf("LAT: %v\nLNG: %v", LAT, LNG)
	data, err := getContent(fmt.Sprintf("%v",radius), LAT, LNG, keyInput, valueInput);
	if err != nil {
		log.Printf("Failed to get XML: %v", err)
		return ""
	} 
		return string(data)
	

}

func main() {

	s := &server{} // TO RUN THE SERVER TO HAVE HTTP REQUESTS FROM OUTSIDE (E.G POSTMAN)
	http.Handle("/location-api/", s)
	log.Fatal(http.ListenAndServe(":8081", nil))

}
