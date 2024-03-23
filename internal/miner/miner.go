package miner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	config "sob-miner"
	"sob-miner/internal/ierrors"
	"sob-miner/internal/mempool"
	"sob-miner/internal/path"
	"sob-miner/pkg/block"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type miner struct {
	block   *block.Block
	mempool mempool.Mempool

	logger         *logrus.Logger
	rejectedTxFile *os.File

	maxBlockSize uint
}

func New(mempool mempool.Mempool, opts Opts) (*miner, error) {
	file, err := os.OpenFile("../rejected_txs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &miner{
		block: &block.Block{},

		logger:       opts.Logger,
		maxBlockSize: opts.MaxBlockSize,
		mempool:      mempool,

		rejectedTxFile: file,
	}, nil
}

// strategy
// pick best tx from mempool [most fee] ✅:
//   - this sorts out RBF and CPFP ✅
//
// do sanity checks on tx
//   - fetch inputs ✅
//   - check is inputs are already spent [if spent reason might RBF or double spending] reject tx [delete from mempool] ✅
//   - fetch and outputs and do sanity checks on inputs and outputs
//   - now do cryptographic checks [signatures and encodings]
//   - verify scripts
//   - if seems ok then push txId into block [we only need txID for this assignment we can flush inputs and outputs]
//
// build coinbase tx from fee collected + witness-commitment
// build block Header
// save block to output.txt
func (m *miner) Mine() error {
	weight := 0
	feeCollected := 0
	wTxids := []string{}

	GivenDifficulty := HexMustDecode("0000ffff00000000000000000000000000000000000000000000000000000000")

	//TODO: use logrus file than file writing
PICK_TX:
	for weight < config.MAX_BLOCK_SIZE {
		tx, err := m.mempool.PickBestTx()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				m.logger.Info("mempool is empty")
				break PICK_TX
			}
			return err
		}

		fmt.Printf("\rProcessing... tx: %s with ID: %d", tx.Hash, tx.ID)

		inputs, err := m.mempool.GetInputs(tx.Hash)
		if err != nil {
			m.logger.Info("unable to get inputs", err)
			return err
		}

		// signature checks and stack execution
		for _, input := range inputs {
			if err := m.mempool.MarkOutPointSpent(input.FundingTxHash, input.FundingIndex); err != nil {
				if errors.Is(err, ierrors.ErrAlreadySpent) {
					m.logger.Info("already spent")
					m.rejectedTxFile.WriteString(tx.Hash + " Reason: Already spent" + "\n")
					continue PICK_TX
				}
			}

			if err := m.mempool.ValidateInput(input); err != nil {
				m.logger.Info("invalid input", err)
				m.rejectedTxFile.WriteString(tx.Hash + " Reason: " + err.Error() + "\n")
				continue PICK_TX
			}
		}

		if err := m.mempool.DeleteTx(tx.ID); err != nil {
			m.logger.Info("unable to delete tx", err)
			return err
		}

		m.logger.Info("tx accepted", tx.Hash)

		// include tx in block
		weight += int(tx.Weight)
		feeCollected += int(tx.FeeCollected)

		m.block.Txs = append(m.block.Txs, tx.Hash) // hash is in LittleEndian
		wTxids = append(wTxids, tx.WTXID)          // wTxid is in LittleEndian
	}

	// build CoinBase Tx
	// - has one input ✅
	// - - in hash  and witness  = bytes32(0x0) ✅
	// - - in vout max ✅
	// - - include block height in sig script ✅
	// - has two outputs
	// - - compute wtxids and witnessCommitement = sha(merkle(wtxids) + bytes32(0x0))
	// - - out scriptputkey == op_return + PushBytes + witnessCommitement
	// - - other output has fee collection
	// serialize Coinbase with Witness
	// append beginning of tx list

	coinbaseVin := mempool.TxIn{
		Txid:       "0000000000000000000000000000000000000000000000000000000000000000",
		Vout:       0xffffffff,
		ScriptSig:  "0368c10c",
		Sequence:   0xffffffff,
		Witness:    []string{"0000000000000000000000000000000000000000000000000000000000000000"},
		IsCoinbase: true,
	}

	coinbaseVouts := []mempool.TxOut{
		{
			Value:        0,
			ScriptPubKey: "6a24aa21a9ed" + Hash256(GenerateMerkleRoot(wTxids)+"0000000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			Value:        uint64(feeCollected),
			ScriptPubKey: "76a914536ffa992491508dca0354e52f32a3a7a679a53a88ac",
		},
	}

	coinbaseTx := mempool.Transaction{
		Version:  2,
		Locktime: 0,
		Vin:      []mempool.TxIn{coinbaseVin},
		Vout:     coinbaseVouts,
	}

	cbTxId, _, _, err := coinbaseTx.Hash()
	if err != nil {
		return err
	}

	m.block.Txs = append([]string{cbTxId}, m.block.Txs...)

	// build block header
	// add block version 2
	// prev block bytes32(0x0)
	// add merklee root
	// add time
	// add nbits 0x1f00ffff
	// mine with nonce 0

	blockHeader := block.BlocKHeader{
		Version:           2,
		TimeStamp:         uint32(time.Now().Unix()),
		NBits:             0x1f00ffff,
		PreviousBlockHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Nonce:             0,
		MerkleRoot:        GenerateMerkleRoot(m.block.Txs),
	}

	respChan := make(chan uint32)
	doneChan := make(chan struct{})

	// spin 10 go routines which listen for nonces
	// after receiving nonces they build block header and send as response
	nextNonce := uint32(0)
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-doneChan:
					return
				default:
					// generate blockHash sha(sha(header_serialized))
					// mine it until less than difficulty (tune nonce)

					blockHeader.Nonce = atomic.AddUint32(&nextNonce, 1)
					blockHash := doubleHash(blockHeader.Serialize())
					if bytes.Compare(blockHash, GivenDifficulty) < 0 {
						respChan <- blockHeader.Nonce
						return
					}
				}
			}
		}()
	}

	// wait for nonces
	nonce := <-respChan
	close(doneChan)
	blockHeader.Nonce = nonce

	os.Remove(path.OutFilePath)

	// open output.txt file and write blockHeader serialized , coinbase serialized , txids
	file, err := os.OpenFile(path.OutFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	cb_ser, _, _, err := coinbaseTx.Serialize()
	if err != nil {
		return err
	}

	file.WriteString(hex.EncodeToString(blockHeader.Serialize()) + "\n")
	file.WriteString(hex.EncodeToString(cb_ser) + "\n")
	for _, txId := range m.block.Txs {
		file.WriteString(txId + "\n")
	}

	m.logger.Infof("mined block %s", blockHeader.Nonce)
	m.logger.Infof("Total Fee Collected %d", feeCollected)
	m.logger.Infof("Total weight %d", weight)

	// Hash must be Le

	return nil
}

func doubleHash(header []byte) []byte {
	h := sha256.New()
	h.Write(header)
	firstHash := h.Sum(nil)

	h.Reset()
	h.Write(firstHash)

	return h.Sum(nil)
}
