package models

import "encoding/json"

type SSCAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SSCSCRFResponse struct {
	SCRFToken string `json:"result"`
}

type SSCDatabasesResponse struct {
	Count  int           `json:"count"`
	IDs    []int         `json:"ids"`
	Result []SSCDatabase `json:"result"`
}

type SSCDatabase struct {
	ID            int    `json:"id"`
	DatabaseName  string `json:"database_name"`
	Backend       string `json:"backend"`
	AllowRunAsync bool   `json:"allow_run_async"`
}

type SSCChartCreateResponse struct {
	ID int64 `json:"id"`
}

type SSCCreateDatasetResponse struct {
	ID int64 `json:"id"`
}

type SSCDatasetResponse struct {
	ID     int64          `json:"id"`
	Result SSCDatasetInfo `json:"result"`
}

type SSCDatasetInfo struct {
	Result struct {
		TableName      string `json:"table_name"`
		DatasourceID   int64  `json:"datasource_id"`
		DatasourceType string `json:"datasource_type"`
	} `json:"result"`
}

type SSCErrorResponse struct {
	Data json.RawMessage `json:"data"`
}
