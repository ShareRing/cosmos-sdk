package secp256k1

import (
	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

type PrivKeyEth struct {
	PrivKey
}

// Sign the provided hash and convert it to the ethereum (r,s,v) format.
func (privKey *PrivKeyEth) Sign(sighash []byte) ([]byte, error) {
	priv, _ := secp256k1.PrivKeyFromBytes(privKey.Key)
	signature, err := ecdsa.SignCompact(priv, sighash, false)
	if err != nil {
		return nil, err
	}
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}
