package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
)

const (
	// According to Wikipedia, the Earth's radius is about 6,371km
	EarthRadius = 6371
)

// Represents a Physical GeoPoint in geographic notation [lat, lng].
type GeoPoint struct {
	Lat float64
	Lng float64
}

// NewPoint returns a new GeoPoint populated by the passed in latitude (lat) and longitude (lng) values.
func NewPoint(lat float64, lng float64) *GeoPoint {
	return &GeoPoint{Lat: lat, Lng: lng}
}

// PointAtDistanceAndBearing returns a GeoPoint populated with the lat and lng coordinates
// by transposing the origin GeoPoint the passed in distance (in kilometers)
// by the passed in compass bearing (in degrees).
// Original Implementation from: http://www.movable-type.co.uk/scripts/latlong.html
func (p *GeoPoint) PointAtDistanceAndBearing(dist float64, bearing float64) *GeoPoint {

	dr := dist / EarthRadius

	bearing = (bearing * (math.Pi / 180.0))

	lat1 := (p.Lat * (math.Pi / 180.0))
	lng1 := (p.Lng * (math.Pi / 180.0))

	lat2_part1 := math.Sin(lat1) * math.Cos(dr)
	lat2_part2 := math.Cos(lat1) * math.Sin(dr) * math.Cos(bearing)

	lat2 := math.Asin(lat2_part1 + lat2_part2)

	lng2_part1 := math.Sin(bearing) * math.Sin(dr) * math.Cos(lat1)
	lng2_part2 := math.Cos(dr) - (math.Sin(lat1) * math.Sin(lat2))

	lng2 := lng1 + math.Atan2(lng2_part1, lng2_part2)
	lng2 = math.Mod((lng2+3*math.Pi), (2*math.Pi)) - math.Pi

	lat2 = lat2 * (180.0 / math.Pi)
	lng2 = lng2 * (180.0 / math.Pi)

	return &GeoPoint{Lat: lat2, Lng: lng2}
}

// GreatCircleDistance: Calculates the Haversine distance between two points in kilometers.
// Original Implementation from: http://www.movable-type.co.uk/scripts/latlong.html
func (p *GeoPoint) GreatCircleDistance(p2 *GeoPoint) float64 {
	dLat := (p2.Lat - p.Lat) * (math.Pi / 180.0)
	dLon := (p2.Lng - p.Lng) * (math.Pi / 180.0)

	lat1 := p.Lat * (math.Pi / 180.0)
	lat2 := p2.Lat * (math.Pi / 180.0)

	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)

	a := a1 + a2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}

// BearingTo: Calculates the initial bearing (sometimes referred to as forward azimuth)
// Original Implementation from: http://www.movable-type.co.uk/scripts/latlong.html
func (p *GeoPoint) BearingTo(p2 *GeoPoint) float64 {

	dLon := (p2.Lng - p.Lng) * math.Pi / 180.0

	lat1 := p.Lat * math.Pi / 180.0
	lat2 := p2.Lat * math.Pi / 180.0

	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	brng := math.Atan2(y, x) * 180.0 / math.Pi

	return brng
}

// MidpointTo: Calculates the midpoint between 'this' GeoPoint and the supplied GeoPoint.
// Original implementation from http://www.movable-type.co.uk/scripts/latlong.html
func (p *GeoPoint) MidpointTo(p2 *GeoPoint) *GeoPoint {
	lat1 := p.Lat * math.Pi / 180.0
	lat2 := p2.Lat * math.Pi / 180.0

	lon1 := p.Lng * math.Pi / 180.0
	dLon := (p2.Lng - p.Lng) * math.Pi / 180.0

	bx := math.Cos(lat2) * math.Cos(dLon)
	by := math.Cos(lat2) * math.Sin(dLon)

	lat3Rad := math.Atan2(
		math.Sin(lat1)+math.Sin(lat2),
		math.Sqrt(math.Pow(math.Cos(lat1)+bx, 2)+math.Pow(by, 2)),
	)
	lon3Rad := lon1 + math.Atan2(by, math.Cos(lat1)+bx)

	lat3 := lat3Rad * 180.0 / math.Pi
	lon3 := lon3Rad * 180.0 / math.Pi

	return NewPoint(lat3, lon3)
}

// MarshalBinary renders the current GeoPoint to a byte slice.
// Implements the encoding.BinaryMarshaler Interface.
func (p *GeoPoint) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, p.Lat)
	if err != nil {
		return nil, fmt.Errorf("unable to encode lat %v: %v", p.Lat, err)
	}
	err = binary.Write(&buf, binary.LittleEndian, p.Lng)
	if err != nil {
		return nil, fmt.Errorf("unable to encode lng %v: %v", p.Lng, err)
	}

	return buf.Bytes(), nil
}

func (p *GeoPoint) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)

	var lat float64
	err := binary.Read(buf, binary.LittleEndian, &lat)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}

	var lng float64
	err = binary.Read(buf, binary.LittleEndian, &lng)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}

	p.Lat = lat
	p.Lng = lng
	return nil
}

func (p *GeoPoint) IsValid() bool {
  if p.Lat == float64(0.0) {
    return false
  }

  if p.Lng == float64(0.0) {
    return false
  }

  return true
}

func (p *GeoPoint) ToString() string {
  return fmt.Sprintf(`{"lat":%v, "lng":%v}`, p.Lat, p.Lng)
}

// MarshalJSON renders the current GeoPoint to valid JSON.
// Implements the json.Marshaller Interface.
func (p *GeoPoint) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf(`{"lat":%v, "lng":%v}`, p.Lat, p.Lng)
	return []byte(res), nil
}

// UnmarshalJSON decodes the current GeoPoint from a JSON body.
// Throws an error if the body of the GeoPoint cannot be interpreted by the JSON body
func (p *GeoPoint) UnmarshalJSON(data []byte) error {
	// TODO throw an error if there is an issue parsing the body.
	dec := json.NewDecoder(bytes.NewReader(data))
	var values map[string]float64
	err := dec.Decode(&values)

	if err != nil {
		log.Print(err)
		return err
	}

	*p = *NewPoint(values["lat"], values["lng"])

	return nil
}
