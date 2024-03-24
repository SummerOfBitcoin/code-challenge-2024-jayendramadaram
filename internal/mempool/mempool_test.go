package mempool_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sob-miner/internal/mempool"
	"sob-miner/internal/path"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	_, b, _, _      = runtime.Caller(0)
	Root            = filepath.Join(filepath.Dir(b), "../../")
	MempoolDataPath = filepath.Join(Root, "data", "mempool")
)

var _ = Describe("Mempool", func() {

	Expect(true).To(BeTrue())
	Context("Test Transaction Hash", func() {
		// When("Tx Version 1 ", func() {
		// 	Skip("Skipping for now")
		// 	txData, err := os.ReadFile(path.MempoolDataPath + "/" + "0a07736090b0677920c14d64e12e81cbb5e9d2fbcfeea536cda7d571b6d4607f.json")

		// 	Expect(err).To(BeNil())
		// 	tx := mempool.Transaction{}
		// 	err = json.Unmarshal(txData, &tx)
		// 	Expect(err).To(BeNil())

		// 	It("should serialize properly", func() {
		// 		serlizedTx, _, _, err := tx.Serialize()
		// 		Expect(err).To(BeNil())
		// 		Expect(hex.EncodeToString(serlizedTx)).To(Equal("0100000001218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a00000000"))
		// 		// fmt.Println("serializedWitnessTx", hex.EncodeToString(serializedWitnessTx))
		// 	})
		// 	It("Should produce a valid Hash", func() {
		// 		hash, wtxid, _, err := tx.Hash()
		// 		Expect(err).To(BeNil())
		// 		Expect(hash).To(Equal("d6ecc69adf1e2e54456052fec60b3cff123151df0bf2dc1bf19ad8113c25c6cc")) // little endian
		// 		Expect(wtxid).To(Equal("b5f985857c82bc98e13819dc17a5eea47d98436e92b171cfb2d687b8e4268fad"))
		// 	})
		// })
		When("Tx Version 2 ", Ordered, func() {
			var randFile string
			var txHash string
			var tx mempool.Transaction

			BeforeAll(func() {
				randFile = selectRandomFile(path.MempoolDataPath)
				txHash = strings.Split(randFile, " ")[0]
				txData, err := os.ReadFile(path.MempoolDataPath + "/" + randFile)

				Expect(err).To(BeNil())
				tx = mempool.Transaction{}
				err = json.Unmarshal(txData, &tx)
				Expect(err).To(BeNil())
			})

			It("should serialize properly", func() {
				fmt.Println("reading from file ", randFile)

				serlizedTx, _, _, _ := tx.Serialize()
				fmt.Println("serializedTx", hex.EncodeToString(serlizedTx))
				// Expect(err).To(BeNil())
				// Expect(hex.EncodeToString(serlizedTx)).To(Equal("020000000128945f452bcb038a679fc33fb03d06561a59585d6f36090894471b17b615b5b90000000000feffffff02e02c4300000000001600140d1c76c89fbba64867349c1ad0f3313e6b4b7d36eab539110000000016001414989c53e65d603069bf506996f24f45f4a121074dbc0c00"))
			})

			It("Should produce a valid Hash", func() {
				hash, _, _, _ := tx.Hash()
				fmt.Println("hash", hash)
				// Expect(err).To(BeNil())
				Expect(hash).To(Equal(reverseByteOrder(txHash))) // little endian
			})
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

func selectRandomFile(dir string) string {
	fmt.Println("selecting random file from ", dir)
	files, _ := os.ReadDir(dir)

	rand.Seed(int64(time.Now().Nanosecond()))
	randomIndex := rand.Intn(len(files))

	return files[randomIndex].Name()
}

// 01000000000101218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a02470199769a581be794924501da96270bd626d0549603380f6051518a1cfa1017874a2002a668a2c8a6742e5dfad67c07bc76bf847a145a50881148bc7d31b600d1efbb742002443021cf74f81331fb8efe555239d63e9f9524f9e06af073a3f8f9ace9be50747e2ffc0200000000
// 01000000000101218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a02473044022074bbefd100b6317dbc481188505a147a84bf76bc077cd6fa5d2e74a6c8a268a602204a871710fa1c8a5151600f38039654d026d60b2796da01459294e71b589a7699012102fc2f7e7450bee9acf9f8a373f06ae0f924959f3ed6395255fe8efb3113f874cf00000000

// 0200000002659a6eaf8d943ad2ff01ec8c79aaa7cb4f57002d49d9b8cf3c9a7974c5bd3608060000006bc485d31179138f66eea9c6cca5046217d58e366962245b65e61605d548736e0a032101095064b13796cf29008a50d673302b386f7db43bbaa773ea6b6323126c38e77d2002f16fecf68aa2ccfee45465b0880fa063fea25891a27a1f7c8e8cd320ac4e81f5002102453048fdffffff2cbc395e5c16b1204f1ced9c0d1699abf5abbbb6b2eee64425c55252131df6c40000000000fdffffff01878a03000000000017a914f043430ec4acf2cc3233309bbd1e43ae5efc81748700000000
// 0200000002659a6eaf8d943ad2ff01ec8c79aaa7cb4f57002d49d9b8cf3c9a7974c5bd3608060000006b483045022100f5814eac20d38c8e7c1f7aa29158a2fe63a00f88b06554e4fecca28af6ec6ff102207de7386c1223636bea73a7ba3bb47d6f382b3073d6508a0029cf9637b16450090121030a6e7348d50516e6655b246269368ed5176204a5ccc6a9ee668f137911d385c4fdffffff2cbc395e5c16b1204f1ced9c0d1699abf5abbbb6b2eee64425c55252131df6c40000000000fdffffff01878a03000000000017a914f043430ec4acf2cc3233309bbd1e43ae5efc81748700000000
