package main

//Server结构体
type Server struct {
	ServerName   string `json:"server_name"`
	ServerIP     string `json:"server_ip"`
	ServerOnline bool   `json:"server_online"`
}
