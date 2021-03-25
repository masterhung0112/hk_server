package model

import "fmt"

// Track Record section

// CreateTrackRecord creates a track record in the system based on the provided user struct.
func (c *Client1) CreateTrackRecord(trackRecord *TrackRecord) (*TrackRecord, *Response) {
	r, err := c.DoApiPost(c.GetTrackRecordsRoute(), trackRecord.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TrackRecordFromJson(r.Body), BuildResponse(r)
}

// CreateTrackRecord creates a track record in the system based on the provided user struct.
func (c *Client1) CreateTrackPointForRecord(trackRecordId string, trackPoint *TrackPoint) (*TrackPoint, *Response) {
	r, err := c.DoApiPost(c.CreateTrackPointForRecordRoute(trackRecordId), trackPoint.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TrackPointFromJson(r.Body), BuildResponse(r)
}

func (c *Client1) StartTrackRecord(trackRecordId string) (*TrackRecord, *Response) {
	r, err := c.DoApiPost(c.GetTrackRecordStartRoute(trackRecordId), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TrackRecordFromJson(r.Body), BuildResponse(r)
}

func (c *Client1) EndTrackRecord(trackRecordId string) (*TrackRecord, *Response) {
	r, err := c.DoApiPost(c.GetTrackRecordEndRoute(trackRecordId), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TrackRecordFromJson(r.Body), BuildResponse(r)
}

// Track Record section

func (c *Client4) GetTrackRecordsRoute() string {
	return "/trackrecords"
}

func (c *Client4) GetTrackRecordRoute(trackRecordId string) string {
	return fmt.Sprintf(c.GetTrackRecordsRoute()+"/%v", trackRecordId)
}

func (c *Client4) CreateTrackPointForRecordRoute(trackRecordId string) string {
	return c.GetTrackRecordRoute(trackRecordId) + "/trackpoint"
}

func (c *Client4) GetTrackRecordStartRoute(trackRecordId string) string {
	return c.GetTrackRecordRoute(trackRecordId) + "/start"
}

func (c *Client4) GetTrackRecordEndRoute(trackRecordId string) string {
	return c.GetTrackRecordRoute(trackRecordId) + "/end"
}
