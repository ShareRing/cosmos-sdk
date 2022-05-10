package secp256k1

import (
	secp256k1 "github.com/btcsuite/btcd/btcec"
)

type PrivKeyEth struct {
	PrivKey
}

// Sign the provided hash and convert it to the ethereum (r,s,v) format.
func (privKey *PrivKeyEth) Sign(sighash []byte) ([]byte, error) { // TODO: Hoai check if this compatible with BEPs?
	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey.Key)
	signature, err := secp256k1.SignCompact(secp256k1.S256(), priv, sighash, false)
	if err != nil {
		return nil, err
	}
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}
