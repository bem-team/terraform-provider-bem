package logging_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"

	"github.com/bem-team/terraform-provider-bem/internal/logging"
)

// The body-restoring behaviour is the load-bearing part of this package: the
// middleware sits in the SDK's request chain, so a consumed-and-not-restored
// body means the request is sent empty or the response decodes to nothing,
// and only when TF_LOG is on.

func TestLogRequestRestoresBody(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &out)

	const body = `{"displayName":"Test Function"}`
	req := httptest.NewRequest(http.MethodPatch, "https://api.example.com/v3/functions/test", strings.NewReader(body))
	req.Header.Set("X-Api-Key", "secret")

	if err := logging.LogRequest(ctx, req); err != nil {
		t.Fatalf("LogRequest: %v", err)
	}

	got, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading body after LogRequest: %v", err)
	}
	if string(got) != body {
		t.Errorf("body after LogRequest = %q, want %q", got, body)
	}

	logged := decodeLog(t, &out)
	for _, want := range []string{"PATCH", "/v3/functions/test", body} {
		if !strings.Contains(logged, want) {
			t.Errorf("log output missing %q; got:\n%s", want, logged)
		}
	}
}

func TestLogRequestWithNilBody(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &out)

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v3/functions", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if req.Body != nil {
		t.Fatal("precondition failed: expected a nil body on a GET built with no body")
	}

	if err := logging.LogRequest(ctx, req); err != nil {
		t.Fatalf("LogRequest: %v", err)
	}
	if logged := decodeLog(t, &out); !strings.Contains(logged, "/v3/functions") {
		t.Errorf("log output missing the request path; got:\n%s", logged)
	}
}

// Header names are lowercased on the way into the log. This is cosmetic, but
// it is the only normalization the package does, so it is worth pinning.
func TestLogRequestLowercasesHeaderNames(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &out)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/v3/functions", nil)
	req.Header.Set("X-Api-Key", "secret")

	if err := logging.LogRequest(ctx, req); err != nil {
		t.Fatalf("LogRequest: %v", err)
	}
	if logged := decodeLog(t, &out); !strings.Contains(logged, "> x-api-key: secret") {
		t.Errorf("expected a lowercased header line; got:\n%s", logged)
	}
}

func TestLogResponseRestoresBody(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &out)

	const body = `{"name":"test","version":2}`
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	if err := logging.LogResponse(ctx, resp); err != nil {
		t.Fatalf("LogResponse: %v", err)
	}

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body after LogResponse: %v", err)
	}
	if string(got) != body {
		t.Errorf("body after LogResponse = %q, want %q", got, body)
	}

	logged := decodeLog(t, &out)
	for _, want := range []string{"200 OK", "content-type: application/json", body} {
		if !strings.Contains(logged, want) {
			t.Errorf("log output missing %q; got:\n%s", want, logged)
		}
	}
}

// tflog emits one JSON object per line with the human-readable text under
// "@message"; assert against that rather than the escaped JSON envelope.
func decodeLog(t *testing.T, out *bytes.Buffer) string {
	t.Helper()

	entries, err := tflogtest.MultilineJSONDecode(out)
	if err != nil {
		t.Fatalf("decoding log output: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no log entries were emitted")
	}

	var sb strings.Builder
	for _, entry := range entries {
		msg, ok := entry["@message"].(string)
		if !ok {
			t.Fatalf("log entry has no string @message: %v", entry)
		}
		sb.WriteString(msg)
		sb.WriteString("\n")
	}
	return sb.String()
}

// A body that fails mid-read must surface the error rather than be swallowed.
func TestLogResponsePropagatesBodyReadError(t *testing.T) {
	t.Parallel()

	ctx := tflogtest.RootLogger(t.Context(), io.Discard)
	sentinel := errors.New("boom")

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		Header:     http.Header{},
		Body:       io.NopCloser(errReader{sentinel}),
	}

	if err := logging.LogResponse(ctx, resp); !errors.Is(err, sentinel) {
		t.Errorf("LogResponse error = %v, want %v", err, sentinel)
	}
}

func TestMiddlewarePassesRequestAndResponseThrough(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &out)

	const reqBody = `{"displayName":"Test"}`
	const respBody = `{"name":"test"}`

	req := httptest.NewRequest(http.MethodPost, "https://api.example.com/v3/functions", strings.NewReader(reqBody))

	var seenByNext string
	mw := logging.Middleware(ctx)
	resp, err := mw(req, func(r *http.Request) (*http.Response, error) {
		// The next hop must still see the full body.
		b, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			return nil, readErr
		}
		seenByNext = string(b)
		return &http.Response{
			Status:     "201 Created",
			StatusCode: http.StatusCreated,
			Proto:      "HTTP/1.1",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(respBody)),
		}, nil
	})
	if err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if seenByNext != reqBody {
		t.Errorf("next handler saw body %q, want %q", seenByNext, reqBody)
	}

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}
	if string(got) != respBody {
		t.Errorf("caller saw response body %q, want %q", got, respBody)
	}
}

// A nil request must not panic, and a transport error must reach the caller
// unchanged rather than being replaced by a logging error.
func TestMiddlewareHandlesNilRequestAndTransportError(t *testing.T) {
	t.Parallel()

	ctx := tflogtest.RootLogger(t.Context(), io.Discard)
	sentinel := errors.New("dial tcp: connection refused")

	mw := logging.Middleware(ctx)
	resp, err := mw(nil, func(r *http.Request) (*http.Response, error) {
		if r != nil {
			t.Errorf("next handler received %v, want nil", r)
		}
		return nil, sentinel
	})
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want %v", err, sentinel)
	}
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }
