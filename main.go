package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/nilsmagnus/grib/griblib"
)

// GFS doc: https://www.emc.ncep.noaa.gov/emc/pages/numerical_forecast_systems/gfs/documentation.php

const (
	baseURL = "https://nomads.ncep.noaa.gov/pub/data/nccf/com/gfs/prod"

	// filtered url: https://nomads.ncep.noaa.gov/gribfilter.php?ds=fnl
	// url = "https://nomads.ncep.noaa.gov/cgi-bin/filter_fnl.pl?dir=%2Fgdas.20240807%2F00%2Fatmos&file=gdas.t00z.pgrb2.1p00.f000&var_PRES=on&var_TMP=on&lev_2_m_above_ground=on&lev_80_m_above_ground=on"
)

type Model string
type ModelCycle string
type GridSize string

const (
	ModelAtmo = Model("atmos")
	ModelWave = Model("wave")
)

const (
	ModelCycle00 = ModelCycle("00")
	ModelCycle06 = ModelCycle("06")
	ModelCycle12 = ModelCycle("12")
	ModelCycle18 = ModelCycle("18")
)

const (
	GridSize0p25 = GridSize("0p25")
	GridSize0p50 = GridSize("0p50")
	GridSize1p00 = GridSize("1p00")
)

func pathBuilder(model Model, day int, month int, year int, cycle ModelCycle, forecastTime int, gridSize GridSize) string {
	return fmt.Sprintf(
		"/gfs.%[1]d%02[2]d%02[3]d/%[4]s/%[7]s/gfs.t%[4]sz.%[8]s.%[6]s.f%03[5]d",
		year,
		month,
		day,
		cycle,
		forecastTime,
		gridSize,
		model,
		"pgrb2",
	)
}

func downloadFile(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("download fail, code: %d", response.StatusCode)
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func saveAs2dArray(data []float64, col, row uint32, fileName string) {
	dFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer dFile.Close()

	resultData := make([][]float64, 0, row)
	totalLen := uint32(len(data))
	for i := uint32(0); i < row; i++ {
		from := i * col
		if from > totalLen {
			resultData = append(resultData, make([]float64, col))
			continue
		}
		to := from + min(totalLen-from, col)
		rowData := data[from:to]
		rowDataLen := uint32(len(rowData))
		if rowDataLen < col {
			rowData = append(rowData, make([]float64, col-rowDataLen)...)
		}
		resultData = append(resultData, rowData)
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		panic(err)
	}
	dFile.Write(jsonData)
}

func processMessage(message *griblib.Message, id int) {
	def := message.Section3.Definition.(*griblib.Grid0)
	fmt.Printf("%[1]d - %[2]d(%[3]d x %[4]d)\n", id, message.Section5.PointsNumber, def.Ni, def.Nj)
	saveAs2dArray(message.Data(), def.Ni, def.Nj, fmt.Sprintf("prod/msg_%d.json", id))
}

func getMessages(fileName string) []*griblib.Message {
	gribfile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer gribfile.Close()

	messages, err := griblib.ReadMessages(gribfile)
	if err != nil {
		panic(err)
	}

	return messages
}

func main() {

	fileName := "./tmp/data.grb2"

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Start downloading...")
		err := downloadFile(baseURL+pathBuilder(ModelAtmo, 8, 8, 2024, ModelCycle06, 0, GridSize1p00), fileName)
		//err := downloadFile(url, fileName)
		if err != nil {
			panic(err)
		}
		fmt.Println("Download complete")
	} else {
		fmt.Println("File exists. Skip download")
	}

	// messages id: https://www.nco.ncep.noaa.gov/pmb/products/gfs/gfs.t00z.pgrb2.0p25.f000.shtml
	fmt.Println("Read messages...")
	messages := getMessages(fileName)
	fmt.Println("Extract data")

	var wg sync.WaitGroup

	layers := []int{
		1,   //pressure_sea_lvl
		580, //temp_2_m_above_ground
		600, //sunshine_duration
		682, //surface_mask
	}

	for _, layer := range layers {
		wg.Add(1)
		go func(msg []*griblib.Message, lId int) {
			defer wg.Done()
			processMessage(msg[lId], lId)
		}(messages, layer-1)
	}

	wg.Wait()

	fmt.Print("Done!")

	// fmt.Print(len(messages))
	// for id, message := range messages {
	// 	processMessage(message, id)
	// }

}
