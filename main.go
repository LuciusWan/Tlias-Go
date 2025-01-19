package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type depts struct {
	Id         uint   `gorm:"primaryKey"`
	Name       string `gorm:"type:varchar(10);uniqueIndex"`
	CreateTime time.Time
	UpdateTime time.Time
}

func (depts) TableName() string {
	return "dept"
}

// ReadDept 从数据库中读取部门信息(读取所有)
func ReadDept(db *gorm.DB) ([]depts, error) {
	var deptList []depts
	// 获取所有记录
	result := db.Find(&deptList)
	if result.Error != nil {
		return nil, result.Error
	}
	return deptList, nil
}
func (d depts) MarshalJSON() ([]byte, error) {
	type Alias depts
	return json.Marshal(&struct {
		Id         uint      `json:"id"`
		Name       string    `json:"name"`
		CreateTime time.Time `json:"createdAt"`
		UpdateTime time.Time `json:"updatedAt"`
	}{
		Id:         d.Id,
		Name:       d.Name,
		CreateTime: d.CreateTime,
		UpdateTime: d.UpdateTime,
	})
}
func UpdateSelect(db *gorm.DB, id uint) (depts, error) {
	var dept depts
	result := db.First(&dept, "id=?", id)
	if result.Error != nil {
		return depts{}, result.Error
	}
	return dept, nil
}
func UpdateSave(db *gorm.DB, dept depts) error {
	result := db.Save(&dept)
	if result.Error != nil {
		return result.Error
	} else {
		return nil
	}
}
func CreateDept(db *gorm.DB, name string) error {
	dept := depts{
		Name:       name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	result := db.Create(&dept)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func main() {
	var US = depts{}
	r := gin.Default()
	// 数据库连接字符串
	dsn := "root:123456@tcp(127.0.0.1:3306)/springboottest?charset=utf8mb4&parseTime=True&loc=Local"
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}
	r.GET("/depts", func(c *gin.Context) {
		dept, err1 := ReadDept(db)
		if err1 != nil {
			c.JSON(401, gin.H{"code": "401", "message": err1.Error()})
		}
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "success",
			"data": dept,
		})
	})
	r.DELETE("/depts/:id", func(c *gin.Context) {
		// 从 URL 路径中获取 id 参数
		idStr := c.Param("id")
		var id uint
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Invalid id format",
				"data": nil,
			})
			return
		}

		// 删除指定 id 的部门
		result := db.Delete(&depts{}, id)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "删除部门信息失败",
				"data": nil,
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  "部门信息不存在",
				"data": nil,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "success",
			"data": nil,
		})
	})
	r.GET("/depts/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id uint
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Invalid id format",
				"data": nil,
			})
			return
		}
		dept, err := UpdateSelect(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 0,
				"msg":  err.Error(),
			})
		} else {
			US = dept
			US.UpdateTime = time.Now()
			c.JSON(http.StatusOK, gin.H{
				"code": 1,
				"msg":  "success",
				"data": dept,
			})
		}

	})
	r.PUT("/depts", func(c *gin.Context) {
		// 从请求体中解析 JSON 数据
		var requestData struct {
			Name string `json:"name"`
		}
		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Invalid request body",
				"data": nil,
			})
			return
		}
		US.Name = requestData.Name
		err := UpdateSave(db, US)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 0,
				"msg":  err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 1,
				"msg":  "success",
			})
		}
	})
	r.POST("/depts", func(c *gin.Context) {
		var requestData struct {
			Name string `json:"name"`
		}
		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Invalid request body",
				"data": nil,
			})
			return
		}
		err := CreateDept(db, requestData.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 0,
				"msg":  err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 1,
				"msg":  "success",
			})
		}
	})
	// 自动迁移模式，创建表
	db.AutoMigrate(&depts{})
	r.Run(":8080")
}
