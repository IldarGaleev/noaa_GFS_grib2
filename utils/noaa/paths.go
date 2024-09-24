package noaa

import "fmt"

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

func URLBuilder(model Model, day int, month int, year int, cycle ModelCycle, forecastTime int, gridSize GridSize) string {
	return fmt.Sprintf(
		"%[9]s/gfs.%[1]d%02[2]d%02[3]d/%[4]s/%[7]s/gfs.t%[4]sz.%[8]s.%[6]s.f%03[5]d",
		year,
		month,
		day,
		cycle,
		forecastTime,
		gridSize,
		model,
		"pgrb2",
		baseURL,
	)
}