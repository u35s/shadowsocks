#默认账户列表
drop table if exists user;
create table user (
 	id int(10) unsigned not null auto_increment,             #ID自动增长
    port varchar(32) not null,         						 # 端口号
	password varchar(32) not null,          				 # 密码
	created_time int(10) unsigned not null,  				 # 创建时间
    primary key(id)
) engine=InnoDB default charset=utf8;

drop table if exists http_log;
create table http_log (
 	id bigint(10) unsigned not null auto_increment,             #ID自动增长
	local_addr varchar(1024) not null,          				 	
	remote_addr varchar(1024) not null,          				 	
	header varchar(1024) not null,          				 	
	compress int(10) not null,          				
    compress_len int(10) not null,         						
    body longtext not null,         						    #内容
    body_len int(10) not null,         						
    primary key(id)
) engine=InnoDB default charset=utf8;
