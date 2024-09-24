package postgres

import (
	"context"
	"errors"
	"fmt"
	"gfsloader/utils/storage"
	"strconv"

	appModels "gfsloader/utils/models"
	models "gfsloader/utils/storage/postgres/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDataProvider struct {
	dsn string
	db  *gorm.DB
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
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return errors.Join(storage.ErrDatabaseError, err)
	}
	d.db = db

	err = db.AutoMigrate(
		&models.PGRecord{},
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
		return nil, storage.ErrDatabaseError
	}

	r = r.Exec(`set transaction isolation level REPEATABLE READ`)
	if r.Error != nil {
		return nil, storage.ErrDatabaseError
	}
	return &PostgresDataProvider{
		db: r,
	}, nil
}

func (d *PostgresDataProvider) Commit(ctx context.Context) error {
	r := d.db.WithContext(ctx).Statement.Commit()
	if r.Error != nil {
		return storage.ErrDatabaseError
	}
	return nil
}

func (d *PostgresDataProvider) Rollback(ctx context.Context) error {
	r := d.db.WithContext(ctx).Statement.Rollback()
	if r.Error != nil {
		return storage.ErrDatabaseError
	}
	return nil
}

func (d *PostgresDataProvider) AddRecord(ctx context.Context, record appModels.Record) error {

	idx, _ := strconv.Atoi(fmt.Sprintf("%d%06d", int(record.Lat*100), int(record.Lng*100)))
	dbRecord := models.PGRecord{
		Geo: int64(idx),
	}

	r := d.db.WithContext(ctx).Create(&dbRecord)

	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}

	return nil
}

func (d *PostgresDataProvider) AddRecords(ctx context.Context, records []appModels.Record) error {

	data := make([]models.PGRecord, 0, len(records))
	for _, record := range records {
		idx, _ := strconv.Atoi(fmt.Sprintf("%d%06d", int(record.Lat*100), int(record.Lng*100)))
		dbRecord := models.PGRecord{
			Geo:         int64(idx),
			DateTime:    record.DateTime,
			IsGround:    record.IsGround,
			Pressure:    record.Pressure,
			Temperature: record.Temperature,
			UWind:       record.UWind,
			VWind:       record.VWind,
		}

		data = append(data, dbRecord)

	}

	r := d.db.WithContext(ctx).CreateInBatches(data, len(data))

	if r.Error != nil {
		return errors.Join(storage.ErrDatabaseError, r.Error)
	}

	return nil
}
