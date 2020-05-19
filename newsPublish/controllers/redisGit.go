package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
)

type RedisGit struct {
	beego.Controller
}

func (this *RedisGit)ShowRedis()  {
	conn,err:=redis.Dial("tcp",":6379")
	if err!=nil{
		fmt.Println("redis连接错误",err)
		return
	}
	resp,err:=conn.Do("mget","kk","ll","s3")
	result,_:=redis.Values(resp,err)
	var v1,v3 string
	var v2 int
	redis.Scan(result,&v1,&v2,&v3)
	fmt.Println(v1,v2,v3)
}
