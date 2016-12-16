package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuse(t *testing.T) {
	triggered := false
	fuseTriggers := make([]FuseTrigger, 0)
	fuseTriggers = append(fuseTriggers,
		NewFuseTrigger("APP", 3, func(kind string, err error) {
			triggered = true
		}))
	fuse := NewFuse(fuseTriggers)
	fuse.Process("APP", fmt.Errorf("APP error"))
	fuse.Process("APP", nil)
	fuse.Process("APP", fmt.Errorf("APP error"))
	fuse.Process("APP", fmt.Errorf("APP error"))
	assert.False(t, triggered)
	fuse.Process("APP", fmt.Errorf("APP error"))
	assert.True(t, triggered)
}
