package mempool_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sob-miner/internal/mempool"
	"sob-miner/internal/path"

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
		When("Tx Version 1 ", func() {
			txData, err := os.ReadFile(path.MempoolDataPath + "/" + "0a07736090b0677920c14d64e12e81cbb5e9d2fbcfeea536cda7d571b6d4607f.json")

			Expect(err).To(BeNil())
			tx := mempool.Transaction{}
			err = json.Unmarshal(txData, &tx)
			Expect(err).To(BeNil())

			It("should serialize properly", func() {
				serlizedTx, _, _, err := tx.Serialize()
				Expect(err).To(BeNil())
				Expect(hex.EncodeToString(serlizedTx)).To(Equal("0100000001218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a00000000"))
				// fmt.Println("serializedWitnessTx", hex.EncodeToString(serializedWitnessTx))
			})
			It("Should produce a valid Hash", func() {
				hash, wtxid, _, err := tx.Hash()
				Expect(err).To(BeNil())
				Expect(hash).To(Equal("d6ecc69adf1e2e54456052fec60b3cff123151df0bf2dc1bf19ad8113c25c6cc")) // little endian
				Expect(wtxid).To(Equal("b5f985857c82bc98e13819dc17a5eea47d98436e92b171cfb2d687b8e4268fad"))
			})
		})
		When("Tx Version 2 ", func() {
			txData, err := os.ReadFile(path.MempoolDataPath + "/" + "0a3fd98f8b3d89d2080489d75029ebaed0c8c631d061c2e9e90957a40e99eb4c.json")

			Expect(err).To(BeNil())
			tx := mempool.Transaction{}
			err = json.Unmarshal(txData, &tx)
			Expect(err).To(BeNil())

			It("should serialize properly", func() {
				serlizedTx, _, _, err := tx.Serialize()
				Expect(err).To(BeNil())
				Expect(hex.EncodeToString(serlizedTx)).To(Equal("020000000128945f452bcb038a679fc33fb03d06561a59585d6f36090894471b17b615b5b90000000000feffffff02e02c4300000000001600140d1c76c89fbba64867349c1ad0f3313e6b4b7d36eab539110000000016001414989c53e65d603069bf506996f24f45f4a121074dbc0c00"))
			})

			It("Should produce a valid Hash", func() {
				hash, _, _, err := tx.Hash()
				Expect(err).To(BeNil())
				Expect(hash).To(Equal("85482b8783e6d0f0a791000f0adda46a904ae69d2603044fa66754c262957d4f")) // little endian
			})
		})
	})
})

// 01000000000101218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a02470199769a581be794924501da96270bd626d0549603380f6051518a1cfa1017874a2002a668a2c8a6742e5dfad67c07bc76bf847a145a50881148bc7d31b600d1efbb742002443021cf74f81331fb8efe555239d63e9f9524f9e06af073a3f8f9ace9be50747e2ffc0200000000
// 01000000000101218883460f07350f167517c909b16b240bb738bb8ba493b1594f9facb45cfac60100000000ffffffff02546000000000000016001437fff1c9ce1d770cf82b38a1cdeba3972cddbb08a2ee8a00000000001600147ef8d1162a3f3691023a6fccb7723edd126ac80a02473044022074bbefd100b6317dbc481188505a147a84bf76bc077cd6fa5d2e74a6c8a268a602204a871710fa1c8a5151600f38039654d026d60b2796da01459294e71b589a7699012102fc2f7e7450bee9acf9f8a373f06ae0f924959f3ed6395255fe8efb3113f874cf00000000
