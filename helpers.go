package testhelpers

import (
	"bytes"
	"database/sql"
	"fmt"
	eden "github.com/byu-oit-ssengineering/tmt-eden"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"os"
)

func GetMockDB() (*sql.DB, error) {
	return sql.Open("mock", "")
}

// An implementation of the ResponseWriter interface that helps with testing.
//   It writes everything to stdout where it can be read from.
type MockResponseWriter struct{}

// Returns an empty header.
func (rw MockResponseWriter) Header() http.Header {
	return make(map[string][]string, 0)
}

// Writes the response code to stdout.
func (rw MockResponseWriter) WriteHeader(code int) {}

// Writes the response to stdout.
func (rw MockResponseWriter) Write(b []byte) (int, error) {

	// JSON to stdout
	fmt.Fprint(os.Stdout, string(b))

	return len(b), nil
}

// Abstracts the details of creating a testing context
func NewTestingContext(query string, p httprouter.Params, f ...eden.Middleware) *eden.Context {
	return eden.NewContext(
		MockResponseWriter{},
		&http.Request{
			URL: &url.URL{RawQuery: query},
		},
		p,
		f,
	)
}

// Tests an API endpoint. The function (or endpoint) to test, along with the context parameter
//   necessary to generate the desired response are passed in. The third parameter is the
//   output byte slice. All output from the api will be stored here.
func CallAPI(f func(*eden.Context), c *eden.Context, output *[]byte) {
	// Connect stdout to a pipe
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call function that is being tested
	f(c)

	// Read output written to stdout through pipe
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Close write end of the pipe
	w.Close()
	out := <-outputChan

	// Store output in []byte pointer
	*output = []byte(out)

	// Restore stdout
	os.Stdout = stdout
}
