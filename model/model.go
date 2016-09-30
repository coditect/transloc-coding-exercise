package model

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
)

// A Location describes a point on the earth's surface.
type Location struct {
	Latitude, Longitude float64
}

// A LocationTable associates locations with the quantity of IP addresses
// assigned there.
type LocationTable map[Location]float64

// A LocationStorage provides a means of storing and querying IP location data.
type LocationStorage interface {
	Query(north, south, east, west float64) (LocationTable, error)
	Save(LocationTable) error
}

// RoundLocations reduces the resolution of data in a location table by rounding
// each latitude to the nearest multiple of y degrees and each longitude to the
// nearest multiple of x degrees, and combining the quantities assigned to
// locations that round to the same pair of coordinates.
func (t LocationTable) RoundLocations(x, y float64) LocationTable {
	rounded := make(LocationTable)

	for location, quantity := range t {
		roundedLocation := Location{
			RoundToNearestMultiple(location.Latitude, y),
			RoundToNearestMultiple(location.Longitude, x),
		}

		if previousQuanity, found := rounded[roundedLocation]; found {
			rounded[roundedLocation] = quantity + previousQuanity
		} else {
			rounded[roundedLocation] = quantity
		}
	}

	return rounded
}

// Logarithmic returns a location table in which the quantity assigned to each
// location is the base
func (t LocationTable) Logarithmic(base float64) LocationTable {
	logTable := make(LocationTable)
	logBase := math.Log(base)

	for location, quantity := range t {
		logTable[location] = math.Log(quantity) / logBase
	}

	return logTable
}

// Flatten converts a location table into a slice of 3-item arrays, each of
// which contains the latitude, longitude, and quantity of a location.
func (t LocationTable) Flatten() [][3]float64 {
	flattened := make([][3]float64, len(t))
	i := 0

	for location, quantity := range t {
		flattened[i] = [3]float64{location.Latitude, location.Longitude, quantity}
		i++
	}

	return flattened
}

// MarshalJSON implements the json.Marshaler interface.
func (t LocationTable) MarshalJSON() ([]byte, error) {
	flattened := t.Flatten()
	return json.Marshal(flattened)
}

// LocationTableFromCSV constructs a location table from a CSV file.  The first
// line of the CSV file must contain column names; this function expects columns
// named latitude, longitude, and network.
func LocationTableFromCSV(r io.Reader) (LocationTable, error) {
	table := make(LocationTable)
	csvReader := csv.NewReader(r)

	// First line contains column names
	columnNames, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	var networkIndex, latitudeIndex, longitudeIndex int
	var networkFound, latitudeFound, longitudeFound bool
	for i, name := range columnNames {
		switch name {
		case "network":
			networkFound = true
			networkIndex = i
		case "latitude":
			latitudeFound = true
			latitudeIndex = i
		case "longitude":
			longitudeFound = true
			longitudeIndex = i
		}
	}

	// No rounding: 30968 rows
	// Floor:       ~2000 rows
	// Floor to nearest 10th of a degree: 17366 rows

	if !networkFound || !latitudeFound || !longitudeFound {
		return nil, HTTPError{fmt.Errorf("Missing required columns"), 400}
	}

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		latitude, err := strconv.ParseFloat(record[latitudeIndex], 64)
		if err != nil {
			return nil, err
		}

		longitude, err := strconv.ParseFloat(record[longitudeIndex], 64)
		if err != nil {
			return nil, err
		}

		_, network, err := net.ParseCIDR(record[networkIndex])
		if err != nil {
			return nil, err
		}

		location := Location{latitude, longitude}
		prefixSize, totalSize := network.Mask.Size()
		var quantity float64

		if totalSize - prefixSize < 1 {
			quantity = 1
		} else {
			quantity = math.Pow(2, float64(totalSize - prefixSize - 1))
		}

		if previousQuanity, found := table[location]; found {
			table[location] = quantity + previousQuanity
		} else {
			table[location] = quantity
		}
	}

	return table, nil
}

// NormalizeLongitude normalizes a longitude value to the range [-180, 180).
func NormalizeLongitude(degrees float64) float64 {
	degrees = math.Mod(degrees, 360)
	if degrees < -180 {
		degrees += 360
	} else if degrees >= 180 {
		degrees -= 360
	}
	return degrees
}

// RoundToNearestMultiple rounds n to the nearest multiple of m.
func RoundToNearestMultiple(n, m float64) float64 {
	if m == 0 {
		return n
	}
	return math.Floor((n + m / 2) / m) * m
}
