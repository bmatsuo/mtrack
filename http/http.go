// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http.go [created: Sat, 27 Jul 2013]

package http

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/bmatsuo/mtrack/jsonapi"
	"github.com/bmatsuo/mtrack/model"
	"github.com/gorilla/mux"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

func AuthorizeUser(req *http.Request) (*model.User, error) {
	auth := req.Header.Get("Authorization")
	if len(auth) == 0 {
		return nil, ErrUnauthorized
	}
	auth = strings.Trim(auth, " ")
	pieces := strings.Fields(auth)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("invalid authorization")
	}
	authType, token := pieces[0], pieces[1]
	if strings.ToLower(authType) != "token" {
		return nil, fmt.Errorf("invalid authorization type")
	}
	return model.FindUserByAccessToken(token)
}

type MissingParameterError string
type InvalidParameterError string

func (err MissingParameterError) Error() string {
	var name = string(err)
	if name == "" {
		name = "$root"
	}
	return fmt.Sprintf("missing parameter: %v", string(err))
}

func (err InvalidParameterError) Error() string {
	var name = string(err)
	if name == "" {
		name = "$root"
	}
	return fmt.Sprintf("invalid parameter: %v", string(err))
}

func StringParameter(js *simplejson.Json, path ...string) (string, error) {
	// simplejson's api makes this kind of misleading in some situations
	present := false
	if len(path) > 1 {
		js = js.GetPath(path[:len(path)-1]...)
		js, present = js.CheckGet(path[len(path)-1])
		if !present {
			return "", MissingParameterError(strings.Join(path, "."))
		}
	} else {
		js, present = js.CheckGet(path[0])
		if !present {
			return "", MissingParameterError(strings.Join(path, "."))
		}
	}
	str, err := js.String()
	if err != nil {
		return "", InvalidParameterError(strings.Join(path, "."))
	}
	return str, nil
}

func HTTPLog(req *http.Request, v ...interface{}) {
	log.Printf("%q: %v", req.URL.Path, fmt.Sprint(v...))
}

var HTTPConfig struct {
	Addr string
}

func HTTPStart() error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(http.NotFound)
	router.Methods("POST").Path("/open").HandlerFunc(Open)
	router.Methods("GET").Path("/media").HandlerFunc(MediaIndex)
	router.Methods("POST").Path("/start").HandlerFunc(Start)
	router.Methods("GET").Path("/in_progress").HandlerFunc(InProgressIndex)
	router.Methods("POST").Path("/finish").HandlerFunc(Finish)
	router.Methods("GET").Path("/finished").HandlerFunc(FinishedIndex)

	log.Printf("Serving HTTP on at %v", HTTPConfig.Addr)
	return http.ListenAndServe(HTTPConfig.Addr, router)
}

func Open(resp http.ResponseWriter, req *http.Request) {
	params, err := jsonapi.Read(req)
	if err == jsonapi.ErrNotJson {
		NotJson(resp, req)
		return
	}
	mediaid, err := params.Get("mediaId").String()
	if err != nil {
		jsonapi.Error(resp, 400, "invalid mediaId")
		return
	}
	var path string
	row := model.DB.QueryRow(`SELECT Path FROM Media WHERE MediaId = ?`, mediaid)
	err = row.Scan(&path)
	if err != sql.ErrNoRows {
		jsonapi.Error(resp, 404, "not found")
		return
	}
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	err = exec.Command("open", "http://google.com").Run()
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	jsonapi.Success(resp, nil)
}

func Start(resp http.ResponseWriter, req *http.Request) {
	params, err := jsonapi.Read(req)
	if err == jsonapi.ErrNotJson {
		NotJson(resp, req)
		return
	}
	if err != nil {
		InvalidJson(resp, req, err)
		return
	}

	mediaid, err := StringParameter(params, "mediaId")
	switch err.(type) {
	case MissingParameterError:
		MissingParameter(resp, req, "mediaId")
		return
	case InvalidParameterError:
		InvalidParameter(resp, req, "mediaId")
		return
	}

	userid, err := StringParameter(params, "userId")
	switch err.(type) {
	case MissingParameterError:
		MissingParameter(resp, req, "userId")
		return
	case InvalidParameterError:
		InvalidParameter(resp, req, "userId")
		return
	}

	user, err := AuthorizeUser(req)
	if err == ErrUnauthorized {
		Unauthorized(resp, req)
		return
	}
	if err != nil {
		BadAuthorization(resp, req)
		return
	}
	if user.Id != userid { // SELF_PROGRESS_UPDATE should be a permission?
		ok, err := model.UserHasPermission(user.Id, model.PermUserProgressUpdate)
		if err != nil {
			InternalError(resp, req, err)
			return
		}
		if !ok {
			Forbidden(resp, req)
			return
		}
	}

	err = model.StartMedia(userid, mediaid)
	if err == model.ErrAlreadyStarted {
		jsonapi.Error(resp, 400, err)
		return
	}
	if err != nil {
		InternalError(resp, req, err)
		return
	}

	jsonapi.Success(resp, nil)
}

func Finish(resp http.ResponseWriter, req *http.Request) {
	params, err := jsonapi.Read(req)
	if err == jsonapi.ErrNotJson {
		NotJson(resp, req)
		return
	}
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "request was not valid json")
		return
	}

	mediaid, err := StringParameter(params, "mediaId")
	switch err.(type) {
	case MissingParameterError:
		MissingParameter(resp, req, "mediaId")
		return
	case InvalidParameterError:
		InvalidParameter(resp, req, "mediaId")
		return
	}

	userid, err := StringParameter(params, "userId")
	switch err.(type) {
	case MissingParameterError:
		MissingParameter(resp, req, "userId")
		return
	case InvalidParameterError:
		InvalidParameter(resp, req, "userId")
		return
	}

	user, err := AuthorizeUser(req)
	if err == ErrUnauthorized {
		Unauthorized(resp, req)
		return
	}
	if err != nil {
		BadAuthorization(resp, req)
		return
	}
	if user.Id != userid { // SELF_PROGRESS_UPDATE should be a permission?
		ok, err := model.UserHasPermission(user.Id, model.PermUserProgressUpdate)
		if err != nil {
			InternalError(resp, req, err)
			return
		}
		if !ok {
			Forbidden(resp, req)
			return
		}
	}

	err = model.FinishMedia(userid, mediaid)
	if err == model.ErrAlreadyFinished {
		jsonapi.Error(resp, 400, err)
		return
	}
	if err != nil {
		InternalError(resp, req, err)
		return
	}

	jsonapi.Success(resp, nil)
}

func MediaIndex(resp http.ResponseWriter, req *http.Request) {
	results, err := model.AllMedia()
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}

func InProgressIndex(resp http.ResponseWriter, req *http.Request) {
	results, err := model.AllInProgress()
	if err != nil {
		InternalError(resp, req, err)
		return
	}

	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}

func FinishedIndex(resp http.ResponseWriter, req *http.Request) {
	results, err := model.AllFinished()
	if err != nil {
		InternalError(resp, req, err)
		return
	}

	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}
