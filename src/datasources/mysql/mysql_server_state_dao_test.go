package mysql_datasources

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// Names of database and table used for unit testing
	testSchemaName = "test_schema"
	testTableName  = "test_table"

	// Statement to drop test database
	dropSchemaStatement = "DROP DATABASE IF EXISTS %s;"
)

// createDao initializes the MySql Dao object
func createDao() *mySqlServerStateDao {
	user := os.Getenv("MYSQL_USERNAME")
	pswd := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOSTNAME")
	port := os.Getenv("MYSQL_HOSTPORT")

	// Open connection to database
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, pswd, host, port)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}

	// Setup database connection parameters
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// Initialize test database: drop existing test database first
	if _, err = db.Exec(fmt.Sprintf(dropSchemaStatement,
		testSchemaName)); err != nil {
		panic(err)
	}

	// Create empty test database
	if _, err = db.Exec(fmt.Sprintf(createSchemaStatement,
		testSchemaName)); err != nil {
		panic(err)
	}

	// Create test table in test database
	if _, err = db.Exec(fmt.Sprintf(createTableStatement, testSchemaName,
		testTableName)); err != nil {
		panic(err)
	}

	// Create record in test table
	if _, err = db.Exec(fmt.Sprintf(fillTableStatement, testSchemaName,
		testTableName)); err != nil {
		panic(err)
	}

	// Create dao object
	return &mySqlServerStateDao{db: db, currentTerm: 0, votedFor: -1,
		schemaName: testSchemaName, tableName: testTableName}
}

func TestExecuteQuery(t *testing.T) {
	// Create DAO object
	dao := createDao()

	// Fetch value of current term from database
	query := fmt.Sprintf(getTermQuery, dao.schemaName, dao.tableName)
	dbTerm, err := executeSearchQuery(dao.db, query)
	if err != nil {
		t.Fatal("executeSearchQuery should not return an error")
	}

	if dbTerm != 0 {
		t.Fatalf("executeSearchQuery returned incorrect term: "+
			"expected: 0, actual: %d", dbTerm)
	}

	// Update term value on database
	var newTerm int64 = 5
	query = fmt.Sprintf(updateTermVotedForQuery, dao.schemaName, dao.tableName)
	sqlResult, err := executeUpdateQuery(dao.db, query, newTerm, -1)
	if err != nil {
		t.Fatal("executeUpdateQuery should not return an error")
	}

	nRowsChanged, err := sqlResult.RowsAffected()
	if err == nil {
		if nRowsChanged != 1 {
			t.Fatal("executeUpdateQuery changes an incorrect number of rows: "+
				"expected: 1, actual: %d", nRowsChanged)
		}
	}

	// Fetch again value of current term from database
	query = fmt.Sprintf(getTermQuery, dao.schemaName, dao.tableName)
	dbTerm, err = executeSearchQuery(dao.db, query)
	if err != nil {
		t.Fatal("executeSearchQuery should not return an error")
	}

	if dbTerm != newTerm {
		t.Fatalf("executeSearchQuery returned incorrect term: "+
			"expected: %d, actual: %d", newTerm, dbTerm)
	}
}

func TestUpdateVotedFor(t *testing.T) {
	// Create DAO object
	dao := createDao()

	// Update votedFor value stored on database
	var newVotedFor int64 = 1
	err := dao.UpdateVotedFor(newVotedFor)
	if err != nil {
		t.Fatal("UpdateVotedFor should not return an error")
	}

	// In-memory value for votedFor should be equal to newVotedFor
	if _, votedFor := dao.VotedFor(); votedFor != newVotedFor {
		t.Fatalf("incorrect value of votedFor after UpdateVotedFor: "+
			"expected: %d, actual: %d", newVotedFor, votedFor)
	}

	// Stored value for votedFor must be equal to in-memory value
	query := fmt.Sprintf(getVotedForQuery, dao.schemaName, dao.tableName)
	dbVotedFor, err := executeSearchQuery(dao.db, query)
	if err != nil {
		t.Fatal("executeSearchQuery should not return an error")
	}

	if _, votedFor := dao.VotedFor(); votedFor != dbVotedFor {
		t.Fatalf("mismatch between values of votedFor stored in DAO object "+
			"and database: expected: %d, actual: %d", dbVotedFor, votedFor)
	}
}

func TestUpdateTerm(t *testing.T) {
	// Create DAO object
	dao := createDao()

	// Update term and votedFor values stored on database
	var newTerm int64 = 4
	err := dao.UpdateTerm(newTerm)
	if err != nil {
		t.Fatal("UpdateTerm should not return an error")
	}

	// In-memory value for currentTerm should be equal to newTerm
	if dao.CurrentTerm() != newTerm {
		t.Fatalf("incorrect value of currentTerm after UpdateVotedFor: "+
			"expected: %d, actual: %d", newTerm, dao.CurrentTerm())
	}

	// In-memory value for votedFor should be equal to -1
	if _, votedFor := dao.VotedFor(); votedFor != -1 {
		t.Fatalf("incorrect value of currentTerm after UpdateVotedFor: "+
			"expected: %d, actual: %d", -1, votedFor)
	}

	// Stored value for votedFor must be equal to in-memory value
	query := fmt.Sprintf(getVotedForQuery, dao.schemaName, dao.tableName)
	dbVotedFor, err := executeSearchQuery(dao.db, query)
	if err != nil {
		t.Fatal("executeSearchQuery should not return an error")
	}

	if _, votedFor := dao.VotedFor(); dbVotedFor != votedFor {
		t.Fatalf("mismatch between values of votedFor stored in DAO object "+
			"and database: expected: %d, actual: %d", dbVotedFor, votedFor)
	}
}
