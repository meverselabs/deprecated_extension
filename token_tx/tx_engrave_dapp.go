package token_tx

import (
	"bytes"
	"io"
	"log"

	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/extension/account_tx"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.EngraveDapp", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &EngraveDapp{
			Base: account_tx.Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*EngraveDapp)
		if !transaction.IsMainChain(loader.ChainCoord()) {
			return ErrNotMainChain
		}
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		log.Println(tx.Seq(), ", ", tx.From().String(), " : ", loader.Seq(tx.From()))
		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*EngraveDapp)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		chainCoord := ctx.ChainCoord()
		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(chainCoord, Fee); err != nil {
			return nil, err
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// EngraveDapp is a fleta.EngraveDapp
// It is engraved dapp on main chain
type EngraveDapp struct {
	account_tx.Base
	TokenAddress common.Address
	Height       uint32
	BlockHash    hash.Hash256
}

// Hash returns the hash value of it
func (tx *EngraveDapp) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo is a serialization function
func (tx *EngraveDapp) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenAddress.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, tx.Height); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.BlockHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *EngraveDapp) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenAddress.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Height = v
	}
	if n, err := tx.BlockHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
