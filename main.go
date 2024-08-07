package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/nilsmagnus/grib/griblib"
)


const (
	baseUrl = "https://nomads.ncep.noaa.gov/pub/data/nccf/com/gfs/prod"
	
	// filtered url: https://nomads.ncep.noaa.gov/gribfilter.php?ds=fnl
	// url = "https://nomads.ncep.noaa.gov/cgi-bin/filter_fnl.pl?dir=%2Fgdas.20240807%2F00%2Fatmos&file=gdas.t00z.pgrb2.1p00.f000&var_PRES=on&var_TMP=on&lev_2_m_above_ground=on&lev_80_m_above_ground=on"
)

type Model string
type ModelCycle string
type GridSize string

const (
	Model_Atmo = Model("atmos")
	Model_Wave = Model("wave")
)

const (
	ModelCycle_00 = ModelCycle("00")
	ModelCycle_06 = ModelCycle("06")
	ModelCycle_12 = ModelCycle("12")
	ModelCycle_18 = ModelCycle("18")
)

const (
	GridSize_0p25 = GridSize("0p25")
	GridSize_0p50 = GridSize("0p50")
	GridSize_1p00 = GridSize("1p00")
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
		return errors.New("download fail")
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

func saveAs2dArray(data []float64, width uint32, fileName string) {
	dFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer dFile.Close()

	pointsCount := len(data)

	dFile.WriteString("[")

	col := 0
	for pI, dValue := range data {
		if col == 0 {
			dFile.WriteString("[")
		}
		dFile.WriteString(fmt.Sprintf("%f", dValue))
		col++
		if col == int(width) {
			col = 0
			if pI == pointsCount-1 {
				dFile.WriteString("]")
			} else {
				dFile.WriteString("],")
			}
		} else {
			dFile.WriteString(",")
		}
	}
	dFile.WriteString("]")
}

func processMessage(message *griblib.Message, id int) {
	def := message.Section3.Definition.(*griblib.Grid0)
	fmt.Printf("%[1]d - %[2]d(%[3]d x %[4]d)\n", id, message.Section5.PointsNumber, def.Ni, def.Nj)
	saveAs2dArray(message.Data(), def.Ni, fmt.Sprintf("prod/msg_%d.json", id))
}

func getMessages(fileName string)[]*griblib.Message{
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
		err := downloadFile(baseUrl+pathBuilder(Model_Atmo, 7, 8, 2024, ModelCycle_00, 0, GridSize_0p50), fileName)
		//err := downloadFile(url, fileName)
		if err != nil {
			panic(err)
		}
		fmt.Println("Download complete")
	} else {
		fmt.Println("File exists. Skip download")
	}

	// messages id: https://www.nco.ncep.noaa.gov/pmb/products/gfs/gfs.t00z.pgrb2.0p25.f000.shtml
	messages:=getMessages(fileName)

	pressure_sea_lvl := 0
	processMessage(messages[pressure_sea_lvl], pressure_sea_lvl)
	
	sunshine_duration  := 600 - 1
	processMessage(messages[sunshine_duration], sunshine_duration)

	temp_2_m_above_ground  := 580 - 1
	processMessage(messages[temp_2_m_above_ground], temp_2_m_above_ground)
	
	surface_mask := 682 - 1
	processMessage(messages[surface_mask], surface_mask)


	// for id, message := range messages {
	// 	processMessage(message, id)
	// }

}
