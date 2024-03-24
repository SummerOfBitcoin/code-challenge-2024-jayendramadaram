package miner_test

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"sob-miner/internal/miner"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Miner", func() {
	Context("Test Utils", func() {
		It("Test Hash256", func() {
			Expect(miner.Hash256("0000000000000000000000000000000000000000000000000000000000000000")).To(Equal("2b32db6c2c0a6235fb1397e8225ea85e0f0e6e8c7b126d0016ccbde0e667151e"))
		})

		It("Test GenerateMerkleRoot", func() {
			file, err := os.Open("../../output.txt")
			Expect(err).To(BeNil())
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var txids []string

			for i := 0; i < 2; i++ {
				if scanner.Scan() {
					continue
				}
				if err := scanner.Err(); err != nil {
					fmt.Println("Error reading file:", err)
					return
				}
			}

			for scanner.Scan() {
				txids = append(txids, reverseByteOrder(scanner.Text()))
			}

			fmt.Println("")
			merkleRoot := miner.GenerateMerkleRoot(txids)
			fmt.Println("merkleRoot", merkleRoot)
		})
	})
})

func reverseByteOrder(hash string) string {
	reverse, _ := hex.DecodeString(hash)
	for i, j := 0, len(reverse)-1; i < j; i, j = i+1, j-1 {
		reverse[i], reverse[j] = reverse[j], reverse[i]
	}
	return hex.EncodeToString(reverse)
}
