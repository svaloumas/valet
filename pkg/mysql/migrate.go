package mysql

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// CreateDBIfNotExists creates the database declared in the
// connection URL only if it does not already exist.
func (mySQL *MySQL) CreateDBIfNotExists() error {
	createDBStatement := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", mySQL.Config.DBName)

	host, port, _ := net.SplitHostPort(mySQL.Config.Addr)

	cmd := fmt.Sprintf("mysql -h%s -P%s -u%s -p'%s' -e '%s'",
		host, port, mySQL.Config.User, mySQL.Config.Passwd, createDBStatement)

	if mySQL.CaPemFile != "" {
		cmd = fmt.Sprintf("%s --ssl-ca=%s", cmd, mySQL.CaPemFile)
	}

	_, err := mySQL.runCommand("/bin/sh", "-c", cmd)
	if err != nil {
		return err
	}
	return nil
}

// Migrate applies all pending migrations.
func (mySQL *MySQL) Migrate(migrationsPath string) error {

	files, err := filepath.Glob(migrationsPath)
	if err != nil {
		mySQL.Log.Errorf("Error fetching migration files: %s", err)
		return err
	}

	sort.Strings(files)

	latest, err := mySQL.getLatestMigration()
	if err != nil {
		mySQL.Log.Errorf("Error selecting latest migration: %s", err)
		return err
	}

	completed, err := mySQL.applyPendingMigrations(files, latest)
	if err != nil {
		mySQL.Log.Errorf("Error applying pending migrations: %s", err)
		return err
	}

	if len(completed) > 0 {
		err := mySQL.setAppliedMigrations(completed)
		if err != nil {
			mySQL.Log.Errorf("Error inserting into migrations: %s", err)
		}

		mySQL.Log.Infof("Migrations completed up to %s\n\n", completed[len(completed)-1])

	} else {
		mySQL.Log.Infof("No migrations to perform at %s\n\n", migrationsPath)
	}

	return err
}

func (mySQL *MySQL) applyPendingMigrations(files []string, latest int) ([]string, error) {
	var completed []string

	for _, file := range files {
		filename := filepath.Base(file)
		version, _ := strconv.Atoi(strings.Split(filename, "_")[0])

		if version > latest {
			mySQL.Log.Infof("Running migration %s", filename)

			host, port, _ := net.SplitHostPort(mySQL.Config.Addr)

			cmd := fmt.Sprintf("mysql -h%s -P%s -u%s -p'%s'",
				host, port, mySQL.Config.User, mySQL.Config.Passwd)

			if mySQL.CaPemFile != "" {
				cmd = fmt.Sprintf("%s --ssl-ca=%s", cmd, mySQL.CaPemFile)
			}

			cmd = fmt.Sprintf("%s %s < %s", cmd, mySQL.Config.DBName, file)

			result, err := mySQL.runCommand("/bin/sh", "-c", cmd)
			if err != nil {
				return nil, err
			}

			completed = append(completed, filename)
			mySQL.Log.Infof("Completed migration %s\n%s\n", filename, result)
		}
	}

	return completed, nil
}

func (mySQL *MySQL) checkIfDBExists() bool {
	host, port, _ := net.SplitHostPort(mySQL.Config.Addr)

	cmd := fmt.Sprintf("mysql -h%s -P%s -u%s -p'%s' -e 'show databases;'",
		host, port, mySQL.Config.User, mySQL.Config.Passwd)

	if mySQL.CaPemFile != "" {
		cmd = fmt.Sprintf("%s --ssl-ca=%s", cmd, mySQL.CaPemFile)
	}

	commands := []string{
		cmd,
		"awk '{ print $1 }'",
		fmt.Sprintf("grep %s", mySQL.Config.DBName),
	}
	result, _ := mySQL.runCommand("/bin/sh", "-c", strings.Join(commands, " | "))
	result = strings.TrimSpace(result)

	return result == mySQL.Config.DBName
}

func (mySQL *MySQL) dropDB() error {
	dropDBStatement := fmt.Sprintf("DROP DATABASE %s;", mySQL.Config.DBName)

	host, port, _ := net.SplitHostPort(mySQL.Config.Addr)

	cmd := fmt.Sprintf("mysql -h%s -P%s -u%s -p'%s' -e '%s'",
		host, port, mySQL.Config.User, mySQL.Config.Passwd, dropDBStatement)

	if mySQL.CaPemFile != "" {
		cmd = fmt.Sprintf("%s --ssl-ca=%s", cmd, mySQL.CaPemFile)
	}

	_, err := mySQL.runCommand("/bin/sh", "-c", cmd)
	if err != nil {
		return err
	}

	return nil
}

func (mySQL *MySQL) getLatestMigration() (int, error) {

	sql := "SELECT version FROM migrations ORDER BY version DESC LIMIT 1;"

	rows, err := mySQL.DB.Query(sql)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return 0, nil
		}
		return -1, err
	}
	defer rows.Close()

	var version string
	latest := 0

	if rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return -1, err
		}

		latest, _ = strconv.Atoi(version)
	}

	return latest, nil
}

func (mySQL *MySQL) runCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	alreadyExists := strings.Contains(stderr.String(), "already exists")
	mysqlErr := "mysql error: " + stderr.String()

	if err != nil && !alreadyExists {
		return stdout.String(), errors.New(err.Error() + mysqlErr)
	}
	if stderr.String() != "" && !alreadyExists {
		return stdout.String(), errors.New(mysqlErr)
	}

	return stdout.String(), nil
}

func (mySQL *MySQL) setAppliedMigrations(migrations []string) error {

	for _, filename := range migrations {
		version, _ := strconv.Atoi(strings.Split(filename, "_")[0])
		sql := "INSERT INTO migrations (version) VALUES(?);"

		_, err := mySQL.DB.Exec(sql, version)
		if err != nil {
			return err
		}
	}

	return nil
}
