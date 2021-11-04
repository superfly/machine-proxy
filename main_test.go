package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSimpleCases(t *testing.T) {
	cases := []struct {
		fn   http.HandlerFunc // upstream handler
		code int              // expected status code
		body string           // expected response body
	}{
		0: {
			// we create an "upstream" server that renders 200 (hi\n)
			fn: func(w http.ResponseWriter, _ *http.Request) {
				io.WriteString(w, "hi\n")
			},
			code: http.StatusOK,
			body: "hi\n",
		},
		1: {
			// we create an "upstream" server that renders 200 (hi\n)
			fn: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusExpectationFailed)
				io.WriteString(w, "panos")
			},
			code: http.StatusExpectationFailed,
			body: "panos",
		},
	}

	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			runTest(t, kase.fn, kase.code, kase.body)
		})
	}
}

func runTest(t *testing.T, fn http.HandlerFunc, code int, body string) {
	upstreamServer, proxyServer := setupTest(t, fn)
	defer upstreamServer.Close()
	defer proxyServer.Close()

	assert(t, proxyServer, code, body)
}

func setupTest(t *testing.T, fn http.HandlerFunc) (upstreamSrv, proxySrv *httptest.Server) {
	t.Helper()

	upstreamSrv = httptest.NewServer(http.HandlerFunc(fn))

	// we create a proxy server for this upstream
	p := &proxy{
		upstream: upstreamSrv.Listener.Addr().String(),
	}

	proxySrv = httptest.NewServer(p)

	return
}

func assert(t *testing.T, proxy *httptest.Server, code int, body string) {
	t.Helper()

	// we make a request against the proxy
	req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed making request: %v", err)
	}
	defer res.Body.Close()

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading response: %v", err)
	}

	if res.StatusCode != code {
		t.Errorf("expected status code to be %d, got %d", code, res.StatusCode)
	}

	if s := string(got); body != s {
		t.Errorf(`expected body to be %q, got %q`, body, s)
	}
}
