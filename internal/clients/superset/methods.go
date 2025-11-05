package supersetclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ummuys/reportify/internal/models"
)

// ----------------------------------------------------------------------
//                           STRUCTURES
// ----------------------------------------------------------------------

type ssclient struct {
	url          string
	username     string
	password     string
	accessToken  string
	refreshToken string
	csrfToken    string
}

// ----------------------------------------------------------------------
//                           CONSTRUCTOR
// ----------------------------------------------------------------------

func NewSupersetClient(username, password string) SupersetClient {
	return &ssclient{
		url:      "http://localhost:9099",
		username: username,
		password: password,
	}
}

// ----------------------------------------------------------------------
//                           AUTH
// ----------------------------------------------------------------------

func (ss *ssclient) Login() error {
	data := models.SSCAuthRequest{
		Username: ss.username,
		Password: ss.password,
		Refresh:  true,
		Provider: "db",
	}

	path := fmt.Sprintf("%s/api/v1/security/login", ss.url)
	resp, err := ss.httpDo(5*time.Second, data, "POST", path, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return err
		}
		fmt.Println("[ERROR] Login failed:")
		fmt.Println(respBody)
		return errors.New("login error")
	}

	var respData models.SSCAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("can't decode response: %v", err)
	}

	ss.accessToken = respData.AccessToken
	ss.refreshToken = respData.RefreshToken
	return nil
}

// ----------------------------------------------------------------------
//                           DATABASE
// ----------------------------------------------------------------------

func (ss *ssclient) CreateDatabaseConn() error {
	req := models.SSCDBCreateRequest{
		ConfigurationMethod: "sqlalchemy_form",
		DatabaseName:        "report_db",
		Engine:              "postgresql",
		SQLAlchemyURI:       "postgresql+psycopg2://admin:admin@db-report:5432/report",
		ExposeInSQLLab:      true,
		AllowCTAS:           true,
		AllowCVAS:           true,
		AllowDML:            true,
		UUID:                uuid.New().String(),
	}

	path := fmt.Sprintf("%s/api/v1/database/", ss.url)
	resp, err := ss.httpDo(180*time.Second, req, "POST", path, true, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return err
		}
		fmt.Println("[ERROR] Database creation failed:")
		fmt.Println(respBody)
		return errors.New("create database error")
	}

	return nil
}

func (ss *ssclient) GetListDatabase() (models.SSCDatabasesResponse, error) {
	path := fmt.Sprintf("%s/api/v1/database/", ss.url)
	resp, err := ss.httpDo(5*time.Second, nil, "GET", path, true, false)
	if err != nil {
		return models.SSCDatabasesResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return models.SSCDatabasesResponse{}, err
		}
		fmt.Println("[ERROR] Database list fetch failed:")
		fmt.Println(respBody)
		return models.SSCDatabasesResponse{}, errors.New("get database list error")
	}

	var respData models.SSCDatabasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return models.SSCDatabasesResponse{}, fmt.Errorf("can't decode response body: %v", err)
	}

	return respData, nil
}

// ----------------------------------------------------------------------
//                           DATASET
// ----------------------------------------------------------------------

func (ss *ssclient) CreateDataset(databaseID int, schema, tableName, sql string) (int64, error) {
	data := models.SSCCreateDatasetRequest{
		AlwaysFilterMainDttm: false,
		Catalog:              "",
		Database:             databaseID,
		ExternalURL:          "",
		IsManagedExternally:  false,
		NormalizeColumns:     false,
		Schema:               schema,
		Owners:               []int{1},
		SQL:                  sql,
		TableName:            tableName,
	}

	path := fmt.Sprintf("%s/api/v1/dataset/", ss.url)
	resp, err := ss.httpDo(5*time.Second, data, "POST", path, true, false)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return -1, err
		}
		fmt.Println("[ERROR] Dataset creation failed:")
		fmt.Println(respBody)
		return -1, errors.New("create dataset error")
	}

	var respData models.SSCCreateDatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return -1, fmt.Errorf("can't decode response body: %v", err)
	}

	return respData.ID, nil
}

// ----------------------------------------------------------------------
//                           CHART
// ----------------------------------------------------------------------

func (ss *ssclient) CreateChart(datasetID int64) (int64, error) {
	data := models.SSCChartCreateRequest{
		DatasourceID:           datasetID,
		DatasourceType:         "table",
		SliceName:              "All records view",
		VizType:                "table",
		QueryContextGeneration: true,
		IsManagedExternally:    false,
		CacheTimeout:           0,
		Params: `{
			"query_mode": "raw",
			"all_columns": ["id", "username", "email", "created_at", "is_active", "balance"],
			"row_limit": 1000
		}`,
	}

	path := fmt.Sprintf("%s/api/v1/chart/", ss.url)
	resp, err := ss.httpDo(5*time.Second, data, "POST", path, true, false)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return -1, err
		}
		fmt.Println("[ERROR] Chart creation failed:")
		fmt.Println(respBody)
		return -1, errors.New("create chart error")
	}

	var respData models.SSCChartCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return -1, err
	}

	return respData.ID, nil
}

// ----------------------------------------------------------------------
//                           HELPERS
// ----------------------------------------------------------------------

// {
//   "datasource": {"id": 1, "type": "table"},
//   "force": false,
//   "queries": [
//     {
//       "columns": ["id", "username", "email", "created_at", "is_active", "balance"],
//       "row_limit": 1000,
//       "order_desc": true
//     }
//   ],
//   "result_format": "json",
//   "result_type": "full"
// }

func (ss *ssclient) GetInfoChart(id int) error {
	return nil
}

func (ss *ssclient) CreateReport(databaseID int, schema, tableName, sql string, toCreateDataset bool) error {
	return nil
}

func (ss *ssclient) ExportChart(chartID int, format string) error {
	return nil
}

func (ss *ssclient) getCSRFToken() error {
	path := fmt.Sprintf("%s/api/v1/security/csrf_token/", ss.url)
	resp, err := ss.httpDo(5*time.Second, nil, "GET", path, true, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var respBody models.SSCErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return err
		}
		fmt.Println("[ERROR] CSRF token request failed:")
		fmt.Println(respBody)
		return errors.New("csrf token error")
	}

	var respData models.SSCSCRFResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("can't decode response body: %v", err)
	}

	ss.csrfToken = respData.SCRFToken
	return nil
}

func (ss *ssclient) updateAccess() error {
	return nil
}

func (ss *ssclient) GetDataSourceInfo(datasetID int64) (models.SSCDatasetInfo, error) {
	var respData models.SSCDatasetResponse
	return respData.Result, nil
}

// ----------------------------------------------------------------------
//                           HTTP CLIENT
// ----------------------------------------------------------------------

func (ss *ssclient) httpDo(
	timeout time.Duration,
	data any,
	method string,
	url string,
	needAccessToken bool,
	needCSRFToken bool,
) (*http.Response, error) {

	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("can't convert to JSON: %v", err)
	}

	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("can't create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if needAccessToken {
		req.Header.Set("Authorization", "Bearer "+ss.accessToken)
	} else if needCSRFToken {
		req.Header.Set("X-CSRF-Token", ss.csrfToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return resp, nil
}
