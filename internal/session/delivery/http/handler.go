package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lightlink/auth-service/internal/session/domain/dto"
	"github.com/lightlink/auth-service/internal/session/usecase"
)

type SessionHandler struct {
	sessionUC usecase.SessionUsecaseI
}

func NewSessionHandler(sessionUsecase usecase.SessionUsecaseI) *SessionHandler {
	return &SessionHandler{
		sessionUC: sessionUsecase,
	}
}

func (h *SessionHandler) Signup(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		/*Handle*/
		fmt.Println("body err")
		return
	}

	signupRequest := &dto.SignupRequest{}
	err = json.Unmarshal(body, signupRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("unmarshal err")
		return
	}

	createdSessionEntity, err := h.sessionUC.Signup(signupRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("session create err", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   createdSessionEntity.JWTAccess,
		Path:    "/",
		Expires: time.Now().Add(15 * time.Minute), /*TODO*/
		Secure:  false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   createdSessionEntity.JWTRefresh,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour), /*TODO*/
		Secure:  false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "user_id",
		Value:  strconv.Itoa(int(createdSessionEntity.UserID)),
		Path:   "/",
		Secure: false,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) Check(w http.ResponseWriter, r *http.Request) {
	tokenKey := []byte(os.Getenv("TOKEN_KEY"))

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Missing token")
		return
	}

	fieldParts := strings.Split(tokenString, " ")
	if len(fieldParts) != 2 || fieldParts[0] != "Bearer" {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Bad token size")
		return
	}
	pureToken := fieldParts[1]

	token, err := jwt.Parse(pureToken, func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, errors.New("bad sign method")
		}
		return tokenKey, nil
	})
	if err != nil || !token.Valid {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Token is invalid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("User claims type cast err")
		return
	}

	claimsUser, ok := claims["user"].(map[string]interface{})
	if !ok {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("User claims are missing")
		return
	}

	userIDString, ok := claimsUser["id"].(string)
	if !ok {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Couldn't parse user id")
		return
	}

	w.Header().Set("X-User-ID", userIDString)
	w.WriteHeader(http.StatusOK)
}
