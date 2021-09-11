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

		{fil: "1", expect: "1 KAK"},
		{fil: "1.1", expect: "1.1 KAK"},
		{fil: "12", expect: "12 KAK"},
		{fil: "123", expect: "123 KAK"},
		{fil: "123456", expect: "123456 KAK"},
		{fil: "123.23", expect: "123.23 KAK"},
		{fil: "123456.234", expect: "123456.234 KAK"},
		{fil: "123456.2341234", expect: "123456.234 KAK"},
		{fil: "123456.234123445", expect: "123456.234 KAK"},

		{fil: "0.1", expect: "100 mKAK"},
		{fil: "0.01", expect: "10 mKAK"},
		{fil: "0.001", expect: "1 mKAK"},

		{fil: "0.0001", expect: "100 μKAK"},
		{fil: "0.00001", expect: "10 μKAK"},
		{fil: "0.000001", expect: "1 μKAK"},

		{fil: "0.0000001", expect: "100 nKAK"},
		{fil: "0.00000001", expect: "10 nKAK"},
		{fil: "0.000000001", expect: "1 nKAK"},

		{fil: "0.0000000001", expect: "100 pKAK"},
		{fil: "0.00000000001", expect: "10 pKAK"},
		{fil: "0.000000000001", expect: "1 pKAK"},

		{fil: "0.0000000000001", expect: "100 fKAK"},
		{fil: "0.00000000000001", expect: "10 fKAK"},
		{fil: "0.000000000000001", expect: "1 fKAK"},

		{fil: "0.0000000000000001", expect: "100 aKAK"},
		{fil: "0.00000000000000001", expect: "10 aKAK"},
		{fil: "0.000000000000000001", expect: "1 aKAK"},

		{fil: "0.0000012", expect: "1.2 μKAK"},
		{fil: "0.00000123", expect: "1.23 μKAK"},
		{fil: "0.000001234", expect: "1.234 μKAK"},
		{fil: "0.0000012344", expect: "1.234 μKAK"},
		{fil: "0.00000123444", expect: "1.234 μKAK"},

		{fil: "0.0002212", expect: "221.2 μKAK"},
		{fil: "0.00022123", expect: "221.23 μKAK"},
		{fil: "0.000221234", expect: "221.234 μKAK"},
		{fil: "0.0002212344", expect: "221.234 μKAK"},
		{fil: "0.00022123444", expect: "221.234 μKAK"},

		{fil: "-1", expect: "-1 KAK"},
		{fil: "-1.1", expect: "-1.1 KAK"},
		{fil: "-12", expect: "-12 KAK"},
		{fil: "-123", expect: "-123 KAK"},
		{fil: "-123456", expect: "-123456 KAK"},
		{fil: "-123.23", expect: "-123.23 KAK"},
		{fil: "-123456.234", expect: "-123456.234 KAK"},
		{fil: "-123456.2341234", expect: "-123456.234 KAK"},
		{fil: "-123456.234123445", expect: "-123456.234 KAK"},

		{fil: "-0.1", expect: "-100 mKAK"},
		{fil: "-0.01", expect: "-10 mKAK"},
		{fil: "-0.001", expect: "-1 mKAK"},

		{fil: "-0.0001", expect: "-100 μKAK"},
		{fil: "-0.00001", expect: "-10 μKAK"},
		{fil: "-0.000001", expect: "-1 μKAK"},

		{fil: "-0.0000001", expect: "-100 nKAK"},
		{fil: "-0.00000001", expect: "-10 nKAK"},
		{fil: "-0.000000001", expect: "-1 nKAK"},

		{fil: "-0.0000000001", expect: "-100 pKAK"},
		{fil: "-0.00000000001", expect: "-10 pKAK"},
		{fil: "-0.000000000001", expect: "-1 pKAK"},

		{fil: "-0.0000000000001", expect: "-100 fKAK"},
		{fil: "-0.00000000000001", expect: "-10 fKAK"},
		{fil: "-0.000000000000001", expect: "-1 fKAK"},

		{fil: "-0.0000000000000001", expect: "-100 aKAK"},
		{fil: "-0.00000000000000001", expect: "-10 aKAK"},
		{fil: "-0.000000000000000001", expect: "-1 aKAK"},

		{fil: "-0.0000012", expect: "-1.2 μKAK"},
		{fil: "-0.00000123", expect: "-1.23 μKAK"},
		{fil: "-0.000001234", expect: "-1.234 μKAK"},
		{fil: "-0.0000012344", expect: "-1.234 μKAK"},
		{fil: "-0.00000123444", expect: "-1.234 μKAK"},

		{fil: "-0.0002212", expect: "-221.2 μKAK"},
		{fil: "-0.00022123", expect: "-221.23 μKAK"},
		{fil: "-0.000221234", expect: "-221.234 μKAK"},
		{fil: "-0.0002212344", expect: "-221.234 μKAK"},
		{fil: "-0.00022123444", expect: "-221.234 μKAK"},
	} {
		s := s
		t.Run(s.fil, func(t *testing.T) {
			f, err := ParseFIL(s.fil)
			require.NoError(t, err)
			require.Equal(t, s.expect, f.Short())
		})
	}
}
