package main

import (
	"encoding/json"
	"net/http"

	"github.com/u35s/glog"
)

type PasswordApi struct{}

type Cmd struct {
	Err     string
	Content interface{}
}

type PortPassword struct {
	Port     string
	Password string
	ReqFlow  float64
	RepFlow  float64
}

func (this *PasswordApi) send(w http.ResponseWriter, err string, content interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	var ret Cmd
	ret.Err = err
	ret.Content = content
	json, jerr := json.Marshal(&ret)
	if jerr == nil {
		w.Write(json)
	}
}

func (this *PasswordApi) listPort(w http.ResponseWriter, req *http.Request) {
	var send []PortPassword
	for port, plistener := range srv.pwm.portListener {
		var pp PortPassword
		pp.Port = port
		pp.Password = plistener.password
		pp.ReqFlow, pp.RepFlow = srv.pwm.getFlowByPort(port)
		send = append(send, pp)
	}
	this.send(w, "", &send)
}

func (this *PasswordApi) addPort(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	port, pwd := req.PostFormValue("port"), req.PostFormValue("password")
	if port == "" || pwd == "" {
		this.send(w, "port or password is null", "")
		return
	}
	if _, ok := srv.pwm.get(port); ok {
		this.send(w, "port already run", "")
		return
	}
	go srv.pwm.run(port, pwd, srv.conf.Auth)
	srv.sqld.addUser(port, pwd)
	glog.Inf("[pwdapi],add,port:%v,pwd:%v", port, pwd)
	this.send(w, "", "")
}

func (this *PasswordApi) updatePort(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	port, pwd := req.FormValue("port"), req.FormValue("password")
	if port == "" || pwd == "" {
		this.send(w, "port or password is null", "")
		return
	}
	go srv.pwm.updatePortPasswd(port, pwd, srv.conf.Auth)
	srv.sqld.updateUser(port, pwd)
	glog.Inf("[pwdapi],update,port:%v,pwd:%v", port, pwd)
	this.send(w, "", "")
}

func (this *PasswordApi) delPort(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	port := req.FormValue("port")
	if port == "" {
		this.send(w, "port is null", "")
		return
	}
	go srv.pwm.del(port)
	srv.sqld.delUser(port, "")
	glog.Inf("[pwdapi],del,port:%v", port)
	this.send(w, "", "")
}

func (this *PasswordApi) init() {
	http.HandleFunc("/port/list", this.listPort)
	http.HandleFunc("/port/add", this.addPort)
	http.HandleFunc("/port/update", this.updatePort)
	http.HandleFunc("/port/del", this.delPort)
	go func() {
		http.ListenAndServe(":6001", nil)
	}()
	glog.Inf("[pwdapi],serve listening to port %v", 6001)
}
