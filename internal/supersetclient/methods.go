package supersetclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ummuys/reportify/internal/models"
)

type ssclient struct {
	username     string
	password     string
	accessToken  string
	refreshToken string
	csrfToken    string
}

func NewSupersetClient(username string, password string) SupersetClient {
	return &ssclient{username: username, password: password}
}

func (ss *ssclient) Login() error {
	data := models.SSCAuthRequest{
		Username: ss.username,
		Password: ss.password,
		Refresh:  true,
		Provider: "db",
	}

	resp, err := ss.httpDo(time.Second*time.Duration(5), data, "POST", "http://localhost:8088/api/v1/security/login", false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return errors.New(string(bud))
	}

	var respData models.SSCAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("can't decode: %v", err)
	}

	ss.accessToken = respData.AccessToken
	ss.refreshToken = respData.RefreshToken
	return nil
}

func (ss *ssclient) CreateDatabaseConn() error {
	req := models.SSCDBCreateRequest{
		ConfigurationMethod: "sqlalchemy_form",
		DatabaseName:        "report_db",
		Engine:              "postgresql",
		SQLAlchemyURI:       "postgresql+psycopg2://admin:admin@localhost:5434/report",
		ExposeInSQLLab:      true,
		AllowCTAS:           true,
		AllowCVAS:           true,
		AllowDML:            true,
	}
	resp, err := ss.httpDo(time.Second*time.Duration(180), req, "POST", "http://localhost:8088/api/v1/database/", true, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bud, _ := io.ReadAll(resp.Body)
	fmt.Println(string(bud))

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return errors.New(string(bud))
	}

	return nil

}

func (ss *ssclient) GetListDatabase() (models.SSCDatabasesResponse, error) {
	resp, err := ss.httpDo(time.Second*time.Duration(5), nil, "GET", "http://localhost:8088/api/v1/database/", true, false)
	if err != nil {
		return models.SSCDatabasesResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return models.SSCDatabasesResponse{}, errors.New(string(bud))
	}

	var respData models.SSCDatabasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return models.SSCDatabasesResponse{}, fmt.Errorf("can't decode response body: %v", err)
	}

	return respData, nil
}

func (ss *ssclient) CreateReport(schema string, tableName string, sql string, toCreateDataset bool) error {
	if toCreateDataset {
		ss.CreateDataset(schema, tableName, sql)
	}

	return nil
}

func (ss *ssclient) CreateDataset(schema string, tableName string, sql string) (int64, error) {
	data := models.SSCCreateDatasetRequest{
		AlwaysFilterMainDttm: false,
		Catalog:              "",
		Database:             2,
		ExternalURL:          "",
		IsManagedExternally:  false,
		NormalizeColumns:     false,
		Schema:               schema,
		Owners:               []int{1},
		SQL:                  sql,
		TableName:            tableName,
	}

	resp, err := ss.httpDo(time.Second*time.Duration(5), data, "POST", "http://localhost:8088/api/v1/dataset/", true, false)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return -1, errors.New(string(bud))
	}

	var respData models.SSCCreateDatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return -1, fmt.Errorf("can't decode response body: %v", err)
	}

	return respData.ID, nil
}

func (ss *ssclient) CreateChart(datasetID int64) (int64, error) {
	data := models.SSCChartCreateRequest{
		DatasourceID:           datasetID,
		DatasourceType:         "table", // или "query", если dataset виртуальный
		SliceName:              "All records view",
		VizType:                "table",
		QueryContextGeneration: true,
		IsManagedExternally:    false,
		CacheTimeout:           0,
		Params: `{
			"query_mode": "raw",
			"all_columns": [],
			"row_limit": 1000
		}`,
	}

	resp, err := ss.httpDo(time.Second*time.Duration(5), data, "POST", "http://localhost:8088/api/v1/chart/", true, false)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return -1, errors.New(string(bud))
	}

	var respData models.SSCChartCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return -1, fmt.Errorf("can't decode response body: %v", err)
	}

	return respData.ID, nil
}

// ####################################################################################
// ############################     HELP FUNCTION    ##################################
// ####################################################################################

func (ss *ssclient) getCSRFToken() error {
	resp, err := ss.httpDo(time.Second*time.Duration(5), nil, "GET", "http://localhost:8088/api/v1/security/csrf_token/", true, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		bud, _ := io.ReadAll(resp.Body)
		return errors.New(string(bud))
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
	// path := fmt.Sprintf("http://localhost:8088/api/v1/dataset/%d", datasetID)
	// resp, err := ss.httpDo(time.Second*time.Duration(5), nil, "GET", path, true, false)
	// if err != nil {
	// 	return models.SSCDatasetInfo{}, err
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode >= http.StatusBadRequest {
	// 	var invalid models.SSCInvalidResponse
	// 	if err := json.NewDecoder(resp.Body).Decode(&invalid); err != nil {
	// 		return models.SSCDatasetInfo{}, fmt.Errorf("can't decode body: %v", err)
	// 	}
	// 	return models.SSCDatasetInfo{}, fmt.Errorf("status code: %d & msg: %s", resp.StatusCode, invalid.Message)
	// }

	var respData models.SSCDatasetResponse
	// // if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
	// // 	return models.SSCDatasetInfo{}, fmt.Errorf("can't decode response body: %v", err)
	// // }

	// body, _ := io.ReadAll(resp.Body)
	// fmt.Println("RAW RESPONSE:")
	// fmt.Println(string(body))

	return respData.Result, nil

}

// Don't forget to close *http.Response
func (ss *ssclient) httpDo(timeout time.Duration, data any, method string, url string, needAccessToken bool, needCSRFToken bool) (*http.Response, error) {
	var (
		body []byte
		err  error
	)

	body, err = json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("can't convert to json: %v", err)
	}

	cl := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("can't create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	switch {
	case needAccessToken:
		req.Header.Set("Authorization", "Bearer "+ss.accessToken)
	case needCSRFToken:
		req.Header.Set("X-CSRF-Token", ss.csrfToken)
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %v", err)
	}

	return resp, nil
}
