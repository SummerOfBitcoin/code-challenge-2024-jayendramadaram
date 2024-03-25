package mempool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sob-miner/internal/ierrors"
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
			err = nil // ignore for now
		case transaction.P2PKH:

			tempTx := *t

			stackElem := strings.Split(input.ScriptSigAsm, " ")

			// compressed pubkey
			pubKey := stackElem[len(stackElem)-1]
			Signature := stackElem[1]

			signatureBytes := MustHexDecode(Signature)
			lastSigByte := signatureBytes[len(signatureBytes)-1]

			MessageHash := generateMessageHash(tempTx, i, input, lastSigByte)

			// fmt.Printf("\n encountered a p2pkh signature: %s", input.Txid)

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
			//TODO
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
		fmt.Printf("signature: %s\n and pubkey: %s", hex.EncodeToString(signature.Serialize()), hex.EncodeToString(publicKey.SerializeUncompressed()))
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

func generateMessageHash(tempTx Transaction, i int, input TxIn, sigHash byte) []byte {
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

// 2102b2fb48ce4536bc0218d0d72d84d791f07649b0650cecb46d9b1ee94afc1785d4ac7364000068
// 2102b2fb48ce4536bc0218d0d72d84d791f07649b0650cecb46d9b1ee94afc1785d4ac736460b268
// 304402201008e236fa8cd0f25df4482dddbb622e8a8b26ef0ba731719458de3ccd93805b022032f8ebe514ba5f672466eba334639282616bb3c2f0ab09998037513d1f9e3d6d
// 304402203ea4d583d963c7eaf6693c94bf30a78af1bf35aef91d9357d4097f0e08886491022073cd31d66c9f3c035e6cf8c8571de6edf342d45a1d71d1ff74a40b08a9a67e9f01
