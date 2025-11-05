package grafanaclient

import "github.com/ummuys/reportify/internal/models"

type GrafanaClient interface {
	Auth() error

	ListFoalders() error
	CreateFoalder() error
	DeleteFoalder() error

	ListDashboards() ([]models.GrafcliDashboardListResponse, error)
	DashboardInfo(uid string) ([]models.GrafcliPanel, error)
	CreateDashboard() error
	DeleteDashboard() error

	ListDatasource() error
	CreateDatasource() error
	DeleteDatasource() error
	CheckConnDatasources() error

	RenderChart(dashboardUID, slug string, chartID, width, height, scale int) error
	RenderDashboard() error

	CreateReport() error
}
