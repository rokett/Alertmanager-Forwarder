package main

import (
	"time"
)

type alertGroup struct {
	Version           string      `json:"version"`
	GroupKey          string      `json:"groupKey"`
	Status            string      `json:"status"`
	Receiver          string      `json:"receiver"`
	GroupLabels       interface{} `json:"groupLabels"`
	CommonLabels      interface{} `json:"commonLabels"`
	CommonAnnotations interface{} `json:"commonAnnotations"`
	ExternalURL       string      `json:"externalURL"`
	Alerts            []alert     `json:"alerts"`
}

type alert struct {
	Status       string      `json:"status"`
	Labels       interface{} `json:"labels"`
	Annotations  interface{} `json:"annotations"`
	StartsAt     time.Time   `json:"startsAt"`
	EndsAt       time.Time   `json:"endsAt"`
	GeneratorURL string      `json:"generatorURL"`
}
