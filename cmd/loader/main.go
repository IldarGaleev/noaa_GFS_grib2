package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"os"
	"path/filepath"

	"gfsloader/internal/models"
	"gfsloader/internal/storage/postgres"
	"gfsloader/utils/indexfile"
	"gfsloader/utils/noaa"

	"github.com/nilsmagnus/grib/griblib"
	"github.com/schollz/progressbar/v3"
)

var (
	ErrProcess  = errors.New("process error")
	ErrDownload = errors.New("download error")
)

func download(label string, destinationPath, downloadURL string, from, to uint64) error {
	tempDestinationPath := destinationPath + ".tmp"
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	if from != 0 || to != 0 {
		end := ""
		if to != 0 {
			end = fmt.Sprint(to)
		}
		req.Header.Set("range", fmt.Sprintf("bytes=%d-%s", from, end))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 206:
		f, _ := os.OpenFile(tempDestinationPath, os.O_CREATE|os.O_WRONLY, 0644)

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			label,
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
		f.Close()
		os.Rename(tempDestinationPath, destinationPath)
		return nil
	default:
		return errors.Join(ErrDownload, fmt.Errorf("status code: %d : %s", resp.StatusCode, downloadURL))
	}

}

func getGribMessages(fileName string) ([]*griblib.Message, error) {
	gribfile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer gribfile.Close()

	messages, err := griblib.ReadMessages(gribfile)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

type message struct {
	param string
	layer string
}

type Records map[string]map[string]float64

func processData(day, month, year, forecastTime int, cycle noaa.ModelCycle, rate chan struct{}) (Records, error) {
	gribURL := noaa.URLBuilder(noaa.ModelAtmo, day, month, year, cycle, forecastTime, noaa.GridSize0p50)
	idxURL := gribURL + ".idx"

	fName := fmt.Sprintf("%d_%d_%d_%d", year, month, day, forecastTime)

	layers := []message{
		{
			param: "PRMSL",
			layer: "mean sea level",
		},
		{
			param: "LAND",
			layer: "surface",
		},
		{
			param: "TMP",
			layer: "2 m above ground",
		},
		{
			param: "UGRD",
			layer: "10 m above ground",
		},
		{
			param: "VGRD",
			layer: "10 m above ground",
		},
		{
			param: "RH",
			layer: "2 m above ground",
		},
		{
			param: "CRAIN",
			layer: "surface",
		},
		{
			param: "VIS",
			layer: "surface",
		},
	}

	layersCount := len(layers)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tempDir := filepath.Join(wd, "grib")

	err = os.Mkdir(tempDir, 0760)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	gribBaseFileName := filepath.Join(tempDir, fName)
	indexFileName := gribBaseFileName + ".idx"

	if _, err := os.Stat(indexFileName); errors.Is(err, os.ErrNotExist) {
		err := download("Get index file", indexFileName, idxURL, 0, 0)
		if err != nil {
			panic(err)
		}
	}

	idxFile, err := indexfile.New(indexFileName)

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	var errsLock sync.Mutex
	errs := make([]error, 0, len(layers))
	errs = append(errs, ErrProcess)
	putErr := func(e error) {
		defer errsLock.Unlock()
		errsLock.Lock()
		errs = append(errs, e)
	}

	var recordsLock sync.Mutex
	records := make(Records)
	putRecord := func(lat, lng float32, tag string, value float64) {
		defer recordsLock.Unlock()
		recordsLock.Lock()

		key := fmt.Sprintf("%+07d%07d%d", int(lat*100), int(lng*100), forecastTime)
		if r, ok := records[key]; ok {
			r[tag] = value
		} else {
			newR := make(map[string]float64, layersCount+2)
			newR["lat"] = float64(lat)
			newR["lng"] = float64(lng)
			newR["ftime"] = float64(forecastTime)
			newR[tag] = value
			records[key] = newR
		}
	}

	for _, layer := range layers {
		wg.Add(1)
		go func(paramName, layerName string) {

			defer wg.Done()

			layerGribFile := fmt.Sprintf("%s_%s_%s", gribBaseFileName, paramName, layerName)
			if _, err := os.Stat(layerGribFile); errors.Is(err, os.ErrNotExist) {
				from, to, err := idxFile.GetOffset(paramName, layerName)

				if err != nil {
					putErr(err)
					return
				}

				rate <- struct{}{}
				err = download(
					fmt.Sprintf("Get %s:%s", paramName, layerName),
					layerGribFile,
					gribURL,
					from,
					to,
				)
				<-rate

				if err != nil {
					putErr(err)
					return
				}
			}

			msgs, err := getGribMessages(layerGribFile)
			if err != nil {
				putErr(err)
				return
			}

			if len(msgs) == 1 {

				def := msgs[0].Section3.Definition.(*griblib.Grid0)
				data := msgs[0].Data()
				lngStep := float32(def.Di) / 1000000.0
				latStep := float32(def.Dj) / 1000000.0

				bar := progressbar.Default(int64(len(data)), fmt.Sprintf("Load forecast %d data", forecastTime))

				for j := uint32(0); j < def.Nj; j += 1 {
					for i := uint32(0); i < def.Ni; i += 1 {
						id := j*def.Ni + i

						lng := float32(i) * lngStep
						lat := 90.0 - float32(j)*latStep

						putRecord(lat, lng, paramName, data[id])
						bar.Add(1)

					}
				}
				bar.Close()
			} else {
				putErr(fmt.Errorf("\"%s-%s\" has wrong message count", paramName, layerName))
				return
			}

		}(layer.param, layer.layer)

	}

	wg.Wait()
	if len(errs) > 1 {
		return nil, errors.Join(errs...)
	}

	return records, nil

}

func main() {

	ctx := context.TODO()

	day := 29
	month := 9
	year := 2024
	cycle := noaa.ModelCycle06
	cycleInt, _ := strconv.Atoi(string(cycle))
	// forecastTime := 0
	gridSize := float32(0.5)

	maxConnections := 3
	rate := make(chan struct{}, maxConnections)

	var wg sync.WaitGroup

	storageProvider := postgres.New("host=localhost port=5555 user=postgres dbname=weather password=postgres sslmode=disable")
	storageProvider.MustRun()
	err := storageProvider.InitGrid(ctx, func() ([]models.GridInfo, func(int)) {
		totalCells := int(180.0 / gridSize * 360.0 / gridSize)
		result := make([]models.GridInfo, 0, totalCells)

		bar := progressbar.Default(int64(totalCells), "Create grid")

		for lat := float32(-90.0); lat <= 90; lat += gridSize {
			for lng := float32(0.0); lng <= 359.0; lng += gridSize {
				result = append(
					result,
					models.GridInfo{
						Lat:  lat,
						Lng:  lng,
						Size: gridSize,
					})
			}
		}
		return result, func(writed int) {
			bar.Add(writed)
		}
	})

	if err != nil {
		panic("failed to init database grid table")
	}

	for i := 0; i < 12; i += 3 {
		wg.Add(1)
		go func(f int) {
			defer wg.Done()
			records, err := processData(day, month, year, f, cycle, rate)
			if err != nil {
				panic(err)
			}

			transact, err := storageProvider.Begin(ctx)

			if err != nil {
				fmt.Println(err)
				return
			}

			bar := progressbar.Default(int64(len(records)), "Prepare data")

			dbRecords := make([]models.Record, 0, len(records))
			for _, record := range records {
				dateTime := time.Date(
					year,
					time.Month(month),
					day,
					cycleInt+int(record["ftime"]),
					0,
					0,
					0,
					time.UTC,
				)

				newRecord := models.Record{
					DateTime:    dateTime,
					Lat:         float32(record["lat"]),
					Lng:         float32(record["lng"]),
					IsGround:    record["LAND"] != 0,
					Pressure:    float32(record["PRMSL"]),
					Temperature: float32(record["TMP"]),
					UWind:       float32(record["UGRD"]),
					VWind:       float32(record["VGRD"]),
					CRain:       float32(record["CRAIN"]),
					RHUmidity:   float32(record["RH"]),
					Visibility:  float32(record["VIS"]),
				}
				dbRecords = append(dbRecords, newRecord)
				bar.Add(1)
				// fmt.Printf("%s : %v\n", key, record)
			}

			bar.Close()

			batchSize := 100
			rCount := len(records)
			dbBar := progressbar.Default(int64(rCount/batchSize), "Write to db")
			for i := 0; i < rCount-1; i += batchSize {
				err = transact.SetRecords(ctx, dbRecords[i:min(i+batchSize, rCount-1)])
				dbBar.Add(1)
				if err != nil {
					storageProvider.Rollback(ctx)
					return
				}
			}
			dbBar.Close()

			transact.Commit(ctx)

		}(i)
	}

	wg.Wait()

	storageProvider.Stop()

	// fmt.Println(records)

}
