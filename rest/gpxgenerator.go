package rest

import (
	"github.com/dfleischhacker/locationhistory-collector/locationdb"
	"github.com/tkrajina/gpxgo/gpx"
)

func GenerateGpx(waypoints []locationhistory.Waypoint) *gpx.GPX {
	gpxDoc := gpx.GPX{}
	track := gpx.GPXTrack{}
	segment := gpx.GPXTrackSegment{}

	for _, wp := range waypoints {
		point := gpx.GPXPoint{}
		point.Longitude = wp.Longitude
		point.Latitude = wp.Latitude
		point.Timestamp = wp.Datetime
		segment.Points = append(segment.Points, point)
	}
	track.Segments = append(track.Segments, segment)
	gpxDoc.Tracks = append(gpxDoc.Tracks, track)

	return &gpxDoc
}

func GetGpxStream(gpxDoc *gpx.GPX) ([]byte, error) {
	return gpxDoc.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
}
