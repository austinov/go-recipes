package flagutils

import (
	"flag"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type useCase struct {
	flagName  string
	flagValue interface{}
}

var (
	once sync.Once
)

func setUp() {
	once.Do(func() {
		var (
			boolFlag     BoolFlag
			intFlag      IntFlag
			int64Flag    Int64Flag
			uintFlag     UintFlag
			uint64Flag   Uint64Flag
			stringFlag   StringFlag
			float64Flag  Float64Flag
			durationFlag DurationFlag
		)
		flag.Var(&boolFlag, "test_bool_flag", "test bool flag")
		flag.Var(&intFlag, "test_int_flag", "test int flag")
		flag.Var(&int64Flag, "test_int64_flag", "test int64 flag")
		flag.Var(&uintFlag, "test_uint_flag", "test uint flag")
		flag.Var(&uint64Flag, "test_uint64_flag", "test uint64 flag")
		flag.Var(&stringFlag, "test_string_flag", "test string flag")
		flag.Var(&float64Flag, "test_float64_flag", "test float64 flag")
		flag.Var(&durationFlag, "test_duration_flag", "test duration flag")
	})
}

func TestNotExistFlags(t *testing.T) {
	args := []string{}
	cases := []useCase{
		{
			flagName:  "test_bool_flag",
			flagValue: &BoolFlag{Exist: false, Value: false},
		},
		{
			flagName:  "test_int_flag",
			flagValue: &IntFlag{Exist: false, Value: 0},
		},
		{
			flagName:  "test_int64_flag",
			flagValue: &Int64Flag{Exist: false, Value: 0},
		},
		{
			flagName:  "test_uint_flag",
			flagValue: &UintFlag{Exist: false, Value: 0},
		},
		{
			flagName:  "test_uint64_flag",
			flagValue: &Uint64Flag{Exist: false, Value: 0},
		},
		{
			flagName:  "test_string_flag",
			flagValue: &StringFlag{Exist: false, Value: ""},
		},
		{
			flagName:  "test_float64_flag",
			flagValue: &Float64Flag{Exist: false, Value: 0},
		},
		{
			flagName:  "test_duration_flag",
			flagValue: &DurationFlag{Exist: false, Value: 0},
		},
	}
	testCases(t, args, cases)
}

func TestExistEmptyFlags(t *testing.T) {
	args := []string{
		"-test_bool_flag",
		"-test_int_flag=",
		"-test_int64_flag=",
		"-test_uint_flag=",
		"-test_uint64_flag=",
		"-test_string_flag=",
		"-test_float64_flag=",
		"-test_duration_flag=",
	}
	cases := []useCase{
		{
			flagName:  "test_bool_flag",
			flagValue: &BoolFlag{Exist: true, Value: true},
		},
		{
			flagName:  "test_int_flag",
			flagValue: &IntFlag{Exist: true, Value: 0},
		},
		{
			flagName:  "test_int64_flag",
			flagValue: &Int64Flag{Exist: true, Value: 0},
		},
		{
			flagName:  "test_uint_flag",
			flagValue: &UintFlag{Exist: true, Value: 0},
		},
		{
			flagName:  "test_uint64_flag",
			flagValue: &Uint64Flag{Exist: true, Value: 0},
		},
		{
			flagName:  "test_string_flag",
			flagValue: &StringFlag{Exist: true, Value: ""},
		},
		{
			flagName:  "test_float64_flag",
			flagValue: &Float64Flag{Exist: true, Value: 0},
		},
		{
			flagName:  "test_duration_flag",
			flagValue: &DurationFlag{Exist: true, Value: 0},
		},
	}
	testCases(t, args, cases)
}

func TestExistNotEmptyFlags(t *testing.T) {
	args := []string{
		"-test_bool_flag=false",
		"-test_int_flag=1",
		"-test_int64_flag=2",
		"-test_uint_flag=3",
		"-test_uint64_flag=4",
		"-test_string_flag=/bin/bash",
		"-test_float64_flag=1.23",
		"-test_duration_flag=5ms",
	}
	cases := []useCase{
		{
			flagName:  "test_bool_flag",
			flagValue: &BoolFlag{Exist: true, Value: false},
		},
		{
			flagName:  "test_int_flag",
			flagValue: &IntFlag{Exist: true, Value: 1},
		},
		{
			flagName:  "test_int64_flag",
			flagValue: &Int64Flag{Exist: true, Value: 2},
		},
		{
			flagName:  "test_uint_flag",
			flagValue: &UintFlag{Exist: true, Value: 3},
		},
		{
			flagName:  "test_uint64_flag",
			flagValue: &Uint64Flag{Exist: true, Value: 4},
		},
		{
			flagName:  "test_string_flag",
			flagValue: &StringFlag{Exist: true, Value: "/bin/bash"},
		},
		{
			flagName:  "test_float64_flag",
			flagValue: &Float64Flag{Exist: true, Value: 1.23},
		},
		{
			flagName:  "test_duration_flag",
			flagValue: &DurationFlag{Exist: true, Value: 5000000},
		},
	}
	testCases(t, args, cases)
}

func testCases(t *testing.T, args []string, cases []useCase) {
	setUp()

	if err := flag.CommandLine.Parse(args); err != nil {
		assert.Error(t, err, "parse arguments error")
	}

	m := make(map[string]interface{})
	visitor := func(f *flag.Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			m[f.Name] = f.Value
		}
	}
	flag.VisitAll(visitor)

	for _, c := range cases {
		assert.Equal(t, c.flagValue, m[c.flagName], c.flagName)
	}
}
