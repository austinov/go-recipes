package common

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			<-time.After(5 * time.Second)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewHTTPClient(3 * time.Second)
	res, err := c.Get(ts.URL)
	assert.NoError(t, err, "normal request must not produce timeout error")
	assert.NotNil(t, res)
	res.Body.Close()

	res, err = c.Get(ts.URL + "/slow")
	assert.Error(t, err, "slow request must produce timeout error")
	assert.Nil(t, res)
}
