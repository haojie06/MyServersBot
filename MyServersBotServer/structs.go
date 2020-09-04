package main

import "time"

type Server struct {
	ServerName   string    `json:"server_name"`
	ServerIP     string    `json:"server_ip"`
	ServerOnline bool      `json:"server_online"`
	LastActive   time.Time `json:"last_active"`
}

type Setting struct {
	Password string `json:"password"`
}
