package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/models"
	"github.com/ummuys/reportify/internal/service"
)

type mdHandler struct {
	logger *zerolog.Logger
	srv    service.MetadataService
}

func NewMetadataHandler(logger *zerolog.Logger, srv service.MetadataService) MetadataHandler {
	return &mdHandler{logger: logger, srv: srv}
}

// GetSchemas godoc
// @Summary      Получить список схем БД
// @Description  Возвращает имена доступных схем.
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   string                 "Имена схем"
// @Failure      401  {object}  models.EmptyResponse   "Неавторизован"
// @Failure      500  {object}  models.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/schemas [get]
func (mdh *mdHandler) GetSchemas(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		mdh.logger.Debug().Str("evt", "call GetSchemas")
		data, err := mdh.srv.GetSchemas(pCtx)
		if err != nil {
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "schema names are returned")
		g.JSON(http.StatusOK, data)

	}
}

// GetTables godoc
// @Summary      Получить таблицы в схеме
// @Description  Возвращает имена таблиц для заданной схемы.
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        schema  query    string  true  "Имя схемы"
// @Success      200     {array}  string                 "Имена таблиц"
// @Failure      400     {object} models.EmptyResponse   "Не указано имя схемы"
// @Failure      401     {object} models.EmptyResponse   "Неавторизован"
// @Failure      500     {object} models.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/tables [get]
func (mdh *mdHandler) GetTables(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		mdh.logger.Debug().Str("evt", "call GetTables")
		schema := g.Query("schema")
		if schema == "" {
			g.Set("msg", "schema name: "+schema)
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "schema name is required"})
			return
		}

		data, err := mdh.srv.GetTables(pCtx, schema)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "table names are returned")
		g.JSON(http.StatusOK, data)

	}
}

// GetColumns godoc
// @Summary      Получить колонки таблицы
// @Description  Возвращает имена колонок для указанной схемы и таблицы.
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        schema  query    string  true  "Имя схемы"
// @Param        table   query    string  true  "Имя таблицы"
// @Success      200     {array}  string                 "Имена колонок"
// @Failure      400     {object} models.EmptyResponse   "Не указаны schema или table"
// @Failure      401     {object} models.EmptyResponse   "Неавторизован"
// @Failure      500     {object} models.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/columns [get]
func (mdh *mdHandler) GetColumns(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		mdh.logger.Debug().Str("evt", "call GetColumns")
		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			g.Set("msg", "schema: "+schema+"; table: "+table)
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "schema and table names are required"})
			return
		}
		data, err := mdh.srv.GetColumns(pCtx, schema, table)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "column names are returned")
		g.JSON(http.StatusOK, data)

	}
}

// GetCacheQueries godoc
// @Summary      Получить сохранённые запросы пользователя
// @Description  Возвращает кэшированный список SQL-запросов для текущего пользователя (берётся по user_id из контекста).
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.QueryList        "Список запросов пользователя"
// @Failure      401  {object}  models.EmptyResponse    "Неавторизован"
// @Failure      500  {object}  models.EmptyResponse    "Внутренняя ошибка сервера"
// @Router       /cache [get]
func (mdh *mdHandler) GetCacheQueries(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		user_id := g.GetInt64("user_id")
		queries, err := mdh.srv.GetCacheQueries(pCtx, strconv.FormatInt(user_id, 10))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "list Queries are returned")
		g.JSON(http.StatusOK, models.QueryList{Queries: queries})
	}
}
