package bigint

import (
	"bytes"
	"fmt"
	"math/big"
)

type Int struct {
	*big.Int
}

func FromDecimal(decimal string) (Int, error) {
	i, ok := new(big.Int).SetString(decimal, 10)
	if !ok {
		return Int{}, fmt.Errorf("bigint: cant convert decimal %q to *big.Int", decimal)
	}
	return Int{i}, nil
}

var quote = []byte(`"`)
var null = []byte(`null`)
var hprx = []byte(`0x`) // `0x`

func (i *Int) UnmarshalJSON(text []byte) error {
	var ok bool
	if bytes.HasPrefix(text, quote) {
		n := text[1 : len(text)-1]
		if bytes.HasPrefix(n, hprx) {
			r := string(n[2:])
			if i.Int, ok = new(big.Int).SetString(r, 16); !ok {
				return fmt.Errorf(`bigint: can't convert "0x%s" to *big.Int`, r)
			}
			return nil
		}

		r := string(n)
		if i.Int, ok = new(big.Int).SetString(r, 10); !ok {
			return fmt.Errorf(`bigint: can't convert "%s" to *big.Int`, r)
		}
		return nil
	}

	if bytes.Equal(text, null) {
		i.Int = new(big.Int)
		return nil
	}

	r := string(text)
	if i.Int, ok = new(big.Int).SetString(r, 10); !ok {
		return fmt.Errorf("bigint: can't convert %s to *big.Int", r)
	}
	return nil
}
