package handlers

import (
	"github.com/gin-gonic/gin"
)

type MetadataHandler interface {
	GetSchemas() gin.HandlerFunc
	GetTables() gin.HandlerFunc
	GetColumns() gin.HandlerFunc
	GetQueries() gin.HandlerFunc
	DeleteAllQueries() gin.HandlerFunc
	DeleteQuery() gin.HandlerFunc
}
