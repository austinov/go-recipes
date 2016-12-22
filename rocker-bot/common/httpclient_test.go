package common

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	expTimeout  = 3 * time.Second
	reqDuration = 5 * time.Second
)

func TestHTTPClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			<-time.After(reqDuration)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewHTTPClient(expTimeout)
	res, err := c.Get(ts.URL)
	assert.NoError(t, err, "normal request must not produce timeout error")
	assert.NotNil(t, res)
	res.Body.Close()

	begin := time.Now()
	res, err = c.Get(ts.URL + "/slow")
	dur := time.Since(begin)
	assert.Error(t, err, "slow request must produce timeout error")
	assert.Nil(t, res)
	assert.Equal(t, expTimeout, round(dur, time.Second))
}

// It was taken from
// http://grokbase.com/t/gg/golang-nuts/1492epp0qb/go-nuts-how-to-round-a-duration
func round(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
