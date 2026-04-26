package main

import "sync"

type Job struct {
	Status  string
	IP      string
	LogPath string
}

var jobs sync.Map