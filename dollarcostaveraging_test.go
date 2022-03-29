package ftxgo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsBuyOrderOkayHappyCase(t *testing.T) {

	type inputs struct {
		desc      string
		orders    []Order
		expectBuy bool
	}
	// NOTE: due to rounding of our bids, our DCA is very sensitive..
	// because we might be buying daily, but we buy 98% of what we wanted and
	// and then, the next day we double spend
	tests := []inputs{
		{
			desc: "order yesterday, no order today, should buy",
			orders: []Order{
				{
					Price:     2837.5,
					Quantity:  0.018,
					Timestamp: time.Date(2022, time.March, 25, 8, 12, 0, 0, time.UTC),
				},
			},
			expectBuy: true,
		},
		{
			desc: "one order yesterday + 1 today, all done",
			orders: []Order{
				{
					Price:     3000,
					Quantity:  0.017,
					Timestamp: time.Date(2022, time.March, 25, 9, 12, 0, 0, time.UTC),
				},
				{
					Price:     2850,
					Quantity:  0.018, // should be 0.0182..
					Timestamp: time.Date(2022, time.March, 26, 9, 12, 0, 0, time.UTC),
				},
			},
			expectBuy: false,
		},
		{
			desc: "one order yesterday + 1 today, but avg is below target due to rounding",
			orders: []Order{
				{
					Price:     2837.5,
					Quantity:  0.018,
					Timestamp: time.Date(2022, time.March, 25, 9, 12, 0, 0, time.UTC),
				},
				{
					Price:     2849.9,
					Quantity:  0.017,
					Timestamp: time.Date(2022, time.March, 26, 8, 12, 0, 0, time.UTC),
				},
			},
			expectBuy: true,
		},
		{
			desc:      "last day was correct, previous day bought too much",
			expectBuy: false,
			orders: []Order{
				{
					Price:     2849.9,
					Quantity:  0.017,
					Timestamp: time.Date(2022, time.March, 26, 8, 12, 0, 0, time.UTC),
				},
				{
					Price:     2844.7,
					Quantity:  0.017,
					Timestamp: time.Date(2022, time.March, 26, 9, 12, 0, 0, time.UTC),
				},
				{
					Price:     2862.2,
					Quantity:  0.017,
					Timestamp: time.Date(2022, time.March, 27, 10, 30, 0, 0, time.UTC),
				},
			},
		},
	}

	const budget = 51
	const intervalCounts = 2
	const interval = 24 * time.Hour
	for _, k := range tests {
		buy, _ := IsBuyOrderOkay(k.orders, float64(budget), interval, interval*intervalCounts)
		require.Equal(t, k.expectBuy, buy, k.desc)
	}
}
