package main

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Server struct {
	ServerName   string    `json:"server_name"`
	ServerIP     string    `json:"server_ip"`
	ServerDescription string `json:"server_description"`
	ServerLocation string `json:"server_location"`
	ServerOnline bool      `json:"server_online"`
	LastActive   time.Time `json:"last_active"`
}

type Setting struct {
	Password string `json:"password"`
}

//保存每个用户的会话，存在一个id——Conversation map里面
type Conversation struct {
	//历史对话map，用于删除过时消息
	HistoryMsg map[string]*tb.Message
	//当前对话进度，用于填写表单等持续回话，配合switch case使用
	CurConversation string
	//当前对话添加服务器表单的阶段
	CurAddServerStep string
	//权限 visitor admin
	Permission string
	//一个表单，用于记录将要添加的服务器
	AddServer *AddServerForm
}

//要添加到数据库中的服务器
type AddServerForm struct {
	ServerName        string `json:"server_name"`
	ServerLocation    string `json:"server_location"`
	ServerDescription string `json:"server_description"`
}
