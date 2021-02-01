package model

import (
	"encoding/json"
	"os"
	"testing"
)

// Ensures that the library can detect if a GeoPoint is in a GeoPolygon.
// Uses Brunei and the capital of Brunei as a set of test points.
func TestPointInPolygon(t *testing.T) {
	brunei, err := polygonFromFile("test/data/brunei.json")
	if err != nil {
		t.Error("brunei json file failed to parse: ", err)
	}

	GeoPoint := GeoPoint{Lng: 114.9480600, Lat: 4.9402900}
	if !brunei.Contains(&GeoPoint) {
		t.Error("Expected the capital of Brunei to be in Brunei, but it wasn't.")
	}
}

// Ensures that the GeoPolygon logic can correctly identify if a GeoPolygon does not contain a GeoPoint.
// Uses Brunei, Seattle, and a GeoPoint directly outside of Brunei limits as test points.
func TestPointNotInPolygon(t *testing.T) {
	brunei, err := polygonFromFile("test/data/brunei.json")
	if err != nil {
		t.Error("brunei json file failed to parse: ", err)
	}

	// Seattle, WA should not be inside of Brunei
	GeoPoint := NewPoint(47.45, 122.30)
	if brunei.Contains(GeoPoint) {
		t.Error("Seattle, WA [47.45, 122.30] should not be inside of Brunei")
	}

	// A GeoPoint just outside of the successful bounds in Brunei
	// Should not be contained in the GeoPolygon
	precision := NewPoint(114.659596, 4.007636)
	if brunei.Contains(precision) {
		t.Error("A GeoPoint just outside of Brunei should not be contained in the GeoPolygon")
	}
}

// Ensures that a GeoPoint can be contained in a complex GeoPolygon (e.g. a donut)
// This particular GeoPolygon has a hole in it.
func TestPointInPolygonWithHole(t *testing.T) {
	nsw, err := polygonFromFile("test/data/nsw.json")
	if err != nil {
		t.Error("nsw json file failed to parse: ", err)
	}

	act, err := polygonFromFile("test/data/act.json")
	if err != nil {
		t.Error("act json file failed to parse: ", err)
	}

	// Look at two contours
	canberra := GeoPoint{Lng: 149.128684300000030000, Lat: -35.2819998}
	isnsw := nsw.Contains(&canberra)
	isact := act.Contains(&canberra)
	if !isnsw && !isact {
		t.Error("Canberra should be in NSW and also in the sub-contour ACT state")
	}

	// Using NSW as a multi-contour GeoPolygon
	nswmulti := &GeoPolygon{}
	for _, p := range nsw.Points() {
		nswmulti.Add(p)
	}

	for _, p := range act.Points() {
		nswmulti.Add(p)
	}

	isnsw = nswmulti.Contains(&canberra)
	if isnsw {
		t.Error("Canberra should not be in NSW as it falls in the donut contour of the ACT")
	}

	sydney := GeoPoint{Lng: 151.209, Lat: -33.866}

	if !nswmulti.Contains(&sydney) {
		t.Error("Sydney should be in NSW")
	}

	losangeles := GeoPoint{Lng: 118.28333, Lat: 34.01667}
	isnsw = nswmulti.Contains(&losangeles)

	if isnsw {
		t.Error("Los Angeles should not be in NSW")
	}

}

// Ensures that jumping over the equator and the greenwich meridian
// Doesn't give us any false positives or false negatives
func TestEquatorGreenwichContains(t *testing.T) {
	point1 := NewPoint(0.0, 0.0)
	point2 := NewPoint(0.1, 0.1)
	point3 := NewPoint(0.1, -0.1)
	point4 := NewPoint(-0.1, -0.1)
	point5 := NewPoint(-0.1, 0.1)
	GeoPolygon, err := polygonFromFile("test/data/equator_greenwich.json")

	if err != nil {
		t.Errorf("error parsing GeoPolygon: %v", err)
	}

	if !GeoPolygon.Contains(point1) {
		t.Errorf("Should contain middle GeoPoint of earth")
	}

	if !GeoPolygon.Contains(point2) {
		t.Errorf("Should contain GeoPoint %v", point2)
	}

	if !GeoPolygon.Contains(point3) {
		t.Errorf("Should contain GeoPoint %v", point3)
	}

	if !GeoPolygon.Contains(point4) {
		t.Errorf("Should contain GeoPoint %v", point4)
	}

	if !GeoPolygon.Contains(point5) {
		t.Errorf("Should contain GeoPoint %v", point5)
	}
}

// A test struct used to encapsulate and
// Unmarshal JSON into.
type testPoints struct {
	Points []*GeoPoint
}

// Opens a JSON file and unmarshals the data into a GeoPolygon
func polygonFromFile(filename string) (*GeoPolygon, error) {
	p := &GeoPolygon{}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	points := new(testPoints)
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&points); err != nil {
		return nil, err
	}

	for _, GeoPoint := range points.Points {
		p.Add(GeoPoint)
	}

	return p, nil
}
