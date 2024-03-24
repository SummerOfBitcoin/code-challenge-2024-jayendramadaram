package miner_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMiner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Miner Suite")
}
