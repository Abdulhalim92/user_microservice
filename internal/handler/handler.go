package handler

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"user/internal/logging"
	"user/internal/model"
	"user/internal/service"
)

type Handler struct {
	Nats    *nats.Conn
	Logger  *logging.Logger
	Service *service.Service
}

func NewHandler(nats *nats.Conn, log *logging.Logger, service *service.Service) *Handler {
	return &Handler{
		Nats:    nats,
		Logger:  log,
		Service: service,
	}
}

func (h *Handler) Init() {
	sub, err := h.Nats.Subscribe("user.sign-up", h.SignUp)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	sub, err = h.Nats.Subscribe("user.sign-in", h.SignIn)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	sub, err = h.Nats.Subscribe("user.refresh", h.Refresh)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	sub, err = h.Nats.Subscribe("user.sign-out", h.SignOut)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	sub, err = h.Nats.Subscribe("user.token-valid", h.TokenValid)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	defer sub.Unsubscribe()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}

func (h *Handler) SignUp(msg *nats.Msg) {

	var u *model.User

	fmt.Println(string(msg.Data))

	err := json.Unmarshal(msg.Data, &u)
	if err != nil {
		h.Logger.Errorf("cannot unmarshal message: %s", err.Error())
		h.Nats.Publish(msg.Reply, []byte(fmt.Sprintf("cannot unmarshal message: %s", err.Error())))
		return
	}

	existsUser, err := h.Service.ExistsUser(u.Name)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}
	if existsUser {
		h.Logger.Println("such user exists")
		h.Nats.Publish(msg.Reply, []byte("such user exists"))
		return
	}

	userID, err := h.Service.CreateUser(u)
	if err != nil {
		h.Logger.Println(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	userId := strconv.Itoa(userID)

	h.Nats.Publish(msg.Reply, []byte(userId))
}

func (h *Handler) SignIn(msg *nats.Msg) {

	var u *model.User

	fmt.Println(string(msg.Data))

	err := json.Unmarshal(msg.Data, &u)
	if err != nil {
		h.Logger.Errorf("cannot unmarshal message: %s", err.Error())
		h.Nats.Publish(msg.Reply, []byte(fmt.Sprintf("cannot unmarshal message: %s", err.Error())))
		return
	}

	userID, err := h.Service.GetUser(u)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	td, err := h.Service.CreateToken(userID)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	err = h.Service.CreateAuth(userID, td)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	tokens := map[string]string{
		"access_token":  td.AccessToken,
		"refresh_token": td.RefreshToken,
	}

	tokensBytes, err := json.Marshal(tokens)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	h.Nats.Publish(msg.Reply, tokensBytes)
}

func (h *Handler) Refresh(msg *nats.Msg) {

	mapToken := map[string]string{}

	err := json.Unmarshal(msg.Data, &mapToken)
	if err != nil {
		h.Logger.Errorf("cannot unmarshal message: %s", err.Error())
		h.Nats.Publish(msg.Reply, []byte(fmt.Sprintf("cannot unmarshal message: %s", err.Error())))
		return
	}

	refreshToken := mapToken["refresh_token"]

	// verify the token
	err = os.Setenv("REFRESH_SECRET", "secret")
	if err != nil {
		h.Logger.Error(err)
		return
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("expired refresh token"))
		return
	}

	// is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		h.Logger.Error("invalid token")
		h.Nats.Publish(msg.Reply, []byte("invalid token"))
		return
	}

	// since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}
		//userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		//if err != nil {
		//	c.JSON(http.StatusUnprocessableEntity, "Error occurred")
		//	return
		//}

		userID := claims["user_id"].(int)

		// delete the previous Refresh Token
		deleted, delErr := h.Service.DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}

		// create new pairs of refresh and access tokens
		ts, createErr := h.Service.CreateToken(userID)
		if createErr != nil {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}

		tokensBytes, err := json.Marshal(tokens)
		if delErr != nil {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}

		h.Nats.Publish(msg.Reply, tokensBytes)

	} else {
		h.Nats.Publish(msg.Reply, []byte("expired token"))
	}

}

func (h *Handler) SignOut(msg *nats.Msg) {

	// extract token
	bearToken := string(msg.Data)

	err := os.Setenv("ACCESS_SECRET", "secret")
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	// verify token
	jwtToken, err := jwt.Parse(bearToken, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method confirm to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.Logger.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte(err.Error()))
		return
	}

	// extract token metadata

	var accessDetails *model.AccessDetails

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok && jwtToken.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}

		userID, ok := claims["user_id"].(int)
		if !ok {
			h.Logger.Error(err)
			h.Nats.Publish(msg.Reply, []byte("internal server error"))
			return
		}

		accessDetails.AccessUuid = accessUuid
		accessDetails.UserId = userID
	}

	deleted, err := h.Service.DeleteAuth(accessDetails.AccessUuid)
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}
	if err != nil || deleted == 0 {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	h.Nats.Publish(msg.Reply, []byte("Successfully logged out"))
}

func (h *Handler) TokenValid(msg *nats.Msg) {

	// extract token
	bearToken := string(msg.Data)

	err := os.Setenv("ACCESS_SECRET", "secret")
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte("internal server error"))
		return
	}

	// verify token
	jwtToken, err := jwt.Parse(bearToken, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method confirm to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.Logger.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		h.Logger.Error(err)
		h.Nats.Publish(msg.Reply, []byte(err.Error()))
		return
	}

	// validation token
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok && !!jwtToken.Valid {
		h.Logger.Println("token is not valid")
		h.Nats.Publish(msg.Reply, []byte("token is not valid"))
		return
	}

	userId := claims["user_id"].(int)
	userID := strconv.Itoa(userId)

	h.Nats.Publish(msg.Reply, []byte(userID))
}
