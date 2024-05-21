package db

import (
	"database/sql"
	"fmt"
	"time"

	// mysql driver needs comment because linter but this blank import is on purpose
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var db *sql.DB

// Init accepts authentication parameters for a mysql db and creates a client
// This function may also be configured to create tables in the db on behalf of the application for setup purposes.
func Init(createDB bool, host, database, user, passwd string) error {
	// Create the db if the create-db flag was given on the command line
	if createDB {
		// Create db client for creating the initial db -- this client doesn't specify a db in the DSN
		client, err := sql.Open("mysql", assembleDataSourceName(host, user, passwd, ""))
		if err != nil {
			return fmt.Errorf("creating database client: %v", err)
		}

		_, err = client.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, database))
		if err != nil {
			return fmt.Errorf("creating %s database (if it didn't exist): %v", database, err)
		}

		err = client.Close()
		if err != nil {
			log.Warnf("failed to close db client for creating the initial database, see original error message: %v", err)
		}
	}

	// Create db client -- this client includes the database name in the DSN so we don't need to select it in subsequent queries
	var err error
	db, err = sql.Open("mysql", assembleDataSourceName(host, user, passwd, database))
	if err != nil {
		return fmt.Errorf("creating database client: %v", err)
	}

	log.Debug("Creating table in mysql db if they don't already exist")

	err = initTable()
	if err != nil {
		return fmt.Errorf("creating creating punch table (if it didn't exist): %v", err)
	}

	log.Info("Finished initializing db package successfully")

	return nil
}

func assembleDataSourceName(host, user, passwd, database string) string {
	if database == "" {
		return fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true", user, passwd, host)
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, passwd, host, database)
}

func initTable() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mysqlpunch (
		id INT PRIMARY KEY AUTO_INCREMENT,
		text VARCHAR(512) NOT NULL,
		added_time DATETIME NOT NULL);`)
	return err
}

// Row represents a row in the mysqlpunch table
type Row struct {
	Text string
	Time time.Time
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(r Row) error {
	_, err := db.Exec(`INSERT INTO mysqlpunch (text,added_time) VALUES(?, ?);`, r.Text, r.Time)
	if err != nil {
		return fmt.Errorf("error adding row to mysqlpunch table: %v", err)
	}

	return nil
}

// ResetAllRecords this function would be called to reset the mysqlpunch table
// First deletes all rows in the table and then resets the auto_increment counter for the id column
func ResetAllRecords() error {
	_, err := db.Exec(`DELETE FROM mysqlpunch;`)
	if err != nil {
		return fmt.Errorf("error encountered deleting records for mysqlpunch table: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE mysqlpunch AUTO_INCREMENT = 1;`)
	if err != nil {
		return fmt.Errorf("error encountered reseting auto_increment for mysqlpunch table: %v", err)
	}
	return nil
}
