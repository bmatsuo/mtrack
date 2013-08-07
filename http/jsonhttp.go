// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// jsonhttp.go [created: Mon, 29 Jul 2013]

package http

import (
	"log"
	"net/http"

	"github.com/bmatsuo/mtrack/http/jsonapi"
)

func InternalError(resp http.ResponseWriter, req *http.Request, v ...interface{}) {
	HTTPLog(req, v...)
	jsonapi.Error(resp, 500, "internal error")
	return
}

func Unauthorized(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("WWW-Authenticate", "algorithm=token")
	jsonapi.Error(resp, 401, "unauthorized")
}

func Forbidden(resp http.ResponseWriter, req *http.Request) {
	jsonapi.Error(resp, 403, "forbidden")
}

func BadAuthorization(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("WWW-Authenticate", "algorithm=token")
	jsonapi.Error(resp, 400, "invalid authorization")
}

func NotJson(resp http.ResponseWriter, req *http.Request) {
	jsonapi.Error(resp, 415, "Content-Type is not application/json")
	return
}

func InvalidJson(resp http.ResponseWriter, req *http.Request, err error) {
	log.Printf("%q: %v", req.URL.Path, err)
	jsonapi.Error(resp, 400, "request was not valid json")
	return
}

func MissingParameter(resp http.ResponseWriter, req *http.Request, param string) {
	jsonapi.Error(resp, 400, "missing parameter: ", param)
	return
}

func InvalidParameter(resp http.ResponseWriter, req *http.Request, param string) {
	jsonapi.Error(resp, 400, "invalid parameter: ", param)
	return
}

func NotFound(resp http.ResponseWriter, req *http.Request, v ...interface{}) {
	jsonapi.Error(resp, 404, "not found")
	return
}

