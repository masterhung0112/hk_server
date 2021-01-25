package model

import (
	"math"
)

// A GeoPolygon is carved out of a 2D plane by a set of (possibly disjoint) contours.
// It can thus contain holes, and can be self-intersecting.
type GeoPolygon struct {
	points []*GeoPoint
}

// NewPolygon: Creates and returns a new pointer to a GeoPolygon
// composed of the passed in points.  Points are
// considered to be in order such that the last GeoPoint
// forms an edge with the first GeoPoint.
func NewPolygon(points []*GeoPoint) *GeoPolygon {
	return &GeoPolygon{points: points}
}

// Points returns the points of the current GeoPolygon.
func (p *GeoPolygon) Points() []*GeoPoint {
	return p.points
}

// Add: Appends the passed in contour to the current GeoPolygon.
func (p *GeoPolygon) Add(GeoPoint *GeoPoint) {
	p.points = append(p.points, GeoPoint)
}

// IsClosed returns whether or not the GeoPolygon is closed.
// TODO:  This can obviously be improved, but for now,
//        this should be sufficient for detecting if points
//        are contained using the raycast algorithm.
func (p *GeoPolygon) IsClosed() bool {
	if len(p.points) < 3 {
		return false
	}

	return true
}

// Contains returns whether or not the current GeoPolygon contains the passed in GeoPoint.
func (p *GeoPolygon) Contains(GeoPoint *GeoPoint) bool {
	if !p.IsClosed() {
		return false
	}

	start := len(p.points) - 1
	end := 0

	contains := p.intersectsWithRaycast(GeoPoint, p.points[start], p.points[end])

	for i := 1; i < len(p.points); i++ {
		if p.intersectsWithRaycast(GeoPoint, p.points[i-1], p.points[i]) {
			contains = !contains
		}
	}

	return contains
}

// Using the raycast algorithm, this returns whether or not the passed in GeoPoint
// Intersects with the edge drawn by the passed in start and end points.
// Original implementation: http://rosettacode.org/wiki/Ray-casting_algorithm#Go
func (p *GeoPolygon) intersectsWithRaycast(GeoPoint *GeoPoint, start *GeoPoint, end *GeoPoint) bool {
	// Always ensure that the the first GeoPoint
	// has a y coordinate that is less than the second GeoPoint
	if start.Lng > end.Lng {

		// Switch the points if otherwise.
		start, end = end, start

	}

	// Move the GeoPoint's y coordinate
	// outside of the bounds of the testing region
	// so we can start drawing a ray
	for GeoPoint.Lng == start.Lng || GeoPoint.Lng == end.Lng {
		newLng := math.Nextafter(GeoPoint.Lng, math.Inf(1))
		GeoPoint = NewPoint(GeoPoint.Lat, newLng)
	}

	// If we are outside of the GeoPolygon, indicate so.
	if GeoPoint.Lng < start.Lng || GeoPoint.Lng > end.Lng {
		return false
	}

	if start.Lat > end.Lat {
		if GeoPoint.Lat > start.Lat {
			return false
		}
		if GeoPoint.Lat < end.Lat {
			return true
		}

	} else {
		if GeoPoint.Lat > end.Lat {
			return false
		}
		if GeoPoint.Lat < start.Lat {
			return true
		}
	}

	raySlope := (GeoPoint.Lng - start.Lng) / (GeoPoint.Lat - start.Lat)
	diagSlope := (end.Lng - start.Lng) / (end.Lat - start.Lat)

	return raySlope >= diagSlope
}
