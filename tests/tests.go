package tests

import (
	"io/ioutil"
	"log"
	"testing"
)

//go:generate sh -c "cd api && go generate"

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
}
