package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// https://github.com/jackc/pgx/blob/d8b38b28be8ca3b4babb3d3ea845be7894562312/pgtype/numeric.go#L163
func parseNumericString(str string) (n *big.Int, exp int32, err error) {
	parts := strings.SplitN(str, ".", 2)
	digits := strings.Join(parts, "")

	if len(parts) > 1 {
		exp = int32(-len(parts[1]))
	} else {
		for len(digits) > 1 && digits[len(digits)-1] == '0' && digits[len(digits)-2] != '-' {
			digits = digits[:len(digits)-1]
			exp++
		}
	}

	accum := &big.Int{}
	if _, ok := accum.SetString(digits, 10); !ok {
		return nil, 0, fmt.Errorf("%s is not a number", str)
	}

	return accum, exp, nil
}

func ParseWeiString(valueStr string) (n *big.Int, err error) {
	parsedValue, exp, err := parseNumericString(valueStr)
	if err != nil {
		return nil, err
	}

	// Adjust the exponent to handle conversion to Wei (1e18)
	weiMultiplier := new(big.Int)
	weiMultiplier.Exp(big.NewInt(10), big.NewInt(int64(18+exp)), nil) // 10^(18 + exp)

	// Multiply parsed value by the Wei multiplier
	result := new(big.Int).Mul(parsedValue, weiMultiplier)

	return result, nil
}
