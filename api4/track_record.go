package api1

import (
	"net/http"

	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/audit"
	"github.com/masterhung0112/hk_server/model"
)

func (api *API) InitTrackRecord() {
	api.BaseRoutes.TrackRecords.Handle("", api.ApiSessionRequired(createTrackRecord)).Methods("POST")
	api.BaseRoutes.TrackRecord.Handle("", api.ApiSessionRequired(getTrackRecord)).Methods("GET")
	api.BaseRoutes.TrackRecord.Handle("/trackpoint", api.ApiSessionRequired(createNewTrackPointForTrackRecord)).Methods("POST")
	api.BaseRoutes.TrackRecord.Handle("/start", api.ApiSessionRequired(startTrackRecord)).Methods("POST")
	api.BaseRoutes.TrackRecord.Handle("/end", api.ApiSessionRequired(endTrackRecord)).Methods("POST")
}

func createTrackRecord(c *Context, w http.ResponseWriter, r *http.Request) {
	trackRecord := model.TrackRecordFromJson(r.Body)
	if trackRecord == nil {
		c.SetInvalidParam("trackRecord")
		return
	}

	auditRec := c.MakeAuditRecord("createTrackRecord", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)
	auditRec.AddMeta("trackRecord", trackRecord)

	rTrackRecord, err := c.App.CreateTrackRecord(trackRecord)
	if err != nil {
		c.Err = err
		return
	}
	auditRec.Success()
	auditRec.AddMeta("trackRecord", rTrackRecord)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rTrackRecord.ToJson()))
}

func getTrackRecord(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTrackRecordId()
	if c.Err != nil {
		return
	}

	trackRecord, err := c.App.GetTrackRecord(c.Params.TrackRecordId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(trackRecord.ToJson()))
}

func createNewTrackPointForTrackRecord(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTrackRecordId()
	if c.Err != nil {
		return
	}

	trackPoint := model.TrackPointFromJson(r.Body)
	if trackPoint == nil {
		c.SetInvalidParam("trackPoint")
		return
	}

	rTrackPoint, err := c.App.CreateTrackPointForTrackRecord(trackPoint, c.Params.TrackRecordId)
	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rTrackPoint.ToJson()))
}

func startTrackRecord(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTrackRecordId()
	if c.Err != nil {
		return
	}

	trackRecord, err := c.App.StartTrackRecord(c.Params.TrackRecordId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(trackRecord.ToJson()))
}

func endTrackRecord(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTrackRecordId()
	if c.Err != nil {
		return
	}

	trackRecord, err := c.App.EndTrackRecord(c.Params.TrackRecordId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(trackRecord.ToJson()))
}
