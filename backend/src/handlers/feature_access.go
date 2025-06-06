package handlers

import (
	"UnlockEdv2/src/database"
	"UnlockEdv2/src/models"
	"errors"
	"net/http"
)

func (srv *Server) registerFeatureFlagRoutes() []routeDef {
	return []routeDef{
		newAdminRoute("PUT /api/auth/features/{feature}", srv.handleToggleFeatureFlag),
		validatedAdminRoute("POST /api/auth/demo-seed", srv.handleRunDemoSeed, func(db *database.DB, r *http.Request) bool {
			return userIsSystemAdmin(r)
		}),
	}
}

func (srv *Server) handleToggleFeatureFlag(w http.ResponseWriter, r *http.Request, log sLog) error {
	user := r.Context().Value(ClaimsKey).(*Claims)
	if user.Role != models.SystemAdmin {
		return newUnauthorizedServiceError()
	}
	feature := r.PathValue("feature")
	if !models.ValidFeature(models.FeatureAccess(feature)) {
		return newBadRequestServiceError(errors.New("feature_flag"), "invalid feature requested")
	}
	if err := srv.Db.ToggleFeatureAccess(feature); err != nil {
		return newInternalServerServiceError(err, "unable to toggle feature")
	}
	features, err := srv.Db.GetFeatureAccess()
	if err != nil {
		return newInternalServerServiceError(err, "unable to fetch features")
	}
	srv.features = features
	return writeJsonResponse(w, http.StatusOK, "feature toggled successfully")
}

func (srv *Server) handleRunDemoSeed(w http.ResponseWriter, r *http.Request, log sLog) error {
	log.info("running seeder for demo environment")
	err := srv.Db.RunOrResetDemoSeed(srv.getFacilityID(r))
	if err != nil {
		return newInternalServerServiceError(err, "unable to run demo seed")
	}
	return writeJsonResponse(w, http.StatusOK, "demo seed ran successfully")
}
