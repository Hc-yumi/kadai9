package controller

import (
	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("view/*.html")

	r.GET("/", getTop)
	r.GET("/signup", getSignup)
	r.POST("/signup", postSignup)
	r.GET("/login", getLogin)
	r.POST("/login", postLogin)

	return r
}