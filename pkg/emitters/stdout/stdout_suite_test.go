package stdout_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStdout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stdout Suite")
}
