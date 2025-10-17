package tests

import (
	"github.com/goravel/framework/testing"

	"aura/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
