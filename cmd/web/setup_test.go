package main

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("Start testing package: main")
	os.Exit(m.Run())
}
