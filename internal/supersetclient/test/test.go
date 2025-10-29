package main

import (
	"fmt"

	"github.com/ummuys/reportify/internal/supersetclient"
)

func main() {
	ssc := supersetclient.NewSupersetClient("admin", "admin")
	if err := ssc.Login(); err != nil {
		fmt.Println(err)
	}

	if err := ssc.CreateDatabaseConn(); err != nil {
		panic(err)
	}

	// gdl, err := ssc.GetListDatabase()
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(gdl)
	// }

	// datasetID, err := ssc.CreateDataset("rzd", "2018_01_1", "select * from rgd.\"2018_01_1\" limit 100;")
	// if err != nil {
	// 	fmt.Print(err)
	// } else {
	// 	fmt.Println("All ok")
	// 	fmt.Printf("dataset id -> %v\n", datasetID)
	// }

	// info, err := ssc.GetDataSourceInfo(28)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(info)
	// }

	// chartID, err := ssc.CreateChart(28)
	// if err != nil {
	// 	fmt.Print(err)
	// } else {
	// 	fmt.Println("All ok")
	// 	fmt.Printf("chart num -> %v", chartID)
	// }
}
