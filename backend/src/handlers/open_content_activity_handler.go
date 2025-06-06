package handlers

import (
	"UnlockEdv2/src/models"
	"net/http"
	"strconv"
)

func (srv *Server) registerOpenContentActivityRoutes() []routeDef {
	axx := models.OpenContentAccess
	return []routeDef{
		featureRoute("GET /api/open-content/activity", srv.handleGetTopFacilityOpenContent, axx),
		featureRoute("GET /api/open-content/activity/{id}", srv.handleGetTopUserOpenContent, axx),
		featureRoute("GET /api/libraries/activity", srv.handleGetTopFacilityLibraries, axx),
	}
}

func (srv *Server) handleGetTopFacilityOpenContent(w http.ResponseWriter, r *http.Request, log sLog) error {
	facilityId := srv.getFacilityID(r)
	topOpenContent, err := srv.Db.GetTopFacilityOpenContent(int(facilityId))
	if err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, topOpenContent)
}

func (srv *Server) handleGetTopUserOpenContent(w http.ResponseWriter, r *http.Request, log sLog) error {
	args := srv.getQueryContext(r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return newInvalidIdServiceError(err, "user ID")
	}
	topOpenContent, err := srv.Db.GetTopUserOpenContent(id, &args)
	if err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, topOpenContent)
}

func (srv *Server) handleGetTopFacilityLibraries(w http.ResponseWriter, r *http.Request, log sLog) error {
	args := srv.getQueryContext(r)
	facilityId := srv.getFacilityID(r)
	_, perPage := srv.getPaginationInfo(r)
	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil {
		days = -1
	}
	topLibraries, err := srv.Db.GetTopFacilityLibraries(&args, int(facilityId), perPage, days)
	if err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, topLibraries)
}
