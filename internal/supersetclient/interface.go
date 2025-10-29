package supersetclient

import "github.com/ummuys/reportify/internal/models"

type SupersetClient interface {
	Login() error

	// CREATE
	CreateDatabaseConn() error
	CreateDataset(schema string, tableName string, sql string) (int64, error)
	CreateChart(datasetID int64) (int64, error)

	// GET
	GetListDatabase() (models.SSCDatabasesResponse, error)
	GetDataSourceInfo(datasetID int64) (models.SSCDatasetInfo, error)
}
