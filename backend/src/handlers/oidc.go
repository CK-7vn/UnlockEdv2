package handlers

import (
	"UnlockEdv2/src/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func (srv *Server) registerOidcRoutes() {
	//custom ones now for brightspace development--work in progress
	srv.Mux.Handle("GET /api/oidc/keighan/brightspace", http.HandlerFunc(srv.handleBrightspaceLogin))
	srv.Mux.Handle("/api/oidc/keighan/brightspace/callback", http.HandlerFunc(srv.handleBrightspaceCallback))
	//custom ones now for brightspace development--work in progress
	srv.Mux.Handle("GET /api/oidc/clients", srv.applyAdminMiddleware(srv.handleGetAllClients))
	srv.Mux.Handle("POST /api/oidc/clients", srv.applyAdminMiddleware(srv.handleRegisterClient))
	srv.Mux.Handle("GET /api/oidc/clients/{id}", srv.applyAdminMiddleware(srv.handleGetOidcClient))
}
//START FOR DEV USE ONLY
var oauthStateString = "random"

type BrightspaceToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// threw this here for develpment placement--WIP
var brightspaceConfig = &oauth2.Config{
	ClientID:     "5196e946-52d9-410b-9851-ffa6390f0b09",                    //in DB
	ClientSecret: "-KIJ5WUmermREabxfgVWnQybvlezGgf0Ef730nnL5VE",             //in DB
	RedirectURL:  "https://127.0.0.1/api/oidc/keighan/brightspace/callback", //SYSTEM PROP
	Scopes:       []string{"accommodations:profile:manage,read accountsettings:locale:read,update attributes:schemas:read attributes:users:delete,read,update datahub:dataexports:download,read datasets:bds:read globalusermapping:mapping:write localauthenticationsecurity:overrides:manage,read organizations:image:read role:detail:create,read users:activation:read,update users:own_profile:read users:own_pronoun:read,update users:password:delete,write users:profile:read users:userdata:create,delete,read,update"},
	Endpoint: oauth2.Endpoint{
		AuthURL:   "https://auth.brightspace.com/oauth2/auth",
		TokenURL:  "https://auth.brightspace.com/core/connect/token",
		AuthStyle: oauth2.AuthStyleInHeader,
	},
}

func (srv *Server) handleBrightspaceLogin(w http.ResponseWriter, r *http.Request) {
	url := brightspaceConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (srv *Server) handleBrightspaceCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateString {
		fmt.Println("Invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Exchange the authorization code for an access token
	code := r.FormValue("code")
	token, err := brightspaceConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Could not exchange code: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//gets the access token
	bsToken := BrightspaceToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	fmt.Println("Login successful! Refresh token: ", bsToken.RefreshToken)
	//get all of them and search
	providers, err := srv.Db.GetAllActiveProviderPlatforms()
	fmt.Println("length of providers: ", len(providers))
	if err != nil {
		fmt.Println("failed to get provider platforms from db, error is: ", err)
		fmt.Fprintf(w, "Login successful! Access token: %s", bsToken.AccessToken)
		return
	}
	//attempt to save, just helping out here
	for _, provider := range providers {
		if provider.Type == models.Brightspace {
			provider.AccessKey = brightspaceConfig.ClientSecret + ";" + bsToken.RefreshToken
			fmt.Println("saving: ", provider.AccessKey)
			if err := srv.Db.Save(&provider).Error; err != nil {
				fmt.Println("failed to get provider platforms from db, error is: ", err)
				fmt.Fprintf(w, "Login successful! Access token: %s", bsToken.AccessToken)
				return
			}
			break
		}
	}
	fmt.Fprintf(w, "Login successful! Access token: %s", bsToken.AccessToken)
}

func (srv *Server) handleGetAllClients(w http.ResponseWriter, r *http.Request, log sLog) error {
	clients, err := srv.Db.GetAllRegisteredClients()
	if err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, clients)
}

type RegisterClientRequest struct {
	RedirectURI        string `json:"redirect_uri"`
	ProviderPlatformID uint   `json:"provider_platform_id"`
	AutoRegister       bool   `json:"auto_register"`
}

func clientToResponse(client *models.OidcClient) *models.ClientResponse {
	return &models.ClientResponse{
		ClientID:      client.ClientID,
		ClientSecret:  client.ClientSecret,
		AuthEndpoint:  os.Getenv("APP_URL") + "/oauth2/auth",
		TokenEndpoint: os.Getenv("APP_URL") + "/oauth2/token",
		Scopes:        client.Scopes,
	}
}

func (srv *Server) handleGetOidcClient(w http.ResponseWriter, r *http.Request, log sLog) error {
	id := r.PathValue("id")
	log.add("oidc_id", id)
	client, err := srv.Db.GetOidcClientById(id)
	if err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, *clientToResponse(client))
}

func (srv *Server) handleRegisterClient(w http.ResponseWriter, r *http.Request, log sLog) error {
	request := RegisterClientRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return newJSONReqBodyServiceError(err)
	}
	provider, err := srv.Db.GetProviderPlatformByID(int(request.ProviderPlatformID))
	log.add("providerPlatformId", request.ProviderPlatformID)
	if err != nil {
		return newDatabaseServiceError(err)
	}
	if provider.OidcID != 0 || provider.ExternalAuthProviderId != "" {
		return newBadRequestServiceError(errors.New("client already registered"), "Client already registered")
	}
	client, externalId, err := models.OidcClientFromProvider(provider, request.AutoRegister, srv.Client)
	if err != nil {
		return newInternalServerServiceError(err, err.Error())
	}
	log.add("externalId", externalId)
	if err := srv.Db.RegisterClient(client); err != nil {
		return newDatabaseServiceError(err)
	}
	provider.ExternalAuthProviderId = externalId
	if _, err := srv.Db.UpdateProviderPlatform(provider, provider.ID); err != nil {
		log.add("clientId", client.ID)
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusCreated, *clientToResponse(client))
}
