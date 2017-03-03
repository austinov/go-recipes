/*
	Package flagutils helps to distinguish if a flag was passed or not.

	Usage:
	    import (
			util "flagutils"
			"flag"
			"fmt"
		)
		var debug util.BoolFlag
		flag.Var(&debug, "debug", "debug mode")
		flag.Parse()

		if debug.Exist {
			fmt.Printf("debug mode not set")
		} else {
			fmt.Printf("debug mode set to %v", debug.Value)
		}
*/
package flagutils

import (
	"fmt"
	"strconv"
	"time"
)

// BoolFlag represents bool flag and implements flag.boolFlag interface
type BoolFlag struct {
	Exist bool
	Value bool
}

func (b *BoolFlag) IsBoolFlag() bool {
	return true
}

func (b *BoolFlag) Set(s string) error {
	var err error
	if s != "" {
		b.Value, err = strconv.ParseBool(s)
	}
	b.Exist = true
	return err
}

func (b *BoolFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// IntFlag represents bool flag and implements flag.Value interface
type IntFlag struct {
	Exist bool
	Value int
}

func (b *IntFlag) Set(s string) error {
	if s != "" {
		if value, err := strconv.ParseInt(s, 0, 64); err == nil {
			b.Value = int(value)
		} else {
			return err
		}
	}
	b.Exist = true
	return nil
}

func (b *IntFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// Int64Flag represents bool flag and implements flag.Value interface
type Int64Flag struct {
	Exist bool
	Value int64
}

func (b *Int64Flag) Set(s string) error {
	var err error
	if s != "" {
		b.Value, err = strconv.ParseInt(s, 0, 64)
	}
	b.Exist = true
	return err
}

func (b *Int64Flag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// UintFlag represents bool flag and implements flag.Value interface
type UintFlag struct {
	Exist bool
	Value uint
}

func (b *UintFlag) Set(s string) error {
	if s != "" {
		if value, err := strconv.ParseUint(s, 0, 64); err == nil {
			b.Value = uint(value)
		} else {
			return err
		}
	}
	b.Exist = true
	return nil
}

func (b *UintFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// Uint64Flag represents bool flag and implements flag.Value interface
type Uint64Flag struct {
	Exist bool
	Value uint64
}

func (b *Uint64Flag) Set(s string) error {
	var err error
	if s != "" {
		b.Value, err = strconv.ParseUint(s, 0, 64)
	}
	b.Exist = true
	return err
}

func (b *Uint64Flag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// StringFlag represents bool flag and implements flag.Value interface
type StringFlag struct {
	Exist bool
	Value string
}

func (b *StringFlag) Set(s string) error {
	b.Value = s
	b.Exist = true
	return nil
}

func (b *StringFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// Float64Flag represents bool flag and implements flag.Value interface
type Float64Flag struct {
	Exist bool
	Value float64
}

func (b *Float64Flag) Set(s string) error {
	var err error
	if s != "" {
		b.Value, err = strconv.ParseFloat(s, 0)
	}
	b.Exist = true
	return err
}

func (b *Float64Flag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// DurationFlag represents bool flag and implements flag.Value interface
type DurationFlag struct {
	Exist bool
	Value time.Duration
}

func (b *DurationFlag) Set(s string) error {
	var err error
	if s != "" {
		b.Value, err = time.ParseDuration(s)
	}
	b.Exist = true
	return err
}

func (b *DurationFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}
