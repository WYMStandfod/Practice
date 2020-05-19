package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"newsPublish/models"
)

type UserController struct{
	beego.Controller
}

func (this *UserController)ShowRegister()  {
	this.TplName="register.html"
}
func (this *UserController)HandleRegister() {
	//把数据插入数据库
 userName:=this.GetString("userName")
 passwd:=this.GetString("password")
 if userName==""||passwd==""{
 	fmt.Println("用户名和密码不能为空")
 	this.TplName="register.html"
	 return
 }
 o:=orm.NewOrm()
 var user models.User
 user.Name=userName
 user.Pwd=passwd
 _,err:=o.Insert(&user)
 if err!=nil{
 	fmt.Println("数据插入失败")
 	this.TplName="register.html"
	 return
 }
 //this.Ctx.WriteString("用户注册成功")
 this.Redirect("/login",302)
}
func (this *UserController)ShowLogin()  {

	username:=this.Ctx.GetCookie("userName")
	if username==""{
		this.Data["userName"]=""
		this.Data["checked"]=""
	}else {
		this.Data["userName"]=username
		this.Data["checked"]="checked"
	}
	this.TplName="login.html"
}
func (this *UserController)HandleLogin()  {
   userName:=this.GetString("userName")
   passwd:=this.GetString("password")

   if userName==""||passwd==""{
   	fmt.Println("用户名和密码不能为空")
	   return
   }
   o:=orm.NewOrm()
   var user  models.User
   //
   user.Name=userName
   err:=o.Read(&user,"Name")
   if err!=nil{
   	fmt.Errorf("用户不存在\n")
   	this.TplName="login.html"
	   return
   }
  if user.Pwd!=passwd{
  	fmt.Println("输入密码错误")
  	this.TplName="login.html"
	  return
  }
  remember:=this.GetString("remember")
  if remember=="on"{
  	this.Ctx.SetCookie("userName",userName,60*60*24)
  }else if remember==""{
  	this.Ctx.SetCookie("userName",userName,-1)
  }
  this.SetSession("userName",userName)
  //this.Ctx.WriteString("登陆成功")
  this.Redirect("/article/index",302)
}
func (this *UserController)LogOut(){
	this.DelSession("userName")
	this.Redirect("/login",302)
}