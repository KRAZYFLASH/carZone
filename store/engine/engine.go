package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KRAZYFLASH/carZone/models"
	"github.com/google/uuid"
)

type EngineStore struct {
	db *sql.DB
}

func New(db *sql.DB) *EngineStore {
	return &EngineStore{db: db}
}

func (e EngineStore) GetEngineById(ctx context.Context, id string) (models.Engine, error) {
	var engine models.Engine

	err := e.db.QueryRowContext(ctx, "SELECT id, displacement, no_of_cylinders, car_range FROM engine WHERE id = $1", id).Scan(
		&engine.EngineID, &engine.Displacement, &engine.NoOfCylinders, &engine.CarRange,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return engine, nil
		}
		return engine, err
	}
	return engine, nil
}

func (e EngineStore) EngineCreate(ctx context.Context, engineReq *models.EngineRequest) (models.Engine, error) {

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Engine{}, err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				fmt.Println("tx rollback error: ", rbErr)
			}
		} else {
			if cmErr := tx.Commit(); cmErr != nil {
				fmt.Println("tx commit error: ", cmErr)
				err = cmErr
			}
		}
	}()

	engineID := uuid.New()

	_, err = tx.ExecContext(ctx, "INSERT INTO engine (id, displacement, no_of_cylinders, car_range) VALUES ($1, $2, $3, $4)",
		engineID, engineReq.Displacement, engineReq.NoOfCylinders, engineReq.CarRange)
		
	if err != nil {
		return models.Engine{}, err
	}

	engine := models.Engine{
		EngineID:      engineID,
		Displacement:  engineReq.Displacement,
		NoOfCylinders: engineReq.NoOfCylinders,
		CarRange:      engineReq.CarRange,
	}

	return engine, nil
}

func (e EngineStore) EngineUpdate(ctx context.Context, id string, engineReq *models.EngineRequest) (models.Engine, error) {

	engineID, err := uuid.Parse(id)
	if err != nil {
		return models.Engine{}, fmt.Errorf("invalid UUID format: %w", err)
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Engine{}, err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				fmt.Println("tx rollback error: ", rbErr)
			}
		} else {
			if cmErr := tx.Commit(); cmErr != nil {
				fmt.Println("tx commit error: ", cmErr)
			}
		}
	}()

	results, err := tx.ExecContext(ctx, "UPDATE engine SET displacement = $1, no_of_cylinders = $2, car_range = $3 WHERE id = $4",
		engineReq.Displacement, engineReq.NoOfCylinders, engineReq.CarRange, engineID)

	if err != nil {
		return models.Engine{}, err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return models.Engine{}, err
	}
	if rowsAffected == 0 {
		return models.Engine{}, sql.ErrNoRows
	}

	updatedEngine := models.Engine{
		EngineID:      engineID,
		Displacement:  engineReq.Displacement,
		NoOfCylinders: engineReq.NoOfCylinders,
		CarRange:      engineReq.CarRange,
	}

	return updatedEngine, nil
}

func (e EngineStore) EngineDelete(ctx context.Context, id string) (models.Engine, error) {

	var engine models.Engine

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Engine{}, err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				fmt.Println("tx rollback error: ", rbErr)
			}
	}else {
			if cmErr := tx.Commit(); cmErr != nil {
				fmt.Println("tx commit error: ", cmErr)
			}
		}
	}()

	err = tx.QueryRowContext(ctx, "SELECT id, displacement, no_of_cylinders, car_range FROM engine WHERE id = $1", id).Scan(
		&engine.EngineID, &engine.Displacement, &engine.NoOfCylinders, &engine.CarRange,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return engine, nil
		}
		return engine, err
	}

	result, err := tx.ExecContext(ctx, "DELETE FROM engine WHERE id = $1", id)

	if err != nil {
		return models.Engine{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Engine{}, err
	}
	if rowsAffected == 0 {
		return models.Engine{}, errors.New("no rows were deleted")
	}

	return engine, nil
}