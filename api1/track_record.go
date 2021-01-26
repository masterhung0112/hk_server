package api1

func (api *API) InitTrackRecord() {
	api.BaseRoutes.TrackRecords.Handle("", api.ApiSessionRequired(createTrackRecord)).Methods("POST")
  api.BaseRoutes.TrackRecord.Handle("", api.ApiSessionRequired(getTrackRecord)).Methods("GET")
}