package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ControllerV1 struct {
}

func NewControllerV1() *ControllerV1 {
	return &ControllerV1{}
}

func (con *ControllerV1) SayHello(c *gin.Context) {
	c.JSON(http.StatusOK, "Hello")
}
