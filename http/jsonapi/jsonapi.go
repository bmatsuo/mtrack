// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// jsonapi.go [created: Sun, 28 Jul 2013]

// Package jsonapi does ....
package jsonapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/bitly/go-simplejson"
)

var ErrNotJson = fmt.Errorf("request is not JSON")

type Map map[string]interface{}

func Read(req *http.Request) (*simplejson.Json, error) {
	contentType := req.Header.Get("Content-Type")
	contentType = strings.SplitN(contentType, ";", 2)[0]
	contentType = strings.TrimFunc(contentType, unicode.IsSpace)
	if contentType == "" || contentType == "application/json" {
		dec := json.NewDecoder(req.Body)
		defer req.Body.Close()
		js := new(simplejson.Json)
		err := dec.Decode(js)
		if err != nil {
			return nil, err
		}
		return js, err
	} else {
		return nil, ErrNotJson
	}
}

func Success(resp http.ResponseWriter, result map[string]interface{}) error {
	fullresult := make(map[string]interface{}, len(result)+1)
	for k, v := range result {
		fullresult[k] = v
	}
	fullresult["status"] = "success"

	p, err := json.Marshal(fullresult)
	if err != nil {
		return err
	}

	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = resp.Write(p)
	return err
}

func Error(resp http.ResponseWriter, code int, v ...interface{}) error {
	p, err := json.Marshal(map[string]interface{}{
		"status": "failure",
		"reason": fmt.Sprint(v...),
	})
	if err != nil {
		return err
	}

	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(code)
	_, err = resp.Write(p)
	return err
}
