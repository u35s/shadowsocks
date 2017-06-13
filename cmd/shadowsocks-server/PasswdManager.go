package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"syscall"

	ss "github.com/u35s/shadowsocks"
	"github.com/u35s/shadowsocks/cmd/shadowsocks-server/zlib"

	"github.com/u35s/glog"
)

type PortListener struct {
	password string
	listener net.Listener
}

type PasswdManager struct {
	sync.Mutex
	portListener       map[string]*PortListener
	flowStatistic      map[string][2]uint
	flowStatisticQueue chan string
}

func (pm *PasswdManager) getFlowByPort(port string) (float64, float64) {
	if flow, ok := pm.flowStatistic[port]; ok {
		x := float64(1 << 20)
		return float64(flow[0]) / x, float64(flow[1]) / x
	}
	return 0, 0
}

func (pm *PasswdManager) init() {
	pm.portListener = make(map[string]*PortListener)
	pm.flowStatistic = make(map[string][2]uint)
	pm.flowStatisticQueue = make(chan string, 1<<8)
	for port, password := range srv.conf.PortPassword {
		go pm.run(port, password, srv.conf.Auth)
	}
	go func() {
		for {
			select {
			case fs := <-pm.flowStatisticQueue:
				if s := strings.Split(fs, "-"); len(s) == 3 {
					idx := zlib.Atoi(s[0])
					slc := pm.flowStatistic[s[1]]
					slc[idx] += zlib.Atou(s[2])
					pm.flowStatistic[s[1]] = slc
					glog.Inf("[流量统计],%v,类型:%v,总流量:%5fM", s[1], zlib.FlowType(idx), float64(slc[idx])/(1<<20))
				} else {
					glog.Err("[流量统计],格式错误,%v", fs)
				}
			}
		}
	}()
}
func (pm *PasswdManager) add(port, password string, listener net.Listener) {
	pm.Lock()
	pm.portListener[port] = &PortListener{password, listener}
	pm.Unlock()
}

func (pm *PasswdManager) get(port string) (pl *PortListener, ok bool) {
	pm.Lock()
	pl, ok = pm.portListener[port]
	pm.Unlock()
	return
}

func (pm *PasswdManager) del(port string) {
	pl, ok := pm.get(port)
	if !ok {
		return
	}
	pl.listener.Close()
	pm.Lock()
	delete(pm.portListener, port)
	pm.Unlock()
}

// Update port password would first close a port and restart listening on that
// port. A different approach would be directly change the password used by
// that port, but that requires **sharing** password between the port listener
// and password manager.
func (pm *PasswdManager) updatePortPasswd(port, password string, auth bool) {
	pl, ok := pm.get(port)
	if !ok {
		log.Printf("new port %s added\n", port)
	} else {
		if pl.password == password {
			return
		}
		log.Printf("closing port %s to update password\n", port)
		pl.listener.Close()
	}
	// run will add the new port listener to passwdManager.
	// So there maybe concurrent access to passwdManager and we need lock to protect it.
	go pm.run(port, password, auth)
}

func (pm *PasswdManager) run(port, password string, auth bool) {
	net.ResolveTCPAddr("tcp4", ":"+port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		glog.Err("[port],%v,error %v", port, err)
		return
	}
	pm.add(port, password, ln)
	var cipher *ss.Cipher
	pm.flowStatistic[port] = [2]uint{}
	glog.Inf("[port],%v,listening", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			glog.Err("[port],%v,acept,%v", port, err)
			return
		}
		// Creating cipher upon first connection.
		if cipher == nil {
			glog.Inf("[port],%v,creating cipher,%v,%v", port, srv.conf.Method, password)
			cipher, err = ss.NewCipher(srv.conf.Method, password)
			if err != nil {
				glog.Err("[port],%v,errror generating cipher %v", port, err)
				conn.Close()
				continue
			}
		}
		go pm.handleConnection(ss.NewConn(conn, cipher.Copy()), auth, port)
	}
}

const logCntDelta = 100

var connCnt int

func (pm *PasswdManager) handleConnection(conn *ss.Conn, auth bool, port string) {
	var host string

	connCnt++ // this maybe not accurate, but should be enough
	if connCnt%logCntDelta == 0 {
		// XXX There's no xadd in the atomic package, so it's difficult to log
		// the message only once with low cost. Also note nextLogConnCnt maybe
		// added twice for current peak connection number level.
		glog.Inf("[client],number of client connections reaches %d", connCnt)
	}
	glog.Inf("[client],new client %s->%s", conn.RemoteAddr(), conn.LocalAddr())
	closed := false
	defer func() {
		glog.Inf("[client],closed pipe %s<->%s\n", conn.RemoteAddr(), host)
		connCnt--
		if !closed {
			conn.Close()
		}
	}()

	host, ota, err := getRequest(conn, auth)
	if err != nil {
		glog.Err("[request],error getting request,%v,%v,%v", conn.RemoteAddr(), conn.LocalAddr(), err)
		return
	}
	glog.Dbg("[request],%v,connecting", host)
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			glog.Err("[request],dial error:%v", err)
		} else {
			glog.Err("[request],error connecting to:%v,%v", host, err)
		}
		return
	}
	defer func() {
		if !closed {
			remote.Close()
		}
	}()
	glog.Dbg("[request],piping %s<->%s ota=%v connOta=%v", conn.RemoteAddr(), host, ota, conn.IsOta())
	if ota {
		go ss.PipeThenCloseOta(conn, remote)
	} else {
		go func() {
			flow := ss.PipeThenClose(conn, remote)
			pm.flowStatisticQueue <- fmt.Sprintf("0-%v-%v", port, flow)
		}()
	}
	flow := ss.PipeThenClose(remote, conn)
	pm.flowStatisticQueue <- fmt.Sprintf("1-%v-%v", port, flow)
	closed = true
	return
}
