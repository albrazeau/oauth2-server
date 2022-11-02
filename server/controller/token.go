package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

var tokenExpirationMinutes int

func init() {
	var err error
	tokenExpirationMinutes, err = strconv.Atoi(os.Getenv("TOKEN_TIMEOUT_MINUTES"))
	if err != nil {
		log.Fatal(err)
	}
}

type TokenRequestBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type Claims struct {
	UID     string   `json:"uid"`
	Refresh string   `json:"refresh"`
	UserID  int      `json:"user_id"`
	Roles   []string `json:"roles"`
	AppID   int      `json:"app_id"`
	Expires int      `json:"exp"`
	jwt.StandardClaims
}

func (trb TokenRequestBody) validateRequestBody() error {
	if trb.ClientID == "" || trb.ClientSecret == "" || trb.Username == "" || trb.Password == "" {
		return errors.New("missing required paramaters: 'client_id', 'client_secret', 'username', 'password'")
	}
	return nil
}

func (trb TokenRequestBody) GetInfo(db *sqlx.DB) (string, Claims, error) {
	roles := make([]string, 0)
	var redirectURL string
	var claims Claims
	if err := trb.validateRequestBody(); err != nil {
		return redirectURL, claims, err
	}

	var rolesB []byte
	row := db.QueryRow(`
		SELECT redirect_url, array_to_json(u.roles), u.id, u.application_id
		FROM oauth2_server.applications a 
		LEFT JOIN oauth2_server.users u ON a.id = u.application_id
		WHERE a.client_id = $1 
		AND a.client_secret = crypt($2, a.client_secret)
		AND u.username = $3
		AND u.password = crypt($4, u.password);`,
		trb.ClientID, trb.ClientSecret, trb.Username, trb.Password,
	)
	if err := row.Scan(&redirectURL, &rolesB, &claims.UserID, &claims.AppID); err != nil {
		if err == sql.ErrNoRows {
			return redirectURL, claims, fmt.Errorf("invalid credentials for username '%s' and client_id '%s'", trb.Username, trb.ClientID)
		}
		return redirectURL, claims, err
	}
	if err := json.Unmarshal(rolesB, &roles); err != nil {
		return redirectURL, claims, err
	}

	claims.Roles = roles

	return redirectURL, claims, nil
}

func (crtl *Controller) Token(c echo.Context) error {

	var trb TokenRequestBody
	if err := c.Bind(&trb); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	redirectURL, claims, err := trb.GetInfo(crtl.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	claims.UID = uuid.New().String()
	claims.Refresh = uuid.New().String()
	claims.Expires = int(time.Now().Add(time.Duration(tokenExpirationMinutes) * time.Minute).Unix())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if err := crtl.InsertTokenIntoDb(claims); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"token": tokenStr, "refresh_token": claims.Refresh, "redirect_url": redirectURL})
}

func (crtl *Controller) InsertTokenIntoDb(claims Claims) error {

	_, err := crtl.db.Exec(`INSERT INTO oauth2_server.tokens (uid, refresh_token, expires_at, user_id, application_id)
					   VALUES ($1, $2, $3, $4, $5);`, claims.UID, claims.Refresh, claims.Expires, claims.UserID, claims.AppID)

	return err
}

func (crtl *Controller) Validate(c echo.Context) error {

	tokenHeader := c.Request().Header["Authorization"]
	if len(tokenHeader) != 1 {
		return c.JSON(http.StatusBadRequest, "expected authorization header containing one bearer token")
	}
	if !strings.Contains(tokenHeader[0], "Bearer ") {
		return c.JSON(http.StatusBadRequest, "unexpected authorization header, expected bearer token")
	}

	tokenStr := strings.ReplaceAll(tokenHeader[0], "Bearer ", "")

	claims := new(Claims)
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return []byte(os.Getenv("JWT_SECRET")), nil })
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if !tkn.Valid {
		return c.JSON(http.StatusOK, map[string]interface{}{"valid": false, "reason": "invalid token"})
	}
	if time.Now().After(time.Unix(int64(claims.Expires), 0)) {
		return c.JSON(http.StatusOK, map[string]interface{}{"valid": false, "reason": "token expired"})
	}
	if !crtl.ValidateClaimsAgainstDb(claims) {
		return c.JSON(http.StatusOK, map[string]interface{}{"valid": false, "reason": "invalid token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"valid": true})
}

func (crtl *Controller) ValidateClaimsAgainstDb(claims *Claims) bool {

	row := crtl.db.QueryRow(`SELECT 1 FROM oauth2_server.tokens WHERE
	 						 uid = $1
							 AND refresh_token = $2
							 AND expires_at = $3
							 AND user_id = $4
							 AND application_id = $5;`,
		claims.UID, claims.Refresh, claims.Expires, claims.UserID, claims.AppID)

	var valid int
	err := row.Scan(&valid)
	if err != nil {
		return false
	}

	return valid == 1
}

func (crtl *Controller) Refresh(c echo.Context) error {

	var refreshTknBody struct {
		Refresh string `json:"refresh_token"`
	}
	if err := c.Bind(&refreshTknBody); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusNotImplemented, "not implemented")

}
