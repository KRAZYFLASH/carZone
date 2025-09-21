package car

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KRAZYFLASH/SimpleApp-CarManagement/models"
	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) Store {
	return Store{db: db}
}

func (s Store) GetCarById(ctx context.Context, id string) (models.Car, error) {
	var car models.Car

	query := `
SELECT
  c.id, c.name, c.brand, c.year, c.fuel_type, c.price, c.created_at, c.updated_at,
  e.id, e.displacement, e.no_of_cylinders, e.car_range
FROM car c
LEFT JOIN engine e ON c.engine_id = e.id
WHERE c.id = $1
`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&car.ID, &car.Name, &car.Brand, &car.Year, &car.FuelType, &car.Price, &car.CreatedAt, &car.UpdatedAt,
		&car.Engine.EngineID, &car.Engine.Displacement, &car.Engine.NoOfCylinders, &car.Engine.CarRange,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// pilihan: balikan kosong tanpa error
			return models.Car{}, nil
		}
		return models.Car{}, err
	}
	return car, nil
}

func (s Store) GetCarByBrand(ctx context.Context, brand string, isEngine bool) ([]models.Car, error) {
	var (
		cars  []models.Car
		query string
	)

	if isEngine {
		query = `
SELECT
  c.id, c.name, c.brand, c.year, c.fuel_type, c.price, c.created_at, c.updated_at,
  e.id, e.displacement, e.no_of_cylinders, e.car_range
FROM car c
LEFT JOIN engine e ON c.engine_id = e.id
WHERE c.brand = $1
`
	} else {
		query = `
SELECT
  c.id, c.name, c.brand, c.year, c.fuel_type, c.price, c.created_at, c.updated_at
FROM car c
WHERE c.brand = $1
`
	}

	rows, err := s.db.QueryContext(ctx, query, brand)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Car
		if isEngine {
			if err := rows.Scan(
				&car.ID, &car.Name, &car.Brand, &car.Year, &car.FuelType, &car.Price, &car.CreatedAt, &car.UpdatedAt,
				&car.Engine.EngineID, &car.Engine.Displacement, &car.Engine.NoOfCylinders, &car.Engine.CarRange,
			); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(
				&car.ID, &car.Name, &car.Brand, &car.Year, &car.FuelType, &car.Price, &car.CreatedAt, &car.UpdatedAt,
			); err != nil {
				return nil, err
			}
		}
		cars = append(cars, car)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cars, nil
}

// CreateCar: insert engine + car dalam SATU transaksi.
func (s Store) CreateCar(ctx context.Context, carReq models.CarRequest) (models.Car, error) {
	var (
		createdCar models.Car
		engineID   uuid.UUID
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return createdCar, err
	}
	// Pastikan commit/rollback aman.
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// 1) Insert engine dulu, ambil id-nya
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO engine (displacement, no_of_cylinders, car_range)
         VALUES ($1, $2, $3) RETURNING id`,
		carReq.Engine.Displacement, carReq.Engine.NoOfCylinders, carReq.Engine.CarRange,
	).Scan(&engineID)
	if err != nil {
		return createdCar, err
	}

	// 2) Siapkan objek car
	carID := uuid.New()
	now := time.Now()

	// Kalau Year ada di CarRequest, pastikan diikutkan
	newCar := models.Car{
		ID:        carID,
		Name:      carReq.Name,
		Brand:     carReq.Brand,
		Year:      carReq.Year,      // <- penting, sebelumnya terlewat
		FuelType:  carReq.FuelType,
		Price:     carReq.Price,
		CreatedAt: now,
		UpdatedAt: now,
		Engine: models.Engine{
			EngineID:      engineID,
			Displacement:  carReq.Engine.Displacement,
			NoOfCylinders: carReq.Engine.NoOfCylinders,
			CarRange:      carReq.Engine.CarRange,
		},
	}

	// 3) Insert car + RETURNING kolom yang diperlukan
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO car (id, name, brand, year, fuel_type, engine_id, price, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING id, name, brand, year, fuel_type, price, created_at, updated_at`,
		newCar.ID, newCar.Name, newCar.Brand, newCar.Year, newCar.FuelType, engineID, newCar.Price, newCar.CreatedAt, newCar.UpdatedAt,
	).Scan(
		&createdCar.ID, &createdCar.Name, &createdCar.Brand, &createdCar.Year, &createdCar.FuelType, &createdCar.Price, &createdCar.CreatedAt, &createdCar.UpdatedAt,
	)
	if err != nil {
		return createdCar, err
	}

	// (opsional) ikutkan engine agar caller dapat paket lengkap
	createdCar.Engine = newCar.Engine

	return createdCar, nil
}

func (s Store) UpdateCar(ctx context.Context, carID string, carReq models.CarRequest) (models.Car, error) {
	var updatedCar models.Car

	// Tidak butuh transaksi jika hanya update satu tabel satu baris.
	err := s.db.QueryRowContext(
		ctx,
		`UPDATE car
         SET name = $1, brand = $2, year = $3, fuel_type = $4, price = $5, updated_at = $6
         WHERE id = $7
         RETURNING id, name, brand, year, fuel_type, price, created_at, updated_at`,
		carReq.Name, carReq.Brand, carReq.Year, carReq.FuelType, carReq.Price, time.Now(), carID,
	).Scan(
		&updatedCar.ID, &updatedCar.Name, &updatedCar.Brand, &updatedCar.Year, &updatedCar.FuelType, &updatedCar.Price, &updatedCar.CreatedAt, &updatedCar.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Car{}, errors.New("car not found")
		}
		return models.Car{}, err
	}
	return updatedCar, nil
}

func (s Store) DeleteCar(ctx context.Context, carID string) (models.Car, error) {
	var deletedCar models.Car

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Car{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Ambil dulu datanya untuk dikembalikan ke caller
	err = tx.QueryRowContext(
		ctx,
		`SELECT id, name, brand, year, fuel_type, price, created_at, updated_at
         FROM car WHERE id = $1`,
		carID,
	).Scan(
		&deletedCar.ID, &deletedCar.Name, &deletedCar.Brand, &deletedCar.Year, &deletedCar.FuelType, &deletedCar.Price, &deletedCar.CreatedAt, &deletedCar.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Car{}, errors.New("car not found")
		}
		return models.Car{}, err
	}

	// Hapus barisnya
	result, err := tx.ExecContext(ctx, "DELETE FROM car WHERE id = $1", carID)
	if err != nil {
		return models.Car{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return models.Car{}, err
	}
	if affected == 0 {
		return models.Car{}, errors.New("car not found")
	}

	return deletedCar, nil
}
