package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

// Tests that a call to NewPoint should return a pointer to a GeoPoint with the specified values assigned correctly.
func TestNewPoint(t *testing.T) {
	p := NewPoint(40.5, 120.5)

	if p == nil {
		t.Error("Expected to get a pointer to a new GeoPoint, but got nil instead.")
	}

	if p.Lat != 40.5 {
		t.Errorf("Expected to be able to specify 40.5 as the Lat value of a new GeoPoint, but got %f instead", p.Lat)
	}

	if p.Lng != 120.5 {
		t.Errorf("Expected to be able to specify 120.5 as the Lng value of a new GeoPoint, but got %f instead", p.Lng)
	}
}

// Tests that calling GetLat() after creating a new GeoPoint returns the expected Lat value.
func TestLat(t *testing.T) {
	p := NewPoint(40.5, 120.5)

	Lat := p.Lat

	if Lat != 40.5 {
		t.Errorf("Expected a call to GetLat() to return the same Lat value as was set before, but got %f instead", Lat)
	}
}

// Tests that calling GetLng() after creating a new GeoPoint returns the expected Lng value.
func TestLng(t *testing.T) {
	p := NewPoint(40.5, 120.5)

	Lng := p.Lng

	if Lng != 120.5 {
		t.Errorf("Expected a call to GetLng() to return the same Lat value as was set before, but got %f instead", Lng)
	}
}

// Seems brittle :\
func TestGreatCircleDistance(t *testing.T) {
	// Test that SEA and SFO are ~ 1091km apart, accurate to 100 meters.
	sea := &GeoPoint{Lat: 47.4489, Lng: -122.3094}
	sfo := &GeoPoint{Lat: 37.6160933, Lng: -122.3924223}
	sfoToSea := 1093.379199082169

	dist := sea.GreatCircleDistance(sfo)

	if !(dist < (sfoToSea+0.1) && dist > (sfoToSea-0.1)) {
		t.Error("Unnacceptable result.", dist)
	}
}

func TestPointAtDistanceAndBearing(t *testing.T) {
	sea := &GeoPoint{Lat: 47.44745785, Lng: -122.308065668024}
	p := sea.PointAtDistanceAndBearing(1090.7, 180)

	// Expected results of transposing GeoPoint
	// ~1091km at bearing of 180 degrees
	resultLat := 37.638557
	resultLng := -122.308066

	withinLatBounds := p.Lat < resultLat+0.001 && p.Lat > resultLat-0.001
	withinLngBounds := p.Lng < resultLng+0.001 && p.Lng > resultLng-0.001
	if !(withinLatBounds && withinLngBounds) {
		t.Error("Unnacceptable result.", fmt.Sprintf("[%f, %f]", p.Lat, p.Lng))
	}
}

func TestBearingTo(t *testing.T) {
	p1 := &GeoPoint{Lat: 40.7486, Lng: -73.9864}
	p2 := &GeoPoint{Lat: 0.0, Lng: 0.0}
	bearing := p1.BearingTo(p2)

	// Expected bearing 60 degrees
	resultBearing := 100.610833

	withinBearingBounds := bearing < resultBearing+0.001 && bearing > resultBearing-0.001
	if !withinBearingBounds {
		t.Error("Unnacceptable result.", fmt.Sprintf("%f", bearing))
	}
}

func TestMidpointTo(t *testing.T) {
	p1 := &GeoPoint{Lat: 52.205, Lng: 0.119}
	p2 := &GeoPoint{Lat: 48.857, Lng: 2.351}

	p := p1.MidpointTo(p2)

	// Expected midpoint 50.5363°N, 001.2746°E
	resultLat := 50.53632
	resultLng := 1.274614

	withinLatBounds := p.Lat < resultLat+0.001 && p.Lat > resultLat-0.001
	withinLngBounds := p.Lng < resultLng+0.001 && p.Lng > resultLng-0.001
	if !(withinLatBounds && withinLngBounds) {
		t.Error("Unnacceptable result.", fmt.Sprintf("[%f, %f]", p.Lat, p.Lng))
	}
}

// Ensures that a GeoPoint can be marhalled into JSON
func TestMarshalJSON(t *testing.T) {
	p := NewPoint(40.7486, -73.9864)
	res, err := json.Marshal(p)

	if err != nil {
		log.Print(err)
		t.Error("Should not encounter an error when attempting to Marshal a GeoPoint to JSON")
	}

	if string(res) != `{"Lat":40.7486,"Lng":-73.9864}` {
		t.Error("GeoPoint should correctly Marshal to JSON")
	}
}

// Ensures that a GeoPoint can be unmarhalled from JSON
func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`{"Lat":40.7486,"Lng":-73.9864}`)
	p := &GeoPoint{}
	err := p.UnmarshalJSON(data)

	if err != nil {
		t.Errorf("Should not encounter an error when attempting to Unmarshal a GeoPoint from JSON")
	}

	if p.Lat != 40.7486 || p.Lng != -73.9864 {
		t.Errorf("GeoPoint has mismatched data after Unmarshalling from JSON")
	}
}

// Ensure that a GeoPoint can be marshalled into slice of binaries
func TestMarshalBinary(t *testing.T) {
	Lat, long := 40.7486, -73.9864
	p := NewPoint(Lat, long)
	actual, err := p.MarshalBinary()
	if err != nil {
		t.Error("Should not encounter an error when attempting to Marshal a GeoPoint to binary", err)
	}

	expected, err := coordinatesToBytes(Lat, long)
	if err != nil {
		t.Error("Unable to convert coordinates to bytes slice.", err)
	}

	if !bytes.Equal(actual, expected) {
		t.Errorf("GeoPoint should correctly Marshal to Binary.\nExpected %v\nBut got %v", expected, actual)
	}
}

// Ensure that a GeoPoint can be unmarshalled from a slice of binaries
func TestUnmarshalBinary(t *testing.T) {
	Lat, long := 40.7486, -73.9864
	coordinates, err := coordinatesToBytes(Lat, long)
	if err != nil {
		t.Error("Unable to convert coordinates to bytes slice.", err)
	}

	actual := &GeoPoint{}
	err = actual.UnmarshalBinary(coordinates)
	if err != nil {
		t.Error("Should not encounter an error when attempting to Unmarshal a GeoPoint from binary", err)
	}

	expected := NewPoint(Lat, long)
	if !assertPointsEqual(actual, expected, 4) {
		t.Errorf("GeoPoint should correctly Marshal to Binary.\nExpected %+v\nBut got %+v", expected, actual)
	}
}

func coordinatesToBytes(Lat, long float64) ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, Lat); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, long); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Asserts true when the latitude and longtitude of p1 and p2 are equal up to a certain number of decimal places.
// Precision is used to define that number of decimal places.
func assertPointsEqual(p1, p2 *GeoPoint, precision int) bool {
	roundedLat1, roundedLng1 := int(p1.Lat*float64(precision))/precision, int(p1.Lng*float64(precision))/precision
	roundedLat2, roundedLng2 := int(p2.Lat*float64(precision))/precision, int(p2.Lng*float64(precision))/precision
	return roundedLat1 == roundedLat2 && roundedLng1 == roundedLng2
}
