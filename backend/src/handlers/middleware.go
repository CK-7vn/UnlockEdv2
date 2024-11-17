package handlers

import (
	"UnlockEdv2/src/models"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

const (
	CsrfTokenCtx contextKey = "csrf_token"
	timeWindow              = time.Minute
	maxRequests  int        = 50
	libraryKey   contextKey = "library"
	videoKey     contextKey = "video"
	// rate limit is 50 requests from a unique user in a minute
)

// regular expression used below for filtering open_content_urls
var resourceRegExpression = regexp.MustCompile(`\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|ttf|map|webp|otf|vtt|webm|json|woff2|pdf)(\?|%3F|$)`)

func (srv *Server) applyMiddleware(h HttpFunc) http.Handler {
	return srv.applyStandardMiddleware(
		srv.handleError(h))
}

func (srv *Server) applyAdminMiddleware(h HttpFunc) http.Handler {
	return srv.applyStandardMiddleware(
		srv.adminMiddleware(
			srv.handleError(h)))
}

func (srv *Server) applyStandardMiddleware(next http.Handler) http.Handler {
	return srv.prometheusMiddleware(srv.setCsrfTokenMiddleware(
		srv.rateLimitMiddleware(
			srv.authMiddleware(
				next))))
}

func (srv *Server) videoProxyMiddleware(next http.Handler) http.Handler {
	return srv.applyStandardMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resourceID := r.PathValue("id")
		var video models.Video
		tx := srv.Db.Model(&models.Video{}).Where("id = ?", resourceID)
		user := r.Context().Value(ClaimsKey).(*Claims)
		switch user.Role {
		case models.Admin:
			tx = tx.First(&video)
		default:
			tx = tx.First(&video, "visibility_status = true AND availability = 'available'")
		}
		if err := tx.Error; err != nil {
			srv.errorResponse(w, http.StatusNotFound, "Video not found, is not available or visibility is not enabled")
			return
		}
		ctx := context.WithValue(r.Context(), videoKey, &video)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (srv *Server) libraryProxyMiddleware(next http.Handler) http.Handler {
	return srv.applyStandardMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		libraryBucket := srv.buckets[LibraryPaths]
		resourceID := r.PathValue("id")
		user := r.Context().Value(ClaimsKey).(*Claims)
		entry, err := libraryBucket.Get(resourceID)
		var proxyParams models.LibraryProxyPO
		if err == nil { //found in bucket going to use cached slice of bytes
			err = json.Unmarshal(entry.Value(), &proxyParams)
			if err != nil {
				srv.errorResponse(w, http.StatusNotFound, "Library not found, issue unmarshaling libary values")
				return
			}
		} else { //not in bucket going to call database to get required data and then add to bucket
			var library models.Library
			if srv.Db.Debug().Model(&models.Library{}).Preload("OpenContentProvider").Where("id = ?", resourceID).First(&library).RowsAffected == 0 {
				srv.errorResponse(w, http.StatusNotFound, "Library not found.")
				return
			}
			proxyParams = models.LibraryProxyPO{
				ID:                    library.ID,
				Path:                  library.Path,
				BaseUrl:               library.OpenContentProvider.BaseUrl,
				OpenContentProviderID: library.OpenContentProvider.ID,
				VisibilityStatus:      library.VisibilityStatus,
			}
			marshaledParams, err := json.Marshal(proxyParams)
			if err != nil {
				srv.errorResponse(w, http.StatusNotFound, "Library not found, issue marshaling libary values")
				return
			}
			if _, err := libraryBucket.Put(resourceID, marshaledParams); err != nil {
				srv.errorResponse(w, http.StatusNotFound, "Library not found, issue adding parameters to cache")
				return
			}
		}
		if user.Role == models.Student && !proxyParams.VisibilityStatus {
			srv.errorResponse(w, http.StatusNotFound, "Visibility is not enabled")
			return
		}
		urlString := r.URL.String()
		if !resourceRegExpression.MatchString(urlString) && !strings.Contains(urlString, "iframe") {
			activity := models.OpenContentActivity{
				OpenContentProviderID: proxyParams.OpenContentProviderID,
				FacilityID:            user.FacilityID,
				UserID:                user.UserID,
				ContentID:             proxyParams.ID,
			}
			srv.createActivity(urlString, activity)
		}
		ctx := context.WithValue(r.Context(), libraryKey, &proxyParams)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}

func (srv *Server) createActivity(urlString string, activity models.OpenContentActivity) {
	url := models.OpenContentUrl{}
	if srv.Db.Where("content_url = ?", urlString).First(&url).RowsAffected == 0 {
		url.ContentURL = urlString
		if err := srv.Db.Create(&url).Error; err != nil {
			log.Warn("unable to create content url for activity")
			return
		}
	}
	activity.OpenContentUrlID = url.ID
	if err := srv.Db.Create(&activity).Error; err != nil {
		log.Warn("unable to create content activity for url, ", urlString)
	}
	// var fav models.OpenContentFavorite
	// if srv.Db.Debug().Model(&models.OpenContentFavorite{}).Where("user_id = ? AND content_id = ? AND open_content_url_id = ?", activity.UserID, activity.ContentID, activity.OpenContentUrlID).First(&fav).RowsAffected > 0 {
	// 	srv.wsClient.notifyUser(activity.UserID, []byte("true"))
	// } else {
	// 	srv.wsClient.notifyUser(activity.UserID, []byte("false"))
	// }
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	}
}

func (srv *Server) setCsrfTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := log.Fields{"handler": "setCsrfTokenMiddleware"}
		bucket := srv.buckets[CsrfToken]
		checkExists, err := r.Cookie("unlocked_csrf_token")
		if err == nil {
			fields["csrf_token"] = checkExists.Value
			// now we check if the token is validjA
			val, err := bucket.Get(checkExists.Value)
			if err != nil {
				srv.clearKratosCookies(w, r)
				if !isAuthRoute(r) {
					http.Redirect(w, r, fmt.Sprintf("%s/browser?return_to=%s", LoginEndpoint, r.URL.Path), http.StatusSeeOther)
					log.WithFields(fields).Traceln("CSRF token is invalid, redirecting user")
					return
				}
			} else {
				log.WithFields(fields).Traceln("CSRF token is valid")
				ctx := context.WithValue(r.Context(), CsrfTokenCtx, string(val.Value()))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		uniqueId := uuid.NewString()
		http.SetCookie(w, &http.Cookie{
			Name:     "unlocked_csrf_token",
			Value:    uniqueId,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
		})
		_, err = bucket.Put(uniqueId, []byte(time.Now().Add(24*time.Hour).String()))
		if err != nil {
			log.WithFields(fields).Errorf("Failed to set CSRF token: %v", err)
			srv.errorResponse(w, http.StatusInternalServerError, "failed to write CSRF token")
			return
		}
		ctx := context.WithValue(r.Context(), CsrfTokenCtx, string(uniqueId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := log.Fields{"handler": "rateLimitMiddleware"}
		kv := srv.buckets[RateLimit]
		hashedValue, err := getUniqueRequestInfo(r)
		if err != nil {
			log.WithFields(fields).Errorf("Failed to get unique request info: %v", err)
			srv.errorResponse(w, http.StatusInternalServerError, "failed to write CSRF token")
			return
		}
		value, err := kv.Get(hashedValue)
		if err != nil {
			// create a new requestInfo
			reqInfo := &requestInfo{
				Count:     1,
				Timestamp: time.Now(),
			}
			if err := putRequestInfo(kv, hashedValue, reqInfo); err != nil {
				log.WithFields(fields).Errorf("Failed to marshal request info: %v", err)
				srv.errorResponse(w, http.StatusInternalServerError, "failed to write CSRF token")
				return
			}
			next.ServeHTTP(w, r)
			return
		} else {
			var reqInfo requestInfo
			if err := json.Unmarshal(value.Value(), &reqInfo); err != nil {
				log.WithFields(fields).Errorf("Failed to unmarshal request info: %v", err)
				srv.errorResponse(w, http.StatusInternalServerError, "failed to decode request info")
				return
			}
			if time.Since(reqInfo.Timestamp) > timeWindow {
				reqInfo.Count = 0
				reqInfo.Timestamp = time.Now()
			} else {
				reqInfo.Count++
				if reqInfo.Count > maxRequests {
					srv.errorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
					return
				}
			}
			if err := putRequestInfo(kv, hashedValue, &reqInfo); err != nil {
				log.WithFields(fields).Errorf("Failed to marshal request info: %v", err)
				srv.errorResponse(w, http.StatusInternalServerError, "failed to write CSRF token")
				return
			}
			next.ServeHTTP(w, r)
			return
		}
	})
}

func putRequestInfo(kv nats.KeyValue, key string, reqInfo *requestInfo) error {
	bytes, err := json.Marshal(reqInfo)
	if err != nil {
		return err
	}
	if _, err := kv.Put(key, bytes); err != nil {
		return err
	}
	return nil
}

type requestInfo struct {
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

func getUniqueRequestInfo(r *http.Request) (string, error) {
	csrf, ok := r.Context().Value(CsrfTokenCtx).(string)
	if !ok {
		return "", errors.New("CSRF token not found")
	}
	uniq := r.Header.Get("X-Real-IP")
	if uniq == "" {
		uniq = r.Header.Get("X-Forwarded-For")
		if uniq == "" {
			uniq = r.RemoteAddr
		}
	}
	unique := r.Header.Get("User-Agent") + uniq + csrf
	hashedValue := shaHashValue(unique)
	return hashedValue, nil
}

func shaHashValue(value string) string {
	hash := sha256.New()
	hash.Write([]byte(value))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
