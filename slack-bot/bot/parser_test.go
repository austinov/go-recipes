package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	cases := []struct {
		text     string
		expQuery Query
	}{
		{
			text: "events of Metallica",
			expQuery: Query{
				Band: "Metallica",
				City: "",
				From: 0,
				To:   0,
			},
		},
		{

			text: "events of A Day to remember",
			expQuery: Query{
				Band: "A Day to remember",
				City: "",
				From: 0,
				To:   0,
			},
		},
		{
			text: "events of A Day to remember in Moscow",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 0,
				To:   0,
			},
		},
		{
			text: "events in Moscow",
			expQuery: Query{
				Band: "",
				City: "Moscow",
				From: 0,
				To:   0,
			},
		},
		{
			text: "events in Sergiev Posad",
			expQuery: Query{
				Band: "",
				City: "Sergiev Posad",
				From: 0,
				To:   0,
			},
		},
		{
			text: "events of A Day to remember in Moscow at 15.12.2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260835200,
				To:   1260921599,
			},
		},
		{
			text: "events of A Day to remember in Moscow at 15 Dec 2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260835200,
				To:   1260921599,
			},
		},
		{
			text: "events of A Day to remember in Moscow at 15/12/2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260835200,
				To:   1260921599,
			},
		},
		{
			text: "events of A Day to remember in Moscow since 12.12.2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   0,
			},
		},
		{
			text: "events of A Day to remember in Moscow till 12.12.2014",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 0,
				To:   1418428799,
			},
		},
		{
			text: "events of A Day to remember in Moscow since 12.12.2009 till 15.12.2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   1260921599,
			},
		},
		{
			text: "events of A Day to remember in Moscow for 12.12.2009-12.09.2014",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   1410566399,
			},
		},
		{
			text: "events of A Day to remember in Moscow for 12.12.2009 - 12.09.2014",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   1410566399,
			},
		},
		{
			text: "events of A Day to remember in Moscow for 12.12.2009 and 12.09.2014",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   1410566399,
			},
		},
		{
			text: "events in Moscow of A Day to remember at 12.12.2009",
			expQuery: Query{
				Band: "A Day to remember",
				City: "Moscow",
				From: 1260576000,
				To:   1260662399,
			},
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expQuery, Parse(c.text), c.text)
	}
}

func TestParserValid(t *testing.T) {
	cases := []struct {
		text     string
		expQuery Query
		expValid bool
	}{
		{
			text: "events of Metallica",
			expQuery: Query{
				Band: "Metallica",
				City: "",
				From: 0,
				To:   0,
			},
			expValid: true,
		},
		{
			text: "events in Moscow",
			expQuery: Query{
				Band: "",
				City: "Moscow",
				From: 0,
				To:   0,
			},
			expValid: true,
		},
		{
			text: "events of Metallica in Moscow",
			expQuery: Query{
				Band: "Metallica",
				City: "Moscow",
				From: 0,
				To:   0,
			},
			expValid: true,
		},
		{
			text: "events at 15 Dec 2009",
			expQuery: Query{
				Band: "",
				City: "",
				From: 1260835200,
				To:   1260921599,
			},
			expValid: false,
		},
		{
			text: "events not valid query",
			expQuery: Query{
				Band: "",
				City: "",
				From: 0,
				To:   0,
			},
			expValid: false,
		},
	}

	for _, c := range cases {
		query := Parse(c.text)
		assert.Equal(t, c.expQuery, query, c.text)
		assert.Equal(t, c.expValid, query.IsValid(), c.text)
	}
}
