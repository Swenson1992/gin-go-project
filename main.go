package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Test struct {
	gorm.Model
	Id     int
	Remark TestObj `gorm:"type:json"`
}
type TestObj struct {
	c1 string
	c2 string
}

//定义struct
type AVL_HDD struct {
	gorm.Model
	AVLType  string
	Capacity int
	Status   string
	Arch     string
	Spec     int
}

type StringMap struct {
	Src   map[string]string
	Valid bool
}

func NewEmptyStringMap() *StringMap {
	return &StringMap{
		Src:   make(map[string]string),
		Valid: true,
	}
}

func NewStringMap(src map[string]string) *StringMap {
	return &StringMap{
		Src:   src,
		Valid: true,
	}
}

func (ls *StringMap) Scan(value interface{}) error {
	if value == nil {
		ls.Src, ls.Valid = make(map[string]string), false
		return nil
	}
	t := make(map[string]string)
	if e := json.Unmarshal(value.([]byte), &t); e != nil {
		return e
	}
	ls.Valid = true
	ls.Src = t
	return nil
}

func (ls *StringMap) Value() (driver.Value, error) {
	if ls == nil {
		return nil, nil
	}
	if !ls.Valid {
		return nil, nil
	}

	b, e := json.Marshal(ls.Src)
	return b, e
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	// gin.ForceConsoleColor()
	// f, _ := os.Create("log/gin.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	// router := gin.New()
	// router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

	// 	// your custom format
	// 	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format(time.RFC1123),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))
	// router.Use(gin.Recovery())
	db, err := gorm.Open("mysql", "root:root@/sod?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.SingularTable(true)

	var checkTable = db.HasTable("avl_hdd")
	fmt.Println("checkTable result:", checkTable)

	// fmt.Println(test)
	var rows []AVL_HDD
	//select
	db.Where("id=?", 1).Select([]string{"AVLType", "Capacity", "Status", "Arch", "Remark"}).Find(&rows)

	fmt.Println("rows:", rows)

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	router.GET("/user/:name/*action?q", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		q := c.Query("id")
		message := name + " is " + action + " q is " + q
		c.String(http.StatusOK, message)
	})

	router.POST("/upload", func(c *gin.Context) {
		name := c.PostForm("name")
		fmt.Println(name)
		file, header, err := c.Request.FormFile("upload")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}
		filename := header.Filename

		fmt.Println(file, err, filename)

		out, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatal(err)
		}
		c.String(http.StatusCreated, "upload successful")
	})

	router.GET("/download/*name", func(c *gin.Context) {
		filePath := c.Param("name")
		// 设置浏览器是否为直接下载文件，且为浏览器指定下载文件的名字
		c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(path.Base(filePath)))
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")
		// 如果缓存过期了，会再次和原来的服务器确定是否为最新数据，而不是和中间的proxy
		c.Header("Cache-Control", "must-revalidate")
		c.Header("Pragma", "public")
		c.File(filePath)
	})

	router.Run(":8000")
}
