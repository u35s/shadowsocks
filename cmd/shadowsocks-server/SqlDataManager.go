package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/u35s/glog"
)

type SqlDataManager struct {
	db    *xorm.Engine
	users map[string]*User
}

func (this *SqlDataManager) init() {
	engine, err := xorm.NewEngine("mysql", srv.conf.Mysql)
	if err != nil {
		glog.Err("[数据库],创建数据库引擎错误,%v", err)
	} else {
		if err = engine.Ping(); err == nil {
			this.db = engine
			glog.Inf("[数据库],数据库连接成功,%v,engine:%v,err:%v", srv.conf.Mysql, engine, err)
			this.users = make(map[string]*User)
			this.load()
		} else {
			glog.Err("[数据库],连接数据库错误,%v", err)
		}
	}
}

func (this *SqlDataManager) load() {
	this.loadUsers()
}

type User struct {
	Id          uint `xorm:"autoincr"`
	Port        string
	Password    string
	CreatedTime uint `xorm:"created"`
}

func (this *SqlDataManager) delUser(port, pwd string) {
	if user, ok := this.users[port]; ok {
		aft, err := this.db.Id(user.Id).Delete(user)
		if err == nil {
			glog.Inf("[用户],删除成功,Id:%v,port:%v,pwd:%v", user.Id, port, pwd)
		} else {
			glog.Err("[用户],删除失败,%v,%v", aft, err)
		}
	} else {
		glog.Err("[用户],找不到端口,%v,%v", port, pwd)
	}

}

func (this *SqlDataManager) updateUser(port, pwd string) {
	user, ok := this.users[port]
	if ok {
		user.Password = pwd
	} else {
		glog.Err("[用户],找不到端口,%v,%v", port, pwd)
		return
	}
	aft, err := this.db.Id(user.Id).Update(user)
	if err == nil {
		glog.Inf("[用户],更新成功,Id:%v,port:%v,pwd:%v", user.Id, port, pwd)
	} else {
		glog.Err("[用户],更新失败,%v,%v", aft, err)
	}
}

func (this *SqlDataManager) addUser(port, pwd string) {

	var user User
	user.Port = port
	user.Password = pwd
	id, err := this.db.Insert(&user)
	if err == nil {
		this.users[port] = &user
		glog.Inf("[用户],添加成功,Id:%v,port:%v,pwd:%v", user.Id, port, pwd)
	} else {
		glog.Err("[用户],添加失败,%v,%v", id, err)
	}
}

func (this *SqlDataManager) loadUsers() {
	u := new(User)
	rows, err := this.db.Rows(u)
	if err != nil {
		glog.Err("[用户],加载读取错误,%v", err)
	}
	var num uint = 0
	defer rows.Close()
	for rows.Next() {
		user := new(User)
		err = rows.Scan(user)
		if err == nil {
			this.users[user.Port] = user
			if _, ok := srv.pwm.get(user.Port); ok {
				glog.Inf("[用户],port:%v already run", user.Port)
				continue
			}
			go srv.pwm.run(user.Port, user.Password, srv.conf.Auth)
			num++
			glog.Inf("[用户],add,port:%v,pwd:%v", user.Port, user.Password)
		} else {
			glog.Err("[用户],映射数据错误,%v", err)
		}
		//...
	}
	glog.Inf("[用户],加载%v个", num)
}
