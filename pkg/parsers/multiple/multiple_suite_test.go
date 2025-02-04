package multiple_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMultiple(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multiple Suite")
}
