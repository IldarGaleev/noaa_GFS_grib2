package main

import (
	"context"
	"fmt"
	"gfsloader/internal/models"
	"gfsloader/internal/storage/postgres"
	"time"
)

func main() {

	ctx := context.TODO()

	storageProvider := postgres.New("host=localhost port=5555 user=postgres dbname=weather password=postgres sslmode=disable")
	storageProvider.MustRun()
	from, e := time.Parse(postgres.PG_TIME_FORMAT, "2024-09-17 02:12:00.000 +0000")

	d, _ := time.ParseDuration("3h")
	from = from.Round(d)

	if e != nil {
		panic(e)
	}
	q := []models.WKTRequestItem{
		{
			From: from,
			WKT:  "LINESTRING (40.95703125000001 57.78037554816888 , 40.38574218750001 56.13330691237569)",
		},
		{
			From: from,
			WKT:  "LINESTRING (40.38574218750001 56.13330691237569, 39.74853515625001 54.635697306063854)",
		},
	}

	r, err := storageProvider.GetForecastBySegments(ctx, q)

	if err != nil {
		panic(err)
	}

	for _, item := range r {
		fmt.Println(item)
	}
}
