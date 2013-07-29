// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http.go [created: Sat, 27 Jul 2013]

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/bmatsuo/mtrack/jsonapi"
	"github.com/bmatsuo/mtrack/model"
	"github.com/gorilla/mux"
)

func InternalError(resp http.ResponseWriter, req *http.Request, v ...interface{}) {
	HTTPLog(req, v...)
	jsonapi.Error(resp, 500, "internal error")
	return
}

func NotJson(resp http.ResponseWriter, req *http.Request) {
	jsonapi.Error(resp, 415, "Content-Type is not application/json")
	return
}

func NotFound(resp http.ResponseWriter, req *http.Request, v ...interface{}) {
	jsonapi.Error(resp, 404, "not found")
	return
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
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "request was not valid json")
		return
	}

	mediaid, err := params.Get("mediaId").String()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "invalid mediaId")
		return
	}

	userid, err := params.Get("userId").String()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "invalid userId")
		return
	}

	row := model.DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserStartedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	if count > 0 {
		jsonapi.Error(resp, 400, "already started")
		return
	}

	row = model.DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserFinishedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	err = row.Scan(&count)
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	tx, err := model.DB.Begin()
	if err != nil {
		InternalError(resp, req, err)
		return
	}
	if count > 0 {
		q := `DELETE FROM UserFinishedMedia WHERE MediaId = ? AND UserId = ?`
		_, err := tx.Exec(q, mediaid, userid)
		if err != nil {
			InternalError(resp, req, "couldn't remove finished:", err)
			err := tx.Rollback()
			if err != nil {
				HTTPLog(req, "couldn't rollback transaction:", err)
			}
			return
		}
	}

	q := `INSERT INTO UserStartedMedia(MediaId, UserId) Values(?, ?)`
	_, err = tx.Exec(q, mediaid, userid)
	if err != nil {
		InternalError(resp, req, "couldn't remove finished:", err)
		err := tx.Rollback()
		if err != nil {
			HTTPLog(req, "couldn't rollback transaction:", err)
		}
		return
	}

	err = tx.Commit()
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

	mediaid, err := params.Get("mediaId").String()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "invalid mediaId")
		return
	}

	userid, err := params.Get("userId").String()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 400, "invalid userId")
		return
	}

	row := model.DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserFinishedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	var count int
	err = row.Scan(&count)
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	if count > 0 {
		jsonapi.Error(resp, 400, "already finished")
		return
	}

	row = model.DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserStartedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	err = row.Scan(&count)
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	tx, err := model.DB.Begin()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}
	if count > 0 {
		q := `DELETE FROM UserStartedMedia WHERE MediaId = ? AND UserId = ?`
		_, err := tx.Exec(q, mediaid, userid)
		if err != nil {
			log.Printf("%q: couldn't remove started: %v", req.URL.Path, err)
			err := tx.Rollback()
			if err != nil {
				log.Printf("%q: couldn't rollback transaction: %v",
					req.URL.Path, err)
			}
			jsonapi.Error(resp, 500, "internal error")
			return
		}
	}

	q := `INSERT INTO UserFinishedMedia(MediaId, UserId) Values(?, ?)`
	_, err = tx.Exec(q, mediaid, userid)
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		err := tx.Rollback()
		if err != nil {
			log.Printf("%q: couldn't rollback transaction: %v",
				req.URL.Path, err)
		}
		jsonapi.Error(resp, 500, "internal error")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("%q: %v", req.URL.Path, err)
		jsonapi.Error(resp, 500, "internal error")
		return
	}

	jsonapi.Success(resp, nil)
}

func MediaIndex(resp http.ResponseWriter, req *http.Request) {
	results := make([]interface{}, 0, 20)
	rows, err := model.DB.Query(`
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
			"mediaId":  mediaid,
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
	rows, err := model.DB.Query(`
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
			"mediaId": mediaid,
			"userId":  userid,
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
	rows, err := model.DB.Query(`
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
			"mediaId":  mediaid,
			"userId":   userid,
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
