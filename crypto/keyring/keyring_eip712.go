package keyring

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	crypto_eth "github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/btcd/btcec"
	"github.com/tendermint/crypto/sha3"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bytes"
)

type keystoreEIP712 struct {
	keystore
}

type PubKeyETH struct {
	types.PubKey
}

type EthAddress struct {
	bytes.HexBytes
}

func (addr EthAddress) hex() []byte {

	buf := make([]byte, 0, len(addr.Bytes())*2+2)
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], addr.Bytes()[:])
	return buf[:]
}

// Address return ETH address style
func (pubKey *PubKeyETH) Address() crypto.Address {
	kb := pubKey.Bytes()
	scp, e := btcec.ParsePubKey(kb, btcec.S256())
	if e != nil {
		panic(e)
	}
	ethPKBytes := crypto_eth.FromECDSAPub(scp.ToECDSA())
	hash := sha3.NewLegacyKeccak256()
	hash.Write(ethPKBytes[1:])
	b := crypto.Address(hash.Sum(nil)[12:])
	return b
}

// VerifySignature eip712 signed data.
func (pubKey *PubKeyETH) VerifySignature(signHash []byte, signature []byte) bool {
	if len(signature) != 65 {
		return false
	}
	btcsig := make([]byte, 65)
	btcsig[0] = signature[64]
	copy(btcsig[1:], signature)
	p, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, signHash)
	if err != nil {
		return false
	}
	parsedPubKey, err := btcec.ParsePubKey(pubKey.PubKey.Bytes(), btcec.S256())
	if err != nil {
		return false
	}
	return parsedPubKey.IsEqual(p)
}

func NewKeyRingEIP712(kr Keyring) Keyring {
	return keystoreEIP712{kr.(keystore)}
}

func (ks keystoreEIP712) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	info, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}
	var priv types.PrivKey

	switch i := info.(type) {
	case localInfo:
		if i.PrivKeyArmor == "" {
			return nil, nil, fmt.Errorf("private key not available")
		}
		priv, err = legacy.PrivKeyFromBytes([]byte(i.PrivKeyArmor))
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("eip712 currently supports for local key")
	}
	return signEIP712(priv, msg)
}

func signEIP712(priv types.PrivKey, msg []byte) ([]byte, types.PubKey, error) {
	secp256k1Priv, ok := priv.(*secp256k1.PrivKey)
	if !ok {
		return nil, nil, fmt.Errorf("prv key could not converted into secp256k1 priv key")
	}
	privEIP712Signer := secp256k1.PrivKeyEIP712Signer{PrivKey: *secp256k1Priv}
	sig, err := privEIP712Signer.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	return sig, &PubKeyETH{privEIP712Signer.PubKey()}, nil
}
