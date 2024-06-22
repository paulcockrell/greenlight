package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)

		return
	}

	// Append new line to JSON. This makes it nicer to view responses in the terminal
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
