package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/validation"
	"github.com/ummuys/reportify/internal/webdto"
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
// @Failure      401  {object}  webdto.EmptyResponse   "Неавторизован"
// @Failure      500  {object}  webdto.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/schemas [get]
func (mdh *mdHandler) GetSchemas() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		mdh.logger.Debug().Str("evt", "call GetSchemas")
		data, err := mdh.srv.GetSchemas(ctx)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternal.Error()})
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
// @Failure      400     {object} webdto.EmptyResponse   "Не указано имя схемы"
// @Failure      401     {object} webdto.EmptyResponse   "Неавторизован"
// @Failure      500     {object} webdto.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/tables [get]
func (mdh *mdHandler) GetTables() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		mdh.logger.Debug().Str("evt", "call GetTables")
		schema := g.Query("schema")
		if schema == "" {
			g.Set("msg", errs.ErrEmptySchemaName.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrEmptySchemaName.Error()})
			return
		}

		data, err := mdh.srv.GetTables(ctx, schema)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
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
// @Failure      400     {object} webdto.EmptyResponse   "Не указаны schema или table"
// @Failure      401     {object} webdto.EmptyResponse   "Неавторизован"
// @Failure      500     {object} webdto.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /db/columns [get]
func (mdh *mdHandler) GetColumns() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		mdh.logger.Debug().Str("evt", "call GetColumns")
		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			err := fmt.Errorf("%v ; %v", errs.ErrEmptySchemaName, errs.ErrEmptyTableName)
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}
		data, err := mdh.srv.GetColumns(ctx, schema, table)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}
		g.Set("msg", "column names are returned")
		g.JSON(http.StatusOK, data)
	}
}

// GetQueries godoc
// @Summary      Получить все сохранённые запросы пользователя
// @Description  Возвращает кэшированный список всех SQL-запросов для текущего пользователя (берётся по user_id из контекста).
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  webdto.QueryList        "Список запросов пользователя"
// @Failure      401  {object}  webdto.EmptyResponse    "Неавторизован"
// @Failure      500  {object}  webdto.EmptyResponse    "Внутренняя ошибка сервера"
// @Router       /cache [get]
func (mdh *mdHandler) GetQueries() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		user_id := g.GetInt64("user_id")
		queries, err := mdh.srv.GetQueries(ctx, strconv.FormatInt(user_id, 10))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		response := append([]byte{'['}, bytes.Join(queries, []byte(","))...)
		response = append(response, ']')

		g.Set("msg", "user queries are returned")
		g.Data(http.StatusOK, "application/json; charset=", response)
	}
}

// DeleteUserQueries godoc
// @Summary      Удалить все запросы пользователя
// @Description  Удаляет кэш всех SQL-запросов для текущего пользователя (user_id берётся из контекста).
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  webdto.EmptyResponse  "Запросы пользователя удалены"
// @Failure      401  {object}  webdto.EmptyResponse  "Неавторизован"
// @Failure      500  {object}  webdto.EmptyResponse  "Внутренняя ошибка сервера"
// @Router       /cache/all [delete]
func (mdh *mdHandler) DeleteAllQueries() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		user_id := g.GetInt64("user_id")
		if err := mdh.srv.DeleteAllQueries(ctx, strconv.FormatInt(user_id, 10)); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}
		msg := "all queries are deleted"
		g.Set("msg", msg)
		g.JSON(http.StatusOK, webdto.EmptyResponse{Message: msg})
	}
}

// DeleteUserQueries godoc
// @Summary      Удалить запрос пользователя
// @Description  Очищает кэш SQL-запроса для текущего пользователя (user_id берётся из контекста).
// @Tags         metadata
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body   webdto.DeleteQuery   true  "SQL-запрос для отчета"
// @Success      200  {object}  webdto.EmptyResponse  "Запрос пользователя удален"
// @Failure      401  {object}  webdto.EmptyResponse  "Неавторизован"
// @Failure      500  {object}  webdto.EmptyResponse  "Внутренняя ошибка сервера"
// @Router       /cache [delete]
func (mdh *mdHandler) DeleteQuery() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		user_id := g.GetInt64("user_id")
		var rawParam webdto.RawReportParams
		if err := g.ShouldBindJSON(&rawParam); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		param, err := validation.RequestParams(rawParam, false)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			g.Set("msg", err.Error())
			return
		}

		if err := mdh.srv.DeleteQuery(ctx, strconv.FormatInt(user_id, 10), param); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}
		msg := "query is deleted"
		g.Set("msg", msg)
		g.JSON(http.StatusOK, webdto.EmptyResponse{Message: msg})
	}
}
