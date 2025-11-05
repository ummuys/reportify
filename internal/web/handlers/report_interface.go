package handlers

import (
	"github.com/gin-gonic/gin"
)

type ReportHandler interface {
	CreateReport() gin.HandlerFunc
}
