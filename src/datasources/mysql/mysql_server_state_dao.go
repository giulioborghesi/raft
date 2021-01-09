package mysql_server_state_dao

import (
	"database/sql"
	"fmt"
)

const (
	// Statements to create schema and table
	createSchemaStatement = "CREATE DATABASE IF NOT EXISTS %s;"
	createTableStatement  = "CREATE TABLE IF NOT EXISTS %s.%s (" +
		"id INT PRIMARY KEY, term BIGINT, votedFor BIGINT);"

	// Statement to pre-fill table
	fillTableStatement = "INSERT IGNORE INTO %s.%s VALUES(0, 0, -1);"

	// Queries to search records
	getTermQuery     = "SELECT term FROM %s.%s WHERE id=0;"
	getVotedForQuery = "SELECT votedFor FROM %s.%s WHERE id=0;"

	// Queries to update records
	updateVotedForQuery     = "UPDATE %s.%s SET votedFor=? WHERE id=0;"
	updateTermVotedForQuery = "UPDATE %s.%s SET term=?, votedFor=? WHERE id=0;"
)

// mySqlServerStateDao implements the ServerStateDao interface. It is intended
// to be used internally by the package: exported implementation should build
// on top of mySqlServerStateDao
type mySqlServerStateDao struct {
	db          *sql.DB
	currentTerm int64
	votedFor    int64
	schemaName  string
	tableName   string
}

// executeStatement executes a SQL statement and returns the SQL result if
// successful, an error otherwise
func executeStatement(db *sql.DB, statement string) (sql.Result, error) {
	return db.Exec(statement)
}

// executeSearchQuery executes a SQL query and returns the SQL result if
// successful, an error otherwise
func executeSearchQuery(db *sql.DB, query string,
	args ...interface{}) (int64, error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var result int64
	row := stmt.QueryRow(args...)
	err = row.Scan(&result)
	return result, err
}

// executeUpdateQuery executes a SQL update query and returns the SQL result
// if successful, an error otherwise
func executeUpdateQuery(db *sql.DB, query string,
	args ...interface{}) (sql.Result, error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(args...)
}

func (d *mySqlServerStateDao) CurrentTerm() int64 {
	return d.currentTerm
}

func (d *mySqlServerStateDao) UpdateTerm(newTerm int64) error {
	return d.UpdateTermVotedFor(newTerm, -1)
}

func (d *mySqlServerStateDao) UpdateTermVotedFor(newTerm int64,
	votedFor int64) error {
	if newTerm <= d.currentTerm {
		panic("term number must increase monotonically")
	}

	query := fmt.Sprintf(updateTermVotedForQuery, d.schemaName, d.tableName)
	_, err := executeUpdateQuery(d.db, query, newTerm, votedFor)

	// Update in-memory state only if data was successfully persisted
	if err == nil {
		d.votedFor = votedFor
		d.currentTerm = newTerm
	}
	return err
}

func (d *mySqlServerStateDao) UpdateVotedFor(votedFor int64) error {
	if d.votedFor != -1 {
		panic("server has already casted a vote in the current term")
	}

	query := fmt.Sprintf(updateVotedForQuery, d.schemaName, d.tableName)
	_, err := executeUpdateQuery(d.db, query, votedFor)

	// Update in-memory state only if data was successfully persisted
	if err == nil {
		d.votedFor = votedFor
	}
	return err
}

func (d *mySqlServerStateDao) VotedFor() (int64, int64) {
	return d.currentTerm, d.votedFor
}
