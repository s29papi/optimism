package main

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/crossdomain"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var UnknownNonceVersion = errors.New("Unknown nonce version")

var ZeroPadding = [32]byte{}

// checkOk checks if ok is false, and panics if so.
// Shorthand to ease go's god awful error handling
func checkOk(ok bool) {
	if !ok {
		panic(fmt.Errorf("checkOk failed"))
	}
}

// checkErr checks if err is not nil, and throws if so.
// Shorthand to ease go's god awful error handling
func checkErr(err error, failReason string) {
	if err != nil {
		panic(fmt.Errorf("%s: %s", failReason, err))
	}
}

// encodeCrossDomainMessage encodes a versioned cross domain message into a byte array.
func encodeCrossDomainMessage(nonce *big.Int, sender common.Address, target common.Address, value *big.Int, gasLimit *big.Int, data []byte) ([]byte, error) {
	_, version := crossdomain.DecodeVersionedNonce(nonce)

	var encoded []byte
	var err error
	if version.Cmp(big.NewInt(0)) == 0 {
		// Encode cross domain message V0
		encoded, err = crossdomain.EncodeCrossDomainMessageV0(target, sender, data, nonce)
	} else if version.Cmp(big.NewInt(1)) == 0 {
		// Encode cross domain message V1
		encoded, err = crossdomain.EncodeCrossDomainMessageV1(nonce, sender, target, value, gasLimit, data)
	} else {
		return nil, UnknownNonceVersion
	}

	return encoded, err
}

// hashWithdrawal hashes a withdrawal transaction.
func hashWithdrawal(nonce *big.Int, sender common.Address, target common.Address, value *big.Int, gasLimit *big.Int, data []byte) (common.Hash, error) {
	// Pack withdrawal
	wdtx := struct {
		Nonce    *big.Int
		Sender   common.Address
		Target   common.Address
		Value    *big.Int
		GasLimit *big.Int
		Data     []byte
	}{
		Nonce:    nonce,
		Sender:   sender,
		Target:   target,
		Value:    value,
		GasLimit: gasLimit,
		Data:     data,
	}
	packed, err := withdrawalTransactionArgs.Pack(&wdtx)
	if err != nil {
		return common.Hash{}, err
	}

	// Hash packed withdrawal (we ignore the pointer)
	return crypto.Keccak256Hash(packed[32:]), nil
}

// hashOutputRootProof hashes an output root proof.
func hashOutputRootProof(version common.Hash, stateRoot common.Hash, messagePasserStorageRoot common.Hash, latestBlockHash common.Hash) (common.Hash, error) {
	// Pack proof
	proof := struct {
		Version                  common.Hash
		StateRoot                common.Hash
		MessagePasserStorageRoot common.Hash
		LatestBlockHash          common.Hash
	}{
		Version:                  version,
		StateRoot:                stateRoot,
		MessagePasserStorageRoot: messagePasserStorageRoot,
		LatestBlockHash:          latestBlockHash,
	}
	packed, err := outputRootProofArgs.Pack(&proof)
	if err != nil {
		return common.Hash{}, err
	}

	// Hash packed proof
	return crypto.Keccak256Hash(packed), nil
}

// makeDepositTx creates a deposit transaction type.
func makeDepositTx(
	from common.Address,
	to common.Address,
	value *big.Int,
	mint *big.Int,
	gasLimit *big.Int,
	isCreate bool,
	data []byte,
	l1BlockHash common.Hash,
	logIndex *big.Int,
) types.DepositTx {
	// Create deposit transaction source
	udp := derive.UserDepositSource{
		L1BlockHash: l1BlockHash,
		LogIndex:    logIndex.Uint64(),
	}

	// Create deposit transaction
	depositTx := types.DepositTx{
		SourceHash:          udp.SourceHash(),
		From:                from,
		Value:               value,
		Gas:                 gasLimit.Uint64(),
		IsSystemTransaction: false, // This will never be a system transaction in the tests.
		Data:                data,
	}

	// Fill optional fields
	if mint.Cmp(big.NewInt(0)) == 1 {
		depositTx.Mint = mint
	}
	if !isCreate {
		depositTx.To = &to
	}

	return depositTx
}
