package mathutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountOfDigits(t *testing.T) {
	cases := []struct {
		input    int64
		expected int
	}{
		{
			input:    0,
			expected: 1,
		},
		{
			input:    1,
			expected: 1,
		},
		{
			input:    42,
			expected: 2,
		},
		{
			input:    -99,
			expected: 2,
		},
		{
			input:    876,
			expected: 3,
		},
		{
			input:    2345,
			expected: 4,
		},
		{
			input:    54321,
			expected: 5,
		},
		{
			input:    543210,
			expected: 6,
		},
		{
			input:    2132435,
			expected: 7,
		},
		{
			input:    21324354,
			expected: 8,
		},
		{
			input:    213243546,
			expected: 9,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, CountOfDigits(c.input))
	}
}

func TestIntToDigits(t *testing.T) {
	cases := []struct {
		input    int64
		expected []int
	}{
		{
			input:    0,
			expected: []int{0},
		},
		{
			input:    1,
			expected: []int{1},
		},
		{
			input:    42,
			expected: []int{4, 2},
		},
		{
			input:    -99,
			expected: []int{-9, -9},
		},
		{
			input:    876,
			expected: []int{8, 7, 6},
		},
		{
			input:    2345,
			expected: []int{2, 3, 4, 5},
		},
		{
			input:    54321,
			expected: []int{5, 4, 3, 2, 1},
		},
		{
			input:    543210,
			expected: []int{5, 4, 3, 2, 1, 0},
		},
		{
			input:    2132435,
			expected: []int{2, 1, 3, 2, 4, 3, 5},
		},
		{
			input:    21324354,
			expected: []int{2, 1, 3, 2, 4, 3, 5, 4},
		},
		{
			input:    213243546,
			expected: []int{2, 1, 3, 2, 4, 3, 5, 4, 6},
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, IntToDigits(c.input))
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		input    uint64
		expected string
	}{
		{
			input:    0,
			expected: "0 B",
		},
		{
			input:    1023,
			expected: "1023 B",
		},
		{
			input:    1046,
			expected: "1.02 KB",
		},
		{
			input:    11838423,
			expected: "11.29 MB",
		},
		{
			input:    2512555868,
			expected: "2.34 GB",
		},
		{
			input:    1715238139330,
			expected: "1.56 TB",
		},
		{
			input:    2634263381319330,
			expected: "2.34 PB",
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, HumanBytes(c.input))
	}
}
