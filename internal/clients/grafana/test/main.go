package main

import grafcli "github.com/ummuys/reportify/internal/clients/grafana"

func main() {
	cli := grafcli.NewGrafClient("http://127.0.0.1:3000", "admin", "admin", "")
	// if err := cli.Auth(); err != nil {
	// 	panic(err)
	// }

	list, err := cli.ListDashboards()
	if err != nil {
		panic(err)
	}

	panels, err := cli.DashboardInfo(list[0].UID)
	if err != nil {
		panic(err)
	}

	if err := cli.RenderChart(list[0].UID, "test-dashboard", panels[0].ID, 1000, 500, 2); err != nil {
		panic(err)
	}
}
