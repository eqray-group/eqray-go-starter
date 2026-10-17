package tests

import (
	"github.com/goravel/framework/testing"

	"eqray-go-starter/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
