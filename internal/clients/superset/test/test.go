package main

// import (
// 	"fmt"

// 	"github.com/ummuys/reportify/internal/client/supersetclient"
// )

// func main() {
// 	ssc := supersetclient.NewSupersetClient("admin", "admin")
// 	if err := ssc.Login(); err != nil {
// 		fmt.Println(err)
// 	}

// 	// if err := ssc.CreateDatabaseConn(); err != nil {
// 	// 	fmt.Println(err)
// 	// }

// 	// gdl, err := ssc.GetListDatabase()
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// } else {
// 	// 	fmt.Println(gdl)
// 	// }

// 	// datasetID, err := ssc.CreateDataset(1, "test", "check", "select * from test.check limit 100;")
// 	// if err != nil {
// 	// 	fmt.Print(err)
// 	// } else {
// 	// 	fmt.Println("All ok")
// 	// 	fmt.Printf("dataset id -> %v\n", datasetID)
// 	// }

// 	chartID, err := ssc.CreateChart(1)
// 	if err != nil {
// 		fmt.Print(err)
// 	} else {
// 		fmt.Println("All ok")
// 		fmt.Printf("chart num -> %v", chartID)
// 	}

// }
