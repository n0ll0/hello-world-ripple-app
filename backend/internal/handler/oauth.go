package handler

import (
	"log"
	"net/http"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
)

type OAuthHandler struct {
	Srv *server.Server
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	err := h.Srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	h.Srv.HandleTokenRequest(w, r)
}

func (h *OAuthHandler) SetErrorHandlers() {
	h.Srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	h.Srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})
}
