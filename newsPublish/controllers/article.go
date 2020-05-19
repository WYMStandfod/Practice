package controllers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"math"
	"path"
	"time"
	"newsPublish/models"
	"github.com/gomodule/redigo/redis"
)


type ArticleController struct {
	beego.Controller
}

func (this *ArticleController)ShowIndex()  {
	session:=this.GetSession("userName")
	if session==nil{
		this.Redirect("/login",302)
		return
	}
	o:=orm.NewOrm()

	var articleTypes []models.ArticleType
	//o.QueryTable("ArticleType").All(&articleTypes)
	conn,err:=redis.Dial("tcp",":6379")
	if err!=nil{
		fmt.Println("redis连接失败",err)
		return
	}
	defer conn.Close()
	data,err:=redis.Bytes((conn.Do("get","articleTypes")))
	if len(data)==0{
		o.QueryTable("ArticleType").All(&articleTypes)
		fmt.Println("从mysql中获取数据")
		var buffer bytes.Buffer
		encoder:=gob.NewEncoder(&buffer)
		encoder.Encode(&articleTypes)
		conn.Do("set","articleTypes",buffer.Bytes())
	}else{
		dec:=gob.NewDecoder(bytes.NewReader(data))
		dec.Decode(&articleTypes)
	}

	this.Data["articleTypes"]=articleTypes

	qs:=o.QueryTable("Article")
	var articles []models.Article
	qs.All(&articles)

	pageSize:=2


	pageIndex,err:=this.GetInt("pageIndex")
	if err!=nil{
		pageIndex=1
	}
	start:=pageSize*(pageIndex-1)

	typeName:=this.GetString("select")
	var Count int64
	if typeName==""{
		qs.Limit(pageSize,start).RelatedSel("ArticleType").All(&articles)
		Count,_=qs.RelatedSel("ArticleType").Count()
	}else{
		qs.Limit(pageSize,start).RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).All(&articles)
		Count,_=qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).Count()
	}
	pageCount:=math.Ceil(float64(Count)/float64(pageSize))

	this.Data["Count"]=Count
	this.Data["pageCount"]=pageCount
	//qs.Limit(pageSize,start).RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).All(&articles)
    this.Data["pageIndex"]=pageIndex
	this.Data["articles"]=articles

	this.Data["typeName"]=typeName
	this.Layout="layout.html"
	this.TplName="index.html"
}
func (this *ArticleController)ShowAdd()  {
	o:=orm.NewOrm()
	var articleTypes []models.ArticleType
	qs:=o.QueryTable("ArticleType")
	qs.All(&articleTypes)
	this.Data["articleTypes"]=articleTypes
	this.TplName="add.html"
}
func (this *ArticleController)HandleAdd()  {
	articleName:=this.GetString("articleName")
	content:=this.GetString("content")
	file,head,err:=this.GetFile("uploadname")
	if articleName==""||content==""||err!=nil{
		fmt.Println("获取用户添加数据失败",err)
		this.TplName="add.html"
		return
	}
	defer file.Close()
	if head.Size>5000000{
		fmt.Println("文件太大")
		this.TplName="add.html"
		return
	}
	ext:=path.Ext(head.Filename)
	if ext!=".png"&&ext!=".jpg"&&ext!=".jpeg"{
		fmt.Println("文件格式错误")
		this.TplName="add.html"
		return
	}
	fileName:=time.Now().Format("2006-01-02 15:04:05")

	this.SaveToFile("uploadname","./static/img/"+fileName+ext)

	o:=orm.NewOrm()
	var article models.Article
	article.Title=articleName
	article.Content=content
	article.Img="/static/img/"+fileName+ext

	TypeName:=this.GetString("select")
	var articleType models.ArticleType
	articleType.TypeName=TypeName
	o.Read(&articleType,"TypeName")

	article.ArticleType=&articleType

	_,err=o.Insert(&article)
	if err!=nil{
		fmt.Println("数据插入失败",err)
		this.TplName="add.html"
		return
	}
	this.Redirect("/index",302)
}
func (this *ArticleController)ShowContent() {
	id,err:=this.GetInt("id")
	if err!=nil{
		fmt.Println("id 获取失败")
		this.TplName="index.html"
		return
	}
	o:=orm.NewOrm()
	var article models.Article
	article.Id2=id
	err=o.Read(&article)
	var users []models.User
   o.QueryTable("User").Filter("Articles__Article__Id2",id).Distinct().All(&users)
	this.Data["users"]=users

	article.ReadCount+=1
	o.Update(&article)
	this.Data["article"]=article
	//this.Data["title"]="商品详情"
	m2m:=o.QueryM2M(&article,"Users")
	var user models.User
	userName:=this.GetSession("userName")
	user.Name=userName.(string)
	o.Read(&user,"Name")
	m2m.Add(user)

/*	this.LayoutSections=make(map[string]string)
	this.LayoutSections["jsFile"]="index.js"*/
    this.Layout="layout.html"
	this.TplName="content.html"
}
func UploadFunc(this *ArticleController,fileName string) string {
	file,head,err:=this.GetFile(fileName)
	if err!=nil{
		fmt.Println("获取用户添加数据失败",err)
		this.TplName="index.html"
		return ""
	}
	defer file.Close()
	if head.Size>5000000{
		fmt.Println("文件太大")
		this.TplName="index.html"
		return ""
	}
	ext:=path.Ext(head.Filename)
	if ext!=".png"&&ext!=".jpg"&&ext!=".jpeg"{
		fmt.Println("文件格式错误")
		this.TplName="index.html"
		return ""
	}
	filePath:=time.Now().Format("2006-01-02 15:04:05")

	this.SaveToFile(fileName,"./static/img/"+filePath+ext)
	return  "/static/img/"+filePath+ext

}
func (this *ArticleController)ShoweditArticle(){
	id,err:=this.GetInt("id")
	if err!=nil{
		fmt.Println("获取数据失败")
		this.TplName="index.html"
	}
	o:=orm.NewOrm()
	var article models.Article
	article.Id2=id
	o.Read(&article)
	this.Data["article"]=article
	this.TplName="update.html"
}
func (this *ArticleController)HandleEditArticle () {
	id,err:=this.GetInt("id")
	Title:=this.GetString("articleName")
	content:=this.GetString("content")
	filePath:=UploadFunc(this,"uploadname")
	if err!=nil||Title==""||content==""||filePath==""{
		fmt.Println("获取数据错误")
		this.TplName="update.html"
		return
	}
	o:=orm.NewOrm()
	var article models.Article
	article.Id2=id
	err=o.Read(&article,"Id2")
	if err!=nil{
		fmt.Println("数据不存在")
		this.TplName="update.html"
		return
	}
	article.Title=Title
	article.Content=content
	article.Img=filePath
	o.Update(&article)
	this.Redirect("/index",302)



}
func (this *ArticleController)HandleDeleteArticle(){
	id,err:=this.GetInt("id")
	if err!=nil{
		fmt.Println("id  获取失败")
		this.TplName="index.html"
		return
	}
	o:=orm.NewOrm()
	var article models.Article
	article.Id2=id
	_,err=o.Delete(&article)
	if err!=nil{
		fmt.Println("删除失败")
	}
	this.Redirect("/index",302)
}
func (this *ArticleController)ShowAddType(){
	o:=orm.NewOrm()
	qs:=o.QueryTable("ArticleType")
	var articleTypes []models.ArticleType
	qs.All(&articleTypes)
	this.Data["articleTypes"]=articleTypes
	this.TplName="addType.html"
}
func (this *ArticleController)HandleAddType() {
	typeName:=this.GetString("typeName")
	if typeName==""{
		fmt.Println("文章类型不能为空")
		this.TplName="addType.html"
		return
	}
	o:=orm.NewOrm()
	var ArticleType models.ArticleType
	ArticleType.TypeName=typeName
	o.Insert(&ArticleType)
	this.Redirect("/addType",302)
}
func (this *ArticleController)DeleteType() {
	id,err:=this.GetInt("id")
	if err!=nil{
		fmt.Println("Id 获取失败",err)
		this.TplName="addType.html"
		return
	}
	o:=orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id=id
	o.Delete(&articleType)
	this.Redirect("/article/addType",302)

}