package keyring

import (
	"context"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"

	btcec2 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/btcd/btcec"
	"github.com/tendermint/crypto/sha3"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bytes"
)

type keystoreEth struct {
	Keyring
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
	if scp.ToECDSA() == nil || scp.ToECDSA().X == nil || scp.ToECDSA().Y == nil {
		return nil
	}
	ethPKBytes := elliptic.Marshal(btcec.S256(), scp.ToECDSA().X, scp.ToECDSA().Y)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(ethPKBytes[1:])
	b := crypto.Address(hash.Sum(nil)[12:])
	return b
}

// VerifySignature signed data.
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

func NewKeyRingETH(kr Keyring) Keyring {
	return keystoreEth{kr}
}

// Sign Implement Signer for cosmos sdk interface for eth
func (ks keystoreEth) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	priv, err := ks.getPriv(uid)
	if err != nil {
		return nil, nil, err
	}
	return sign(priv, msg)
}

func (ks keystoreEth) getPriv(uid string) (types.PrivKey, error) {
	record, err := ks.Key(uid)
	if err != nil {
		return nil, err
	}
	var priv types.PrivKey
	switch i := record.Item.(type) {
	case *Record_Local_:
		if i.Local.PrivKey == nil {
			return nil, fmt.Errorf("private key not available")
		}
		priv, err = legacy.PrivKeyFromBytes(i.Local.PrivKey.Value)
		return priv, err
	default:
		return nil, fmt.Errorf("currently supports for local key only")
	}
}

func sign(priv types.PrivKey, msg []byte) ([]byte, types.PubKey, error) {
	secp256k1Priv, ok := priv.(*secp256k1.PrivKey)

	if !ok {
		return nil, nil, fmt.Errorf("prv key could not converted into secp256k1 priv key")
	}
	privSigner := secp256k1.PrivKeyEth{PrivKey: *secp256k1Priv}
	sig, err := privSigner.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	return sig, &PubKeyETH{privSigner.PubKey()}, nil
}

func (ks keystoreEth) SignTx(tx *ethtypes.Transaction, signer ethtypes.Signer, uid string) (*ethtypes.Transaction, error) {
	priv, err := ks.getPriv(uid)
	if err != nil {
		return nil, err
	}
	btcePriv, _ := btcec2.PrivKeyFromBytes(priv.Bytes())
	ecdaPriv := btcePriv.ToECDSA()
	return ethtypes.SignTx(tx, signer, ecdaPriv)
}

// NewKeyedTransactorWithChainID return SignerFn for go-eth
// it uses uid name for getting priv key and address instead of passed eth.Address
func NewKeyedTransactorWithChainID(kr Keyring, uid string, chainID *big.Int) (*bind.TransactOpts, error) {
	ks := keystoreEth{kr}
	priv, err := ks.getPriv(uid)
	if err != nil {
		return nil, err
	}
	pubKey := PubKeyETH{priv.PubKey()}
	keyAddr := ethcommon.BytesToAddress(pubKey.Address())
	signer := ethtypes.LatestSignerForChainID(chainID)
	return &bind.TransactOpts{
		Context: context.Background(),
		From:    keyAddr,
		Signer: func(addr ethcommon.Address, tx *ethtypes.Transaction) (*ethtypes.Transaction, error) {
			if addr != keyAddr {
				return nil, bind.ErrNotAuthorized
			}
			return ks.SignTx(tx, signer, uid)
		},
	}, nil
}
