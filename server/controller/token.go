package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type TokenRequestBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func (trb TokenRequestBody) validateRequestBody() error {
	if trb.ClientID == "" || trb.ClientSecret == "" || trb.Username == "" || trb.Password == "" {
		return errors.New("missing required paramaters: 'client_id', 'client_secret', 'username', 'password'")
	}
	return nil
}

func (trb TokenRequestBody) GetInfo(db *sqlx.DB) (string, []string, error) {
	roles := make([]string, 0)
	var redirectURL string
	if err := trb.validateRequestBody(); err != nil {
		return redirectURL, roles, err
	}

	var rolesB []byte
	row := db.QueryRow(`
		SELECT redirect_url, array_to_json(u.roles) 
		FROM oauth2_server.applications a 
		LEFT JOIN oauth2_server.users u ON a.id = u.application_id
		WHERE a.client_id = $1 
		AND a.client_secret = crypt($2, a.client_secret)
		AND u.username = $3
		AND u.password = crypt($4, u.password);`,
		trb.ClientID, trb.ClientSecret, trb.Username, trb.Password,
	)
	if err := row.Scan(&redirectURL, &rolesB); err != nil {
		if err == sql.ErrNoRows {
			return redirectURL, roles, fmt.Errorf("invalid credentials for username '%s' and client_id '%s'", trb.Username, trb.ClientID)
		}
		return redirectURL, roles, err
	}
	if err := json.Unmarshal(rolesB, &roles); err != nil {
		return redirectURL, roles, err
	}

	return redirectURL, roles, nil
}

func (crtl *Controller) Token(c echo.Context) error {

	var trb TokenRequestBody
	if err := c.Bind(&trb); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	redirectURL, roles, err := trb.GetInfo(crtl.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// put these into redis?
	token := uuid.New().String()
	refreshToken := uuid.New().String()
	result := map[string]interface{}{
		"redirect_url":  redirectURL,
		"roles":         roles,
		"token":         token,
		"refresh_token": refreshToken,
	}

	return c.JSON(http.StatusOK, result)
}
