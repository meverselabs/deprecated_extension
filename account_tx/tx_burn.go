package account_tx

import (
	"bytes"
	"io"

	"git.fleta.io/fleta/core/amount"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.Burn", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &Burn{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
			TokenCoord: coord.Clone(),
			Amount:     amount.NewCoinAmount(0, 0),
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Burn)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}
		if tx.Amount.Less(amount.COIN.DivC(10)) {
			return ErrDustAmount
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*Burn)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(tx.TokenCoord, Fee); err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(tx.TokenCoord, tx.Amount); err != nil {
			return nil, err
		}
		ctx.Commit(sn)
		return nil, nil
	})
}

// Burn is a fleta.Burn
// It is used to burn coin from the account
type Burn struct {
	Base
	TokenCoord *common.Coordinate
	Amount     *amount.Amount
}

// Hash returns the hash value of it
func (tx *Burn) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo is a serialization function
func (tx *Burn) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, tx.Seq_); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.From_.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenCoord.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *Burn) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Seq_ = v
	}
	if n, err := tx.From_.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenCoord.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
