package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilShort(t *testing.T) {
	for _, s := range []struct {
		fil    string
		expect string
	}{

		{fil: "1", expect: "1 KAKH"},
		{fil: "1.1", expect: "1.1 KAKH"},
		{fil: "12", expect: "12 KAKH"},
		{fil: "123", expect: "123 KAKH"},
		{fil: "123456", expect: "123456 KAKH"},
		{fil: "123.23", expect: "123.23 KAKH"},
		{fil: "123456.234", expect: "123456.234 KAKH"},
		{fil: "123456.2341234", expect: "123456.234 KAKH"},
		{fil: "123456.234123445", expect: "123456.234 KAKH"},

		{fil: "0.1", expect: "100 mKAKH"},
		{fil: "0.01", expect: "10 mKAKH"},
		{fil: "0.001", expect: "1 mKAKH"},

		{fil: "0.0001", expect: "100 μKAKH"},
		{fil: "0.00001", expect: "10 μKAKH"},
		{fil: "0.000001", expect: "1 μKAKH"},

		{fil: "0.0000001", expect: "100 nKAKH"},
		{fil: "0.00000001", expect: "10 nKAKH"},
		{fil: "0.000000001", expect: "1 nKAKH"},

		{fil: "0.0000000001", expect: "100 pKAKH"},
		{fil: "0.00000000001", expect: "10 pKAKH"},
		{fil: "0.000000000001", expect: "1 pKAKH"},

		{fil: "0.0000000000001", expect: "100 fKAKH"},
		{fil: "0.00000000000001", expect: "10 fKAKH"},
		{fil: "0.000000000000001", expect: "1 fKAKH"},

		{fil: "0.0000000000000001", expect: "100 aKAKH"},
		{fil: "0.00000000000000001", expect: "10 aKAKH"},
		{fil: "0.000000000000000001", expect: "1 aKAKH"},

		{fil: "0.0000012", expect: "1.2 μKAKH"},
		{fil: "0.00000123", expect: "1.23 μKAKH"},
		{fil: "0.000001234", expect: "1.234 μKAKH"},
		{fil: "0.0000012344", expect: "1.234 μKAKH"},
		{fil: "0.00000123444", expect: "1.234 μKAKH"},

		{fil: "0.0002212", expect: "221.2 μKAKH"},
		{fil: "0.00022123", expect: "221.23 μKAKH"},
		{fil: "0.000221234", expect: "221.234 μKAKH"},
		{fil: "0.0002212344", expect: "221.234 μKAKH"},
		{fil: "0.00022123444", expect: "221.234 μKAKH"},

		{fil: "-1", expect: "-1 KAKH"},
		{fil: "-1.1", expect: "-1.1 KAKH"},
		{fil: "-12", expect: "-12 KAKH"},
		{fil: "-123", expect: "-123 KAKH"},
		{fil: "-123456", expect: "-123456 KAKH"},
		{fil: "-123.23", expect: "-123.23 KAKH"},
		{fil: "-123456.234", expect: "-123456.234 KAKH"},
		{fil: "-123456.2341234", expect: "-123456.234 KAKH"},
		{fil: "-123456.234123445", expect: "-123456.234 KAKH"},

		{fil: "-0.1", expect: "-100 mKAKH"},
		{fil: "-0.01", expect: "-10 mKAKH"},
		{fil: "-0.001", expect: "-1 mKAKH"},

		{fil: "-0.0001", expect: "-100 μKAKH"},
		{fil: "-0.00001", expect: "-10 μKAKH"},
		{fil: "-0.000001", expect: "-1 μKAKH"},

		{fil: "-0.0000001", expect: "-100 nKAKH"},
		{fil: "-0.00000001", expect: "-10 nKAKH"},
		{fil: "-0.000000001", expect: "-1 nKAKH"},

		{fil: "-0.0000000001", expect: "-100 pKAKH"},
		{fil: "-0.00000000001", expect: "-10 pKAKH"},
		{fil: "-0.000000000001", expect: "-1 pKAKH"},

		{fil: "-0.0000000000001", expect: "-100 fKAKH"},
		{fil: "-0.00000000000001", expect: "-10 fKAKH"},
		{fil: "-0.000000000000001", expect: "-1 fKAKH"},

		{fil: "-0.0000000000000001", expect: "-100 aKAKH"},
		{fil: "-0.00000000000000001", expect: "-10 aKAKH"},
		{fil: "-0.000000000000000001", expect: "-1 aKAKH"},

		{fil: "-0.0000012", expect: "-1.2 μKAKH"},
		{fil: "-0.00000123", expect: "-1.23 μKAKH"},
		{fil: "-0.000001234", expect: "-1.234 μKAKH"},
		{fil: "-0.0000012344", expect: "-1.234 μKAKH"},
		{fil: "-0.00000123444", expect: "-1.234 μKAKH"},

		{fil: "-0.0002212", expect: "-221.2 μKAKH"},
		{fil: "-0.00022123", expect: "-221.23 μKAKH"},
		{fil: "-0.000221234", expect: "-221.234 μKAKH"},
		{fil: "-0.0002212344", expect: "-221.234 μKAKH"},
		{fil: "-0.00022123444", expect: "-221.234 μKAKH"},
	} {
		s := s
		t.Run(s.fil, func(t *testing.T) {
			f, err := ParseFIL(s.fil)
			require.NoError(t, err)
			require.Equal(t, s.expect, f.Short())
		})
	}
}
