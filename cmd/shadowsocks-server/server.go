package main

import (
	ss "github.com/u35s/shadowsocks"
	"github.com/u35s/shadowsocks/cmd/shadowsocks-server/zlib"
)

var srv *Server

type Server struct {
	conf  *ss.Config
	pwm   PasswdManager
	pwapi PasswordApi
	sqld  SqlDataManager
}

func (this *Server) init() {
	dealArgs()
	this.pwm.init()
	this.pwapi.init()
	this.sqld.init()
}

func (this *Server) run() {
	zlib.WaitSignal()
}

func main() {
	srv = new(Server)
	srv.init()

	srv.run()
}
