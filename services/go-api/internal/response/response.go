package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: "ok", Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{Code: "ok", Message: "created", Data: data})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Body{Code: code, Message: message})
}
