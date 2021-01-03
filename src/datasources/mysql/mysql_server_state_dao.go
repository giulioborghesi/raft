package mysql_server_state_dao

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryGetTerm            = "SELECT value from ServerStatus where id=0;"
	queryGetVotedFor        = "SELECT value from ServerSatatus where id=1;"
	queryUpdateVotedFor     = "UPDATE ServerStatus SET value=? where id=1;"
	queryUpdateTermVotedFor = "UPDATE ServerStatus SET value=? where id=0, " +
		"SET value=? WHERE id=1;"
)

var (
	db  *sql.DB
	Dao *MySqlServerStateDao
)

func init() {
	// Fetch MySql parameters from environment variables
	usr := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOSTNAME")
	port := os.Getenv("MYSQL_HOSTPORT")
	dbn := os.Getenv("MYSQL_DB_NAME")

	// Initialize database
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", usr, pwd, host, port, dbn)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}

	// Setup database
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// Initialize server state DAO
	Dao = new(MySqlServerStateDao)
}

type MySqlServerStateDao struct {
	currentTerm int64
	votedFor    int64
}

func currentTerm() (int64, error) {
	stmt, err := db.Prepare(queryGetTerm)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var currentTerm int64
	row := stmt.QueryRow()
	err = row.Scan(&currentTerm)
	return currentTerm, err
}

func votedFor() (int64, error) {
	stmt, err := db.Prepare(queryGetVotedFor)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var votedFor int64
	row := stmt.QueryRow()
	err = row.Scan(&votedFor)
	return votedFor, err
}

// initialize is used to initialize the server state DAO object by reading the
// current term and the voted for variables from the database
func (s *MySqlServerStateDao) initialize() {
	currentTerm, err := currentTerm()
	if err != nil {
		panic(err)
	}

	votedFor, err := votedFor()
	if err != nil {
		panic(err)
	}

	s.currentTerm = currentTerm
	s.votedFor = votedFor
}

func (s *MySqlServerStateDao) CurrentTerm() int64 {
	return s.currentTerm
}

func (s *MySqlServerStateDao) UpdateTerm(newTerm int64) error {
	return s.UpdateTermVotedFor(newTerm, -1)
}

func (s *MySqlServerStateDao) UpdateTermVotedFor(newTerm int64,
	votedFor int64) error {
	if newTerm <= s.currentTerm {
		panic("term number must increase monotonically")
	}
	s.currentTerm = newTerm
	s.votedFor = votedFor

	stmt, err := db.Prepare(queryUpdateTermVotedFor)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(newTerm, votedFor)
	return err
}

func (s *MySqlServerStateDao) UpdateVotedFor(votedFor int64) error {
	if s.votedFor != -1 {
		panic("cannot vote for more than one server in the same term")
	}
	s.votedFor = votedFor

	stmt, err := db.Prepare(queryUpdateVotedFor)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(votedFor)
	return err
}

func (s *MySqlServerStateDao) VotedFor() (int64, int64) {
	return s.currentTerm, s.votedFor
}
