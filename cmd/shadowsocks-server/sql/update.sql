#20161106
alter table http_log add local_addr varchar(1024) not null after id;
alter table http_log add remote_addr varchar(1024) not null after local_addr;
