package fanout_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFanout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fanout Suite")
}
