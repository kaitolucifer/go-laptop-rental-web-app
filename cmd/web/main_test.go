package main

import (
	"flag"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestRun(t *testing.T) {
	flag.CommandLine.Set("production", "false")
	flag.CommandLine.Set("cache", "false")

	err := godotenv.Load("../../.env")
	if err != nil {
		t.Error("Error loading .env file")
	}

	dbHost = os.Getenv("dbhost")
	dbName = os.Getenv("dbname")
	dbUser = os.Getenv("dbuser")
	dbPassword = os.Getenv("dbpassword")
	dbPort = os.Getenv("dbport")
	dbSSL = os.Getenv("dbssl") // (disbale, prefer, require)

	_, err = run()
	if err != nil {
		t.Errorf("failed run(): %s", err)
	}
}
