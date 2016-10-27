package backoff

import (
	"fmt"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	eb := NewExpBackoffWithConfig(
		100*time.Millisecond,
		5*time.Second,
		2.0)
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
	eb.Reset()
	fmt.Println(eb.Delay())
	fmt.Println(eb.Delay())
}
