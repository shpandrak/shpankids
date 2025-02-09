package api

import (
	"errors"
	"net/http"
	"shpankids/internal/infra/util"
)

func HandleErrors(w http.ResponseWriter, r *http.Request, err error) {
	var ufe util.UserFacingError
	ok := errors.As(err, &ufe)
	if ok {
		http.Error(w, err.Error(), ufe.HttpRet())
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
