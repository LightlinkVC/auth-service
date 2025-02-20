package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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
