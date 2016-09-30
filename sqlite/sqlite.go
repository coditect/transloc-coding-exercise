package sqlite

import (
	"database/sql"

	"github.com/coditect/transloc-coding-exercise/model"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

func New(file string) (*Database, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "locations" ("latitude" REAL, "longitude" REAL, "addresses" INTEGER, PRIMARY KEY ("latitude", "longitude"))`)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (db *Database) Query(north, south, east, west float64) (model.LocationTable, error) {
	result := make(model.LocationTable)
	rows, err := db.DB.Query("SELECT latitude, longitude, addresses FROM locations WHERE latitude BETWEEN ? AND ? AND longitude > ? AND longitude <= ?", south, north, west, east)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var latitude, longitude, quantity float64
		err := rows.Scan(&latitude, &longitude, &quantity)
		if err != nil {
			return nil, err
		}

		location := model.Location{latitude, longitude}
		result[location] = quantity
	}

	return result, nil
}

func (db *Database) Save(locations model.LocationTable) error {
	trans, err := db.Begin()
	if err != nil {
		return err
	}

	var success bool
	defer func() {
		if !success {
			trans.Rollback()
		}
	}()

	// Truncate the table
	_, err = trans.Exec("DELETE FROM locations")
	if err != nil {
		return err
	}

	// Initialize prepared statement
	stmt, err := trans.Prepare("INSERT INTO locations (latitude, longitude, addresses) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert values
	for location, quantity := range locations {
		_, err := stmt.Exec(location.Latitude, location.Longitude, quantity)
		if err != nil {
			return err
		}
	}

	err = trans.Commit()
	if err != nil {
		return err
	}

	success = true;
	return nil
}
