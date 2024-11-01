package main

import (
	"UnlockEdv2/src/models"
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type DataSetPlugin struct {
	PluginId     string  `json:"PluginId"`
	Name         string  `json:"Name"`
	Description  string  `json:"Description"`
	CreatedDate  string  `json:"CreatedDate"`
	DownloadLink string  `json:"DownloadLink"`
	DownloadSize float64 `json:"DownloadSize"`
}

type BrightspaceUser struct {
	UserId        string `csv:"UserId"`
	UserName      string `csv:"UserName"`
	OrgDefinedId  string `csv:"OrgDefinedId"`
	FirstName     string `csv:"FirstName"`
	LastName      string `csv:"LastName"`
	IsActive      string `csv:"IsActive"`
	Organization  string `csv:"Organization"`
	ExternalEmail string `csv:"ExternalEmail"`
}

type BrightspaceCourse struct {
	OrgUnitId     string `csv:"OrgUnitId"`
	Organization  string `csv:"Organization"`
	Type          string `csv:"Type"`
	Name          string `csv:"Name"`
	Code          string `csv:"Code"`
	IsActive      string `csv:"IsActive"`
	IsDeleted     string `csv:"IsDeleted"`
	OrgUnitTypeId string `csv:"OrgUnitTypeId"`
}

type BrightspaceEnrollment struct {
	OrgUnitId      string `csv:"OrgUnitId"`
	UserId         string `csv:"UserId"`
	RoleName       string `csv:"RoleName"`
	EnrollmentType string `csv:"EnrollmentType"`
}

func (srv *BrightspaceService) IntoImportUser(bsUser BrightspaceUser) *models.ImportUser {
	return nil
}

func (srv *BrightspaceService) IntoCourse(bsCourse BrightspaceCourse) *models.Course {
	//step 1 - send
	id := bsCourse.OrgUnitId
	courseImageUrl := fmt.Sprintf(srv.BaseURL+"/d2l/api/lp/1.28/courses/%s/image", id)
	response, err := srv.SendRequest(courseImageUrl)
	if err != nil {
		return nil
	}
	if bsCourse.OrgUnitId == "6684" {
		fmt.Println("stop here")
	}
	defer response.Body.Close()
	var imgPath string
	if response.StatusCode == http.StatusOK {
		imgBytes, err := io.ReadAll(response.Body)
		if err == nil {
			//check error here:   errors.New("no image data available or decoding failed")
			imgPath, err = UploadBrightspaceImage(imgBytes, id)
			if err != nil {
				fmt.Println(" no images ")
			}
		} else {
			imgPath = ""
		}
	}

	externalUrl, err := url.JoinPath(srv.BaseURL, "")
	course := models.Course{
		ProviderPlatformID:      srv.ProviderPlatformID,
		ExternalID:              bsCourse.OrgUnitId,
		Name:                    bsCourse.Name,
		Type:                    "open_content",
		OutcomeTypes:            "completion",
		TotalProgressMilestones: uint(0), //come back to this one
		Description:             "Brightspace Course: " + bsCourse.Name,
		ThumbnailURL:            imgPath,
		ExternalURL:             externalUrl,
	}

	fmt.Println("image path: ", imgPath)
	return &course
}

func UploadBrightspaceImage(imgBytes []byte, bsCourseId string) (string, error) {
	filename := "image_brightspace" + "/" + bsCourseId + ".jpg"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err = part.Write(imgBytes); err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("POST", os.Getenv("APP_URL")+"/api/upload", body)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Content-Length", fmt.Sprintf("%d", len(body.Bytes())))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned non-OK status: %s", response.Status)
	}
	urlRes := struct {
		Data struct {
			Url string `json:"url"`
		}
		Message string `json:"message"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&urlRes)
	if err != nil {
		return "", err
	}
	// fmt.Println("url with lowercase: ", urlRes.data.Url)
	fmt.Println("url with uppercase: ", urlRes.Data.Url)
	return urlRes.Data.Url, nil
}

func (srv *BrightspaceService) getPluginId(pluginName string) (string, error) {
	var pluginId string
	resp, err := srv.SendRequest(DataSetsEndpoint)
	if err != nil {
		return pluginId, err
	}
	defer resp.Body.Close()
	pluginData := []DataSetPlugin{}
	if err = json.NewDecoder(resp.Body).Decode(&pluginData); err != nil {
		return pluginId, err
	}
	for _, plugin := range pluginData {
		if plugin.Name == pluginName {
			pluginId = plugin.PluginId
			break
		} //end if
	}
	return pluginId, nil
}

func (srv *BrightspaceService) downloadAndUnzipFile(targetDirectory string, targetFileName string, endpointUrl string) (string, error) {
	var destPath string
	resp, err := srv.SendRequest(endpointUrl)
	if err != nil {
		return destPath, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		zipFilePath := filepath.Join(targetDirectory, targetFileName)
		file, err := os.Create(zipFilePath)
		if err != nil {
			return destPath, err
		}
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return destPath, err
		}
		file.Close()
		zipFile, err := zip.OpenReader(zipFilePath) //open the zip file
		if err != nil {
			return destPath, err
		}
		defer zipFile.Close() //close it later
		for _, zippedFile := range zipFile.File {
			destPath = filepath.Join(targetDirectory, zippedFile.Name)
			if zippedFile.FileInfo().IsDir() {
				if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
					fmt.Println("error occurred while trying to make directories, error is: ", err)
				}
				continue
			}
			//there is going to be no directory for this file, as these are csv files
			if err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return destPath, err
			}
			outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zippedFile.Mode())
			if err != nil {
				return destPath, err
			}
			defer outFile.Close()
			rc, err := zippedFile.Open()
			if err != nil {
				return destPath, err
			}
			defer rc.Close()
			_, err = io.Copy(outFile, rc)
			if err != nil {
				return destPath, err
			}
		}
	}
	return destPath, err
}
