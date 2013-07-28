// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http.go [created: Sat, 27 Jul 2013]

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/bmatsuo/mtrack/jsonapi"
)

var HTTPConfig struct {
	Addr string
}

func HTTPStart() error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(http.NotFound)
	router.Methods("GET").Path("/media").HandlerFunc(MediaIndex)
	//router.Methods("POST").Path("/start").HandlerFunc(Start)
	router.Methods("GET").Path("/in_progress").HandlerFunc(InProgressIndex)
	//router.Methods("POST").Path("/finish").HandlerFunc(Finish)
	router.Methods("GET").Path("/finished").HandlerFunc(FinishedIndex)

	log.Printf("Serving HTTP on at %v", HTTPConfig.Addr)
	return http.ListenAndServe(HTTPConfig.Addr, router)
}

func MediaIndex(resp http.ResponseWriter, req *http.Request) {
	results := make([]interface{}, 0, 20)
	rows, err := DB.Query(`
		SELECT MediaId, Path, ModTime
		FROM Media
		ORDER BY ModTime DESC
	`)
	if err != nil {
		log.Printf("%q: database error: %v", err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	for rows.Next() {
		var mediaid string
		var path string
		var modtime time.Time
		err = rows.Scan(&mediaid, &path, &modtime)
		if err != nil {
			cols, _ := rows.Columns()
			log.Printf("%q: error scanning row %v: %v",
				req.URL.Path, cols, err)
			continue
		}
		results = append(results, map[string]interface{}{
			"id":       mediaid,
			"path":     path,
			"modified": modtime,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q: result iteration error: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}

	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}

func InProgressIndex(resp http.ResponseWriter, req *http.Request) {
	results := make([]interface{}, 0, 20)
	rows, err := DB.Query(`
		SELECT S.MediaId, S.UserId, S.Started
		FROM UserStartedMedia as S
		ORDER BY Started DESC
	`)
	if err != nil {
		log.Printf("%q: database error: %v", err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	for rows.Next() {
		var mediaid string
		var userid string
		var started time.Time
		err = rows.Scan(&mediaid, &userid, &started)
		if err != nil {
			cols, _ := rows.Columns()
			log.Printf("%q: error scanning row %v: %v",
				req.URL.Path, cols, err)
			continue
		}
		results = append(results, map[string]interface{}{
			"id":      mediaid,
			"user_id": userid,
			"started": started,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q: result iteration error: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}

	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}

func FinishedIndex(resp http.ResponseWriter, req *http.Request) {
	results := make([]interface{}, 0, 20)
	rows, err := DB.Query(`
		SELECT MediaId, UserId, Finished
		FROM UserFinishedMedia
		ORDER BY Finished DESC
	`)
	if err != nil {
		log.Printf("%q: database error: %v", err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	for rows.Next() {
		var mediaid string
		var userid string
		var finished time.Time
		err = rows.Scan(&mediaid, &userid, &finished)
		if err != nil {
			cols, _ := rows.Columns()
			log.Printf("%q: error scanning row %v: %v",
				req.URL.Path, cols, err)
			continue
		}
		results = append(results, map[string]interface{}{
			"id":       mediaid,
			"user_id":  userid,
			"finished": finished,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q: result iteration error: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}

	jsonapi.Success(resp, jsonapi.Map{
		"results": results,
	})
}
