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
		Expires: createdSessionEntity.AccessExpiresAt, /*TODO*/
		Secure:  false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   createdSessionEntity.JWTRefresh,
		Path:    "/",
		Expires: createdSessionEntity.RefreshExpiresAt, /*TODO*/
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

func (h *SessionHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		/*Handle*/
		fmt.Println("body err")
		return
	}

	loginRequest := &dto.LoginRequest{}
	err = json.Unmarshal(body, loginRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("unmarshal err")
		return
	}

	createdSessionEntity, err := h.sessionUC.Login(loginRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("login err", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    createdSessionEntity.JWTAccess,
		Path:     "/",
		Expires:  createdSessionEntity.AccessExpiresAt,
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    createdSessionEntity.JWTRefresh,
		Path:     createdSessionEntity.JWTRefresh,
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(int(createdSessionEntity.UserID)),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userIDString := r.Header.Get("X-User-ID")
	userID64, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	userID := uint(userID64)

	err = h.sessionUC.Delete(userID)
	if err != nil {
		/*Handle*/
		fmt.Println("uc logout err", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
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

func (h *SessionHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Refresh token missing")
		return
	}

	tokenKey := []byte(os.Getenv("TOKEN_KEY"))
	token, err := jwt.Parse(refreshCookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return tokenKey, nil
	})
	if err != nil || !token.Valid {
		/*Handle*/
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Invalid refresh token")
		return
	}

	refreshedSession, err := h.sessionUC.RefreshSession(token)
	if err != nil {
		/*Handle*/
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   refreshedSession.JWTAccess,
		Path:    "/",
		Expires: refreshedSession.AccessExpiresAt,
		Secure:  false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshedSession.JWTRefresh,
		Path:    "/",
		Expires: refreshedSession.RefreshExpiresAt,
		Secure:  false,
	})

	w.WriteHeader(http.StatusOK)
}
