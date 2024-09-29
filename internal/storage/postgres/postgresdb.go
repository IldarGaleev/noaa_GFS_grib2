package postgres

import (
	"context"
	"errors"
	"fmt"
	"gfsloader/internal/storage"
	"strconv"
	"strings"

	appModels "gfsloader/internal/models"
	models "gfsloader/internal/storage/postgres/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	MAX_BATCH_SIZE = 100
	PG_TIME_FORMAT = "2006-01-02 15:04:05.000 -0700"
)

type PostgresDataProvider struct {
	dsn        string
	db         *gorm.DB
	gridIsInit bool
}

// New create DatabaseApp
func New(dsn string) *PostgresDataProvider {
	return &PostgresDataProvider{
		dsn: dsn,
	}
}

// MustRun create postgres database connection. Panic if failed
func (d *PostgresDataProvider) MustRun() {
	err := d.Run()
	if err != nil {
		panic(err)
	}
}

// Run create postgres database connection
func (d *PostgresDataProvider) runWithDialector(dialector gorm.Dialector) error {
	db, err := gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return errors.Join(storage.ErrDatabaseError, err)
	}
	d.db = db

	recordsTable := &models.PGRecord{}
	gridTable := &models.PGGridInfo{}

	d.gridIsInit = db.Migrator().HasTable(gridTable)

	err = db.AutoMigrate(
		recordsTable,
		gridTable,
	)

	if err != nil {
		return errors.Join(storage.ErrDatabaseError, err)
	}

	return nil
}

func (d *PostgresDataProvider) Run() error {
	return d.runWithDialector(postgres.Open(d.dsn))
}

// Stop close postgres database connection
func (d *PostgresDataProvider) Stop() error {
	if d.db == nil {
		return storage.ErrDatabaseError
	}

	conn, err := d.db.DB()

	if err != nil {
		return errors.Join(storage.ErrDatabaseError, err)
	}

	err = conn.Close()
	if err != nil {
		return errors.Join(storage.ErrDatabaseError, err)
	}

	return nil
}

func (d *PostgresDataProvider) Begin(ctx context.Context) (*PostgresDataProvider, error) {
	r := d.db.WithContext(ctx).Statement.Begin()
	if r.Error != nil {
		return nil, errors.Join(storage.ErrDatabaseError, r.Error)
	}

	r = r.Exec(`set transaction isolation level REPEATABLE READ`)
	if r.Error != nil {
		return nil, errors.Join(storage.ErrDatabaseError, r.Error)
	}
	return &PostgresDataProvider{
		db: r,
	}, nil
}

func (d *PostgresDataProvider) Commit(ctx context.Context) error {
	r := d.db.WithContext(ctx).Statement.Commit()
	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}
	return nil
}

func (d *PostgresDataProvider) Rollback(ctx context.Context) error {
	r := d.db.WithContext(ctx).Statement.Rollback()
	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}
	return nil
}

func (d *PostgresDataProvider) AddRecord(ctx context.Context, record appModels.Record) error {

	idx := coordToIndex(record.Lat, record.Lng)
	dbRecord := models.PGRecord{
		GridID: idx,
	}

	r := d.db.WithContext(ctx).Create(&dbRecord)

	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}

	return nil
}

func coordToIndex(lat, lng float32) int64 {
	idx, _ := strconv.Atoi(fmt.Sprintf("%d%06d", int(lat*100), int(lng*100)))
	return int64(idx)
}

// SetRecords create or update records
func (d *PostgresDataProvider) SetRecords(ctx context.Context, records []appModels.Record) error {
	if len(records) > MAX_BATCH_SIZE {
		return storage.ErrBatchSize
	}

	data := make([]models.PGRecord, 0, len(records))
	for _, record := range records {
		idx := coordToIndex(record.Lat, record.Lng)
		dbRecord := models.PGRecord{
			GridID:      idx,
			DateTime:    record.DateTime,
			IsGround:    record.IsGround,
			Pressure:    record.Pressure,
			Temperature: record.Temperature,
			UWind:       record.UWind,
			VWind:       record.VWind,
			CRain:       record.CRain,
			RHumidity:   record.RHUmidity,
			Visibility:  record.Visibility,
		}

		data = append(data, dbRecord)
	}

	r := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "grid_id"},
			{Name: "date_time"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"temperature",
			"pressure",
			"u_wind",
			"v_wind",
			"c_rain",
			"r_humidity",
		}),
	}).CreateInBatches(data, len(data))

	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}

	return nil
}

type GridGenerate func() ([]appModels.GridInfo, func(writed int))

func (d *PostgresDataProvider) InitGrid(ctx context.Context, gridGenerate GridGenerate) (err error) {
	tbl := &models.PGGridInfo{}
	dbCtx := d.db.WithContext(ctx)

	defer func() {

		if err != nil {
			dbCtx.Migrator().DropTable(tbl)
		}
		dbCtx.Callback().Create().After("gorm:create").Remove("create_grid")
	}()

	if !d.gridIsInit {

		grid, clbk := gridGenerate()
		err = d.db.Callback().Create().After("gorm:create").Register("create_grid", func(d *gorm.DB) {
			if d.Statement.Table == tbl.TableName() {
				clbk(MAX_BATCH_SIZE)
			}
		})

		if err != nil {
			return errors.Join(storage.ErrDatabaseError, err)
		}

		gridCount := len(grid)
		dbRecords := make([]models.PGGridInfo, 0, MAX_BATCH_SIZE)

		for i := 0; i < gridCount; i += MAX_BATCH_SIZE {
			gridBatch := grid[i:min(i+MAX_BATCH_SIZE, gridCount-1)]
			dbRecords = dbRecords[:0]
			db, err := d.Begin(ctx)
			if err != nil {
				return errors.Join(storage.ErrDatabaseError, err)
			}
			for _, record := range gridBatch {

				idx := coordToIndex(record.Lat, record.Lng)
				hSz := record.Size / 2
				dbRecord := models.PGGridInfo{
					ID: idx,
					Geometry: models.GISRectangle{
						Y1: float64(record.Lat - hSz),
						X1: float64(record.Lng - hSz),
						Y2: float64(record.Lat + hSz),
						X2: float64(record.Lng + hSz),
					},
				}

				dbRecords = append(dbRecords, dbRecord)
			}

			err = db.db.CreateInBatches(dbRecords, len(dbRecords)).Error
			if err != nil {
				return errors.Join(storage.ErrDatabaseError, err)
			}
			err = db.Commit(ctx)
			if err != nil {
				rErr := db.Rollback(ctx)
				return errors.Join(storage.ErrDatabaseError, err, rErr)
			}
		}
	}

	return nil
}

func toSQLValueItem(items []appModels.WKTRequestItem) (string, []interface{}) {
	result := make([]string, 0, len(items))
	vals := make([]interface{}, 0, len(items))
	for i, item := range items {
		if i == 0 {
			result = append(result, "(?::timestamptz,?::timestamptz,st_geomfromtext(?,4326))")
		} else {
			result = append(result, "(?,?,st_geomfromtext(?,4326))")
		}
		to := item.From
		if item.To != nil {
			to = *item.To
		}
		vals = append(
			vals,
			item.From.Format(PG_TIME_FORMAT),
			to.Format(PG_TIME_FORMAT),
			item.WKT,
		)
	}

	return strings.Join(result, ","), vals
}

func (d *PostgresDataProvider) GetForecastBySegments(ctx context.Context, segments []appModels.WKTRequestItem) ([]models.PGResponse, error) {

	db := d.db.WithContext(ctx)

	var result []models.PGResponse
	valsSQL, data := toSQLValueItem(segments)
	err := db.Raw(fmt.Sprintf("WITH "+
		"q(f,t,geo) AS (VALUES %s),"+
		"cells AS (SELECT g.geometry AS geo, g.id AS p, q.f as f, q.t as t,ST_Intersection(g.geometry,q.geo) AS s FROM grid g JOIN q ON ST_Intersects(g.geometry ,q.geo))"+
		"SELECT st_astext(c.s) AS sec,r.temperature - 273.15 AS temperature, r.pressure AS pressure,r.c_rain AS c_rain,r.r_humidity AS r_humidity,r.date_time AT TIME ZONE 'UTC' AS date FROM records r JOIN cells c ON c.p=r.grid_id AND r.date_time BETWEEN c.f AND c.t",
		valsSQL,
	),
		data...,
	).Scan(&result).Error

	if err != nil {
		return nil, errors.Join(storage.ErrDatabaseError, err)
	}

	return result, nil
}
