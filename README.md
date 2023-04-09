# GeoLoc

This project is an HTTP server that responds to requests with information about nearby places, based on the provided query parameters. The server sends a GET request to the Overpass API, which returns data about the requested places. The server then constructs an XML response with the query parameters and nearby places information.

## Prerequisites

- Go 1.15 or higher

## How to use

1. Clone this repository.
2. Set your Google Geocoding API key as an environment variable named `API_KEY`.
3. Navigate to the cloned directory and run `go build` to build the binary.
4. Run the binary with `./binary-name`.
5. Send HTTP requests to `http://localhost:8080?key=<KEY>&value=<VALUE>&radius=<RADIUS>`, where `KEY` is the key for the requested places, `VALUE` is the value for the requested places, and `RADIUS` is the search radius in meters.

## Code Overview

- `getContent` function fetches data from the Overpass API based on the specified parameters.
- `ServeHTTP` method handles HTTP requests and writes an XML response with nearby places information.
- `getLatLonJSON` function sends a POST request to the Google Geocoding API to obtain the latitude and longitude as JSON.

## Contributors

- [Murad Shahsuvarov](https://github.com/muradshahsuvarov/geoloc) - maintainer
