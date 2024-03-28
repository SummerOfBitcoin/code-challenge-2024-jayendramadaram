package mempool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sob-miner/internal/ierrors"
	"sob-miner/pkg/encoding"
	"sob-miner/pkg/opcode"
	"sob-miner/pkg/transaction"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"golang.org/x/crypto/ripemd160"
)

func (t *Transaction) ValidateTxScripts() error {
	// iter through inputs and validate each one of em based on their type
	for i, input := range t.Vin {
		var err error = nil
		switch transaction.Type(input.Prevout.ScriptPubKeyType) {
		case transaction.OP_RETURN_TYPE:
			err = ierrors.ErrUsingOpReturnAsInput
		case transaction.P2PK:
			err = nil // ignore for now no p2pk txs in assignment
		case transaction.P2PKH:

			tempTx := *t

			stackElem := strings.Split(input.ScriptSigAsm, " ")

			// compressed pubkey
			pubKey := stackElem[len(stackElem)-1]
			Signature := stackElem[1]

			signatureBytes := MustHexDecode(Signature)
			lastSigByte := signatureBytes[len(signatureBytes)-1]

			MessageHash := generateMessageHashLegacy(tempTx, i, input, lastSigByte)

			err = ECVerify(MessageHash, signatureBytes, MustHexDecode(pubKey))

		case transaction.P2SH:
			redeemScript := MustDecodeAsmScript(strings.Split(input.InnerRedeemScriptAsm, " "))
			redeemScripExpectedtHash := strings.Split(input.Prevout.ScriptPubKeyAsm, " ")[2]

			if hex.EncodeToString(H160(redeemScript)) != redeemScripExpectedtHash {
				err = ierrors.ErrRedeemScriptMismatch
			}

		case transaction.P2MS:
			err = nil

		case transaction.P2WSH:
			redeemScript := MustDecodeAsmScript(strings.Split(input.InnerWitnessScriptAsm, " "))
			redeemScripExpectedtHash := strings.Split(input.Prevout.ScriptPubKeyAsm, " ")[2]

			if hex.EncodeToString(Sha256(redeemScript)) != redeemScripExpectedtHash {
				err = ierrors.ErrRedeemScriptMismatch
			}

		case transaction.P2WPKH:
			tempTx := *t

			if (len(input.Witness)) != 2 {
				err = ierrors.ErrInvalidWitnessLength
				break
			}

			Signature := MustHexDecode(input.Witness[0])
			pubKeyHash := MustHexDecode(input.Witness[1])

			lastSigByte := Signature[len(Signature)-1]

			MessageHash := generateMessageHashSegwit(tempTx, i, input, lastSigByte)

			err = ECVerify(MessageHash, Signature, pubKeyHash)

		case transaction.P2TR:
			err = nil
		default:
			err = ierrors.ErrScriptValidation
		}
		if err != nil {
			fmt.Printf("\n encountered an error: %s for inputTxid: %s and vout number: %d", err, input.Txid, input.Vout)
			return err
		}
	}
	return nil
}

// verifies ecdsa signature from der encoding
// digest MessageHash Signed
// sig: signature with r and s in DER encoding
// pubkey: compressed 33 byte pubkey
func ECVerify(digest []byte, sig []byte, pubkey []byte) error {
	signature, err := ecdsa.ParseDERSignature(sig)
	if err != nil {
		return err
	}

	publicKey, err := btcec.ParsePubKey(pubkey)
	if err != nil {
		return err
	}

	if !signature.Verify(digest, publicKey) {
		// fmt.Printf("signature: %s\n and pubkey: %s", hex.EncodeToString(signature.Serialize()), hex.EncodeToString(publicKey.SerializeUncompressed()))
		return ierrors.ErrInvalidSignature
	}
	return nil
}

func MustHexDecode(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func MustDecodeAsmScript(asmScript []string) []byte {
	decoded_script := []byte{}
	for _, item := range asmScript {
		if len(item) > 3 && item[:3] == "OP_" {
			byteCode, ok := opcode.OpCodeMap[item]
			if !ok {
				panic("invalid opcode: " + item)
			}
			decoded_script = append(decoded_script, byteCode)
			continue
		}

		byteItem, _ := hex.DecodeString(item)

		decoded_script = append(decoded_script, byteItem...)
	}
	return decoded_script
}

func H160(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	firstHash := h.Sum(nil)

	h = ripemd160.New()
	h.Write(firstHash)

	return h.Sum(nil)
}

func Sha256(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	return h.Sum(nil)
}

func generateMessageHashLegacy(tempTx Transaction, i int, input TxIn, sigHash byte) []byte {
	var serializedTx []byte

	switch sigHash {
	case 0x01: // sighashAll
		for j := 0; j < len(tempTx.Vin); j++ {
			tempTx.Vin[j].ScriptSig = ""
		}
		tempTx.Vin[i].ScriptSig = input.Prevout.ScriptPubKey
		serializedTx = tempTx.MustSerializeWithSigHashAll()
	case 0x81:
		tempTx.Vin = []TxIn{input}
		tempTx.Vin[0].ScriptSig = input.Prevout.ScriptPubKey
		serializedTx = tempTx.MustSerializeWithSigHashAnyOneCanPaySigHashAll()
	default:
		for j := 0; j < len(tempTx.Vin); j++ {
			tempTx.Vin[j].ScriptSig = ""
		}
		tempTx.Vin[i].ScriptSig = input.Prevout.ScriptPubKey
		serializedTx = tempTx.MustSerializeWithSigHashAll()
	}

	return chainhash.DoubleHashB(serializedTx)
}

func generateMessageHashSegwit(tempTx Transaction, pos int, input TxIn, sigHash byte) []byte {
	var serializedTx []byte
	switch sigHash {
	case 0x01:
		serializedTx = SegwitSerializeAll(tempTx, input, []byte{0x01, 0x00, 0x00, 0x00})
	case 0x81:
		// tempTx.Vin = []TxIn{input}
		// tempTx.Vin[0].ScriptSig = input.Prevout.ScriptPubKey
		serializedTx = SegwitSerializeAllAnyOne(tempTx, input, []byte{0x81, 0x00, 0x00, 0x00})
	case 0x83:
		serializedTx = SegwitSerializeSingleAnyOne(tempTx, pos, input, []byte{0x83, 0x00, 0x00, 0x00})
	default:
		serializedTx = SegwitSerializeAll(tempTx, input, []byte{0x01, 0x00, 0x00, 0x00})
	}
	if tempTx.Vin[0].Txid == "f3898029a8699bd8b71dc6f20e7ec2762a945a30d6a9f18034ce92a9d6cdd26c" {
		fmt.Println("serializedTx", serializedTx)
	}
	return chainhash.DoubleHashB(serializedTx)
}

// returns pre image
// preimage = version ✅ + hash256(inputs) ✅ + hash256(sequences) ✅ + input ✅ + scriptcode ✅ + amount ✅ + sequence ✅ + hash256(outputs) + locktime ✅ + SIGHASH ✅
func SegwitSerializeAll(tempTx Transaction, inp TxIn, sigHash []byte) []byte {
	preImage := encoding.NewLEBuffer()

	preImage.Set(tempTx.Version)

	InputsBytes := encoding.NewLEBuffer()
	SequencesBytes := encoding.NewLEBuffer()

	for _, input := range tempTx.Vin {
		InputsBytes.SetBytes(MustHexDecode(input.Txid), true)
		InputsBytes.Set(input.Vout)

		SequencesBytes.Set(input.Sequence)
	}

	preImage.SetBytes(chainhash.DoubleHashB(InputsBytes.GetBuffer()), false)
	preImage.SetBytes(chainhash.DoubleHashB(SequencesBytes.GetBuffer()), false)

	preImage.SetBytes(MustHexDecode(inp.Txid), true)
	preImage.Set(inp.Vout)

	pubKeyHash := MustHexDecode(inp.Prevout.ScriptPubKey[4:])

	scriptCode := make([]byte, 0)
	scriptCode = append(scriptCode, []byte{
		0x19, 0x76, 0xa9, 0x14,
	}...)
	scriptCode = append(scriptCode, pubKeyHash...)
	scriptCode = append(scriptCode, 0x88, 0xac)

	preImage.SetBytes(scriptCode, false)
	preImage.Set(inp.Prevout.Value)
	preImage.Set(inp.Sequence)

	outputBytes := encoding.NewLEBuffer()

	for _, output := range tempTx.Vout {
		outputBytes.Set(output.Value)
		scriptPubKey := MustHexDecode(output.ScriptPubKey)
		outputBytes.Set(encoding.CompactSize(uint64(len(scriptPubKey))))
		outputBytes.SetBytes(scriptPubKey, false)
	}

	preImage.SetBytes(chainhash.DoubleHashB(outputBytes.GetBuffer()), false)
	preImage.Set(tempTx.Locktime)

	preImage.SetBytes(sigHash, false) // sighash
	return preImage.GetBuffer()
}

func SegwitSerializeAllAnyOne(tempTx Transaction, inp TxIn, sigHash []byte) []byte {
	preImage := encoding.NewLEBuffer()

	preImage.Set(tempTx.Version)

	preImage.SetBytes(MustHexDecode("0000000000000000000000000000000000000000000000000000000000000000"), false) // hashPrevouts
	preImage.SetBytes(MustHexDecode("0000000000000000000000000000000000000000000000000000000000000000"), false) // hashSequence

	preImage.SetBytes(MustHexDecode(inp.Txid), true)
	preImage.Set(inp.Vout)

	pubKeyHash := MustHexDecode(inp.Prevout.ScriptPubKey[4:])

	scriptCode := make([]byte, 0)
	scriptCode = append(scriptCode, []byte{
		0x19, 0x76, 0xa9, 0x14,
	}...)

	scriptCode = append(scriptCode, pubKeyHash...)
	scriptCode = append(scriptCode, 0x88, 0xac)

	preImage.SetBytes(scriptCode, false)
	preImage.Set(inp.Prevout.Value)
	preImage.Set(inp.Sequence)

	outputBytes := encoding.NewLEBuffer()

	for _, output := range tempTx.Vout {
		outputBytes.Set(output.Value)
		scriptPubKey := MustHexDecode(output.ScriptPubKey)
		outputBytes.Set(encoding.CompactSize(uint64(len(scriptPubKey))))
		outputBytes.SetBytes(scriptPubKey, false)
	}

	preImage.SetBytes(chainhash.DoubleHashB(outputBytes.GetBuffer()), false)
	preImage.Set(tempTx.Locktime)

	preImage.SetBytes(sigHash, false) // sighash

	return preImage.GetBuffer()
}

func SegwitSerializeSingleAnyOne(tempTx Transaction, pos int, inp TxIn, sigHash []byte) []byte {
	preImage := encoding.NewLEBuffer()

	preImage.Set(tempTx.Version)

	preImage.SetBytes(MustHexDecode("0000000000000000000000000000000000000000000000000000000000000000"), false) // hashPrevouts
	preImage.SetBytes(MustHexDecode("0000000000000000000000000000000000000000000000000000000000000000"), false) // hashSequence

	preImage.SetBytes(MustHexDecode(inp.Txid), true)
	preImage.Set(inp.Vout)

	pubKeyHash := MustHexDecode(inp.Prevout.ScriptPubKey[4:])

	scriptCode := make([]byte, 0)
	scriptCode = append(scriptCode, []byte{
		0x19, 0x76, 0xa9, 0x14,
	}...)

	scriptCode = append(scriptCode, pubKeyHash...)
	scriptCode = append(scriptCode, 0x88, 0xac)

	preImage.SetBytes(scriptCode, false)
	preImage.Set(inp.Prevout.Value)
	preImage.Set(inp.Sequence)

	outputBytes := encoding.NewLEBuffer()

	output := tempTx.Vout[pos]

	outputBytes.Set(output.Value)
	scriptPubKey := MustHexDecode(output.ScriptPubKey)
	outputBytes.Set(encoding.CompactSize(uint64(len(scriptPubKey))))
	outputBytes.SetBytes(scriptPubKey, false)

	preImage.SetBytes(chainhash.DoubleHashB(outputBytes.GetBuffer()), false)
	preImage.Set(tempTx.Locktime)

	preImage.SetBytes(sigHash, false) // sighash

	return preImage.GetBuffer()
}

// 2102b2fb48ce4536bc0218d0d72d84d791f07649b0650cecb46d9b1ee94afc1785d4ac7364000068
// 2102b2fb48ce4536bc0218d0d72d84d791f07649b0650cecb46d9b1ee94afc1785d4ac736460b268
// 304402201008e236fa8cd0f25df4482dddbb622e8a8b26ef0ba731719458de3ccd93805b022032f8ebe514ba5f672466eba334639282616bb3c2f0ab09998037513d1f9e3d6d
// 304402203ea4d583d963c7eaf6693c94bf30a78af1bf35aef91d9357d4097f0e08886491022073cd31d66c9f3c035e6cf8c8571de6edf342d45a1d71d1ff74a40b08a9a67e9f01

// p2wpkh
//02000000cbfaca386d65ea7043aaac40302325d0dc7391a73b585571e28d3287d6b162033bb13029ce7b1f559ef5e747fcac439f1455a2ec7c5f09b72290795e70665044ac4994014aa36b7f53375658ef595b3cb2891e1735fe5b441686f5e53338e76a010000001976a914aa966f56de599b4094b61aa68a2b3df9e97e9c4888ac3075000000000000ffffffff14fac4817a9374ced1fe58178dcc895a07b061f673ce52d91a6b74b0eda1dff50000000001000000
//02000000cbfaca386d65ea7043aaac40302325d0dc7391a73b585571e28d3287d6b162033bb13029ce7b1f559ef5e747fcac439f1455a2ec7c5f09b72290795e70665044ac4994014aa36b7f53375658ef595b3cb2891e1735fe5b441686f5e53338e76a010000001976a914aa966f56de599b4094b61aa68a2b3df9e97e9c4888ac3075000000000000ffffffff900a6c6ff6cd938bf863e50613a4ed5fb1661b78649fe354116edaf5d4abb9520000000001000000

//outpoints bytes
// 204e0000000000003276a914ce72abfd0e6d9354a660c18f2825eb392f060fdc88ac
// 204e0000000000001976a914ce72abfd0e6d9354a660c18f2825eb392f060fdc88ac
