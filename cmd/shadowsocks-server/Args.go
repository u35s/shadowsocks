package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	ss "github.com/u35s/shadowsocks"

	"github.com/u35s/glog"
)

func dealArgs() {
	log.SetOutput(os.Stdout)

	var cmdConfig ss.Config
	var configFile string

	flag.StringVar(&configFile, "c", "config.json", "specify config file")

	flag.StringVar(&cmdConfig.Password, "k", "", "password")
	flag.IntVar(&cmdConfig.ServerPort, "p", 0, "server port")
	flag.IntVar(&cmdConfig.Timeout, "t", 300, "timeout in seconds")
	flag.StringVar(&cmdConfig.Method, "m", "", "encryption method, default: aes-256-cfb")

	flag.Parse()
	var err error
	srv.conf, err = ss.ParseConfig(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", configFile, err)
			os.Exit(1)
		}
		glog.Err("[config],read err:%v", err)
		srv.conf = &cmdConfig
	} else {
		ss.UpdateConfig(srv.conf, &cmdConfig)
	}
	if srv.conf.Method == "" {
		srv.conf.Method = "aes-256-cfb"
	}
	if err = ss.CheckCipherMethod(srv.conf.Method); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = unifyPortPassword(); err != nil {
		os.Exit(1)
	}
	glog.Inf("[config],cur:%v", srv.conf)
}

func enoughOptions() bool {
	return srv.conf.ServerPort != 0 && srv.conf.Password != ""
}

func unifyPortPassword() (err error) {
	if len(srv.conf.PortPassword) == 0 { // this handles both nil PortPassword and empty one
		if !enoughOptions() {
			fmt.Fprintln(os.Stderr, "must specify both port and password")
			return errors.New("not enough options")
		}
		port := strconv.Itoa(srv.conf.ServerPort)
		srv.conf.PortPassword = map[string]string{port: srv.conf.Password}
	} else {
		if srv.conf.Password != "" || srv.conf.ServerPort != 0 {
			fmt.Fprintln(os.Stderr, "given port_password, ignore server_port and password option")
		}
	}
	return
}
