package main

import (
	"UnlockEdv2/src/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	TokenEndpoint       = "https://auth.brightspace.com/core/connect/token"
	DataSetsEndpoint    = "https://unlocked.brightspacedemo.com/d2l/api/lp/1.28/dataExport/bds/list"
	DataDownloadEnpoint = "https://unlocked.brightspacedemo.com/d2l/api/lp/1.28/dataExport/bds/download/%s"
)

type BrightspaceService struct {
	ProviderPlatformID uint
	Client             *http.Client
	BaseURL            string
	ClientID           string
	ClientSecret       string
	RefreshToken       string
	Scope              string
	AccessToken        string
	BaseHeaders        *map[string]string
	JobParams          *map[string]interface{}
}

func newBrightspaceService(provider *models.ProviderPlatform, db *gorm.DB, params *map[string]interface{}) (*BrightspaceService, error) {
	//scope will be saved as system property???? for now??? until further notice
	keysSplit := strings.Split(provider.AccessKey, ";")
	if len(keysSplit) < 2 {
		return nil, errors.New("unable to find refresh token, unable to intialize BrightspaceService")
	}
	//scope := os.Getenv("BRIGHTSPACE_SCOPE"
	scope := "accommodations:profile:manage,read accountsettings:locale:read,update attributes:schemas:read attributes:users:delete,read,update datahub:dataexports:download,read datasets:bds:read globalusermapping:mapping:write localauthenticationsecurity:overrides:manage,read organizations:image:read role:detail:create,read users:activation:read,update users:own_profile:read users:own_pronoun:read,update users:password:delete,write users:profile:read users:userdata:create,delete,read,update"
	fmt.Println("RTS Scopes-----> ", scope)
	if scope == "" {
		return nil, errors.New("no brightspace scope found, unable to intialize BrightspaceService")
	}
	brightspaceService := BrightspaceService{
		ProviderPlatformID: provider.ID,
		Client:             &http.Client{},
		BaseURL:            provider.BaseUrl,
		ClientID:           provider.AccountID,
		ClientSecret:       keysSplit[0],
		RefreshToken:       keysSplit[1],
		Scope:              scope,
		JobParams:          params,
	}
	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", brightspaceService.RefreshToken)
	data.Add("client_id", brightspaceService.ClientID)
	data.Add("client_secret", brightspaceService.ClientSecret)
	data.Add("scope", brightspaceService.Scope)
	fmt.Println("RTS made it to before send post request -----> ")
	resp, err := brightspaceService.SendPostRequest(TokenEndpoint, data)
	if err != nil {
		return nil, err
	}
	// //check response here!!!
	// if resp.StatusCode != http.StatusOK {
	// 	//grab brightspace's error message
	// 	return nil, errors.New("unable to request new refresh token from brightspace, response error message is: ")
	// }
	fmt.Println("RTS made it to after send post request -----> ")
	var tokenRespMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRespMap); err != nil {
		return nil, err
	}
	fmt.Println("tokenRespMap: ", tokenRespMap)
	brightspaceService.AccessToken = tokenRespMap["access_token"].(string)
	brightspaceService.RefreshToken = tokenRespMap["refresh_token"].(string)
	provider.AccessKey = brightspaceService.ClientSecret + ";" + brightspaceService.RefreshToken
	if err := db.Debug().Save(&provider).Error; err != nil {
		//send admin email??? maybe but not now
		return nil, err
	}
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + brightspaceService.AccessToken
	//headers["Accept"] = "application/json"
	brightspaceService.BaseHeaders = &headers
	return &brightspaceService, nil
}

func (srv *BrightspaceService) SendPostRequest(url string, data url.Values) (*http.Response, error) {
	encodedUrl := data.Encode()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(encodedUrl))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded") //standard header for url.Values (encoded)
	resp, err := srv.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (srv *BrightspaceService) SendRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range *srv.BaseHeaders {
		req.Header.Add(key, value)
	}
	resp, err := srv.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (srv *BrightspaceService) GetUsers(db *gorm.DB) ([]models.ImportUser, error) {
	fmt.Println("GetUsers...")
	return nil, nil
}

func (srv *BrightspaceService) ImportCourses(db *gorm.DB) error {
	//1. get plugin ID
	pluginId, err := srv.getPluginId("Organizational Units")
	if err != nil {
		return err
	}
	fmt.Println("made it, got the pluginId,", pluginId)
	//2. download
	csvFile, err := srv.downloadAndUnzipFile("csvs", "OrganizationalUnits.zip", fmt.Sprintf(DataDownloadEnpoint, pluginId))
	if err != nil {
		return err
	}
	fmt.Println("made it, got the csvFile,", csvFile)

	//3. read CSV file and parse organizational units
	bsCourses := []BrightspaceCourse{}
	ReadCSV(&bsCourses, csvFile)
	for _, bsCourse := range bsCourses {
		if bsCourse.IsActive == "TRUE" && bsCourse.IsDeleted == "FALSE" && bsCourse.Type == "Course Offering" {
			//total progress is going to be on hold for now
			course := srv.IntoCourse(bsCourse)
			fmt.Println(course.Description)
			if err := db.Create(&course).Error; err != nil {
				log.Errorln("error creating course in db, error is: ", err)
				continue
			}
		}
	}

	// keys := []string{"Users", "Quiz Attempts", "Grade Objects Log", "Quiz Attempts Log", "Assignment Submissions", "Organizational Units", "User Enrollments", "Users"}
	//"Organizational Units"

	// bulkDataIdMap, err := service.getBulkDataDownloadLinks()
	// 1. ///
	// 2. Make REST API call is to get the map of bulk data pointers then we will have access to all of the zip files. This file is JSON and contains IDs required to obtain bulk data files.
	// 3. Download all zip files needed and extract the CSV from within each zip file.
	fmt.Println("ImportCourses...")

	return nil
}

func ReadCSV[T any](values *T, csvFilePath string) {
	coursesFile, err := os.OpenFile(csvFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer coursesFile.Close()
	if err := gocsv.UnmarshalFile(coursesFile, values); err != nil { // Load clients from file
		panic(err)
	}
}

func (srv *BrightspaceService) ImportMilestones(coursePair map[string]interface{}, mappings []map[string]interface{}, db *gorm.DB, lastRun time.Time) error {
	fmt.Println("ImportMilestones...")
	return nil
}

func (srv *BrightspaceService) ImportActivityForCourse(coursePair map[string]interface{}, db *gorm.DB) error {
	fmt.Println("ImportActivityForCourse...")
	return nil
}

func (srv *BrightspaceService) GetJobParams() *map[string]interface{} {
	fmt.Println("GetJobParams...")
	return srv.JobParams
}
