package main

type Server struct {
	ServerName   string `json:"server_name"`
	ServerIP     string `json:"server_ip"`
	ServerOnline bool   `json:"server_online"`
}
