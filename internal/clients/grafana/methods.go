package grafanaclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ummuys/reportify/internal/models"
)

type grafcli struct {
	baseUrl string

	username string
	password string
	key      string
}

func NewGrafClient(baseURL, username, password, key string) GrafanaClient {
	return &grafcli{
		baseUrl:  baseURL,
		username: username,
		password: password,
		key:      key,
	}
}

func (g *grafcli) Auth() error {
	fmt.Println("Auth called")
	path := g.baseUrl + "/api/auth/keys"
	resp, err := g.httpDo(time.Second*time.Duration(5), nil, "GET", path, false, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	var body models.GrafcliAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	g.key = body.Key
	fmt.Printf("[OK] Key = %s", g.key)
	return nil
}

func (g *grafcli) ListFoalders() error {
	fmt.Println("ListFoalders called")
	return nil
}

func (g *grafcli) CreateFoalder() error {
	fmt.Println("CreateFoalder called")
	return nil
}

func (g *grafcli) DeleteFoalder() error {
	fmt.Println("DeleteFoalder called")
	return nil
}

func (g *grafcli) ListDashboards() ([]models.GrafcliDashboardListResponse, error) {
	fmt.Println("ListDashboards called")
	path := g.baseUrl + "/api/search?type=dash-db"
	resp, err := g.httpDo(time.Second*time.Duration(5), nil, "GET", path, false, false)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	var body []models.GrafcliDashboardListResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	for i, d := range body {
		fmt.Printf("%2d. %s\n", i+1, d.Title)
		fmt.Printf("	ID: %d\n", d.ID)
		fmt.Printf("    UID: %s\n", d.UID)
		fmt.Printf("    URL: %s\n", d.URL)
		fmt.Println(strings.Repeat("-", 40))
	}
	return body, nil
}

func (g *grafcli) DashboardInfo(uid string) ([]models.GrafcliPanel, error) {
	fmt.Println("DashboardInfo called")
	path := g.baseUrl + "/api/dashboards/uid/" + uid
	resp, err := g.httpDo(time.Second*time.Duration(5), nil, "GET", path, false, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	var body models.GrafcliDashboardInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	for i, p := range body.Dashboard.Panels {
		fmt.Printf("%2d. %s\n", i+1, p.Title)
		fmt.Printf("    ID:   %d\n", p.ID)
		fmt.Printf("    Type: %s\n", p.Type)
		fmt.Println(strings.Repeat("-", 40))
	}
	return body.Dashboard.Panels, nil
}

func (g *grafcli) CreateDashboard() error {
	fmt.Println("CreateDashboard called")
	return nil
}

func (g *grafcli) DeleteDashboard() error {
	fmt.Println("DeleteDashboard called")
	return nil
}

func (g *grafcli) ListDatasource() error {
	fmt.Println("ListDatasouse called")
	return nil
}

func (g *grafcli) CreateDatasource() error {
	fmt.Println("CreateDatasource called")
	return nil
}

func (g *grafcli) DeleteDatasource() error {
	fmt.Println("DeleteDatasouerse called")
	return nil
}

func (g *grafcli) CheckConnDatasources() error {
	fmt.Println("CheckConnDatasources called")
	return nil
}

func (g *grafcli) RenderChart(dashboardUID, slug string, chartID, width, height, scale int) error {
	fmt.Println("RenderChart called")
	path := fmt.Sprintf("%s/render/d-solo/%s/%s?orgId=1&panelId=%d&width=%d&height=%d&scale=%d",
		g.baseUrl, dashboardUID, slug, chartID, width, height, scale)

	resp, err := g.httpDo(time.Second*time.Duration(10), nil, "GET", path, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	fileName := fmt.Sprintf("chart_%d.png", chartID)
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("[OK] file saved")

	return nil
}

func (g *grafcli) RenderDashboard() error {
	fmt.Println("RenderDashboard called")
	return nil
}

func (g *grafcli) httpDo(timeout time.Duration, data any, method string, url string, needAccessToken bool, needBasicAuth bool) (*http.Response, error) {
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
		req.Header.Set("Authorization", "Bearer "+g.key)
	case needBasicAuth:
		req.SetBasicAuth("admin", "admin")
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %v", err)
	}

	return resp, nil
}

func (g *grafcli) CreateReport() error {
	return nil
}
