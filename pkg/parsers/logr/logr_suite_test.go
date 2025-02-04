package logr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logr Suite")
}
