package models

type GrafcliAuthResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type GrafcliDashboardListResponse struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type GrafcliDashboardInfoResponse struct {
	Dashboard struct {
		Title  string         `json:"title"`
		UID    string         `json:"uid"`
		Panels []GrafcliPanel `json:"panels"`
	} `json:"dashboard"`
}

type GrafcliPanel struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
