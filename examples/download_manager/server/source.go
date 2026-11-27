// source.go — the places the manager reads the bytes of a file from.
//
// The manager splits a file in parts, so every source needs two things: the
// size of the file and the bytes of a range of it. An http server answers
// with the Range header, an ftp server with the restart command (REST).

package downloadmanager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// source reads the bytes of a file from a remote place.
type source interface {
	// Size returns the size of the file and tells whether the source can
	// read a range of it.
	Size(ctx context.Context) (int64, bool, error)
	// Open returns the reader of the bytes from start to end. An end below
	// zero means "until the end of the file".
	Open(ctx context.Context, start int64, end int64) (io.ReadCloser, error)
}

// rangeProbe is implemented by the sources that can tell, before the
// download starts, whether they can read a range of the file. A source that
// cannot is downloaded in one single part.
type rangeProbe interface {
	ProbeRange(ctx context.Context) (bool, error)
}

// newSource returns the source that can read the file at the URL: an http
// server or an ftp one.
func newSource(u *url.URL, client *http.Client) (source, error) {
	switch u.Scheme {
	case "http", "https":
		return newHTTPSource(u, client), nil
	case "ftp":
		return newFTPSource(u)
	}
	return nil, fmt.Errorf("unsupported scheme %q: only http, https and ftp are allowed", u.Scheme)
}

// httpSource reads a file from an http server. Every part asks for its own
// range.
type httpSource struct {
	url    string
	client *http.Client
	// total is the size the server announced: it is needed to accept a
	// plain 200 answer for a part that covers the whole file.
	total int64
}

// newHTTPSource reads the address of an http:// URL.
func newHTTPSource(u *url.URL, client *http.Client) *httpSource {
	return &httpSource{url: u.String(), client: client}
}

// Size asks the server for the first byte. A 206 answers with the total size
// and tells that the server honours the Range requests; a 200 means the
// server ignored the range and the file cannot be split.
func (s *httpSource) Size(ctx context.Context) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusPartialContent:
		total := parseContentRangeTotal(resp.Header.Get("Content-Range"))
		if total <= 0 {
			return 0, false, fmt.Errorf("the server did not announce the file size")
		}
		s.total = total
		return total, true, nil
	case http.StatusOK:
		if resp.ContentLength > 0 {
			s.total = resp.ContentLength
			return resp.ContentLength, false, nil
		}
		return 0, false, nil
	default:
		return 0, false, fmt.Errorf("unexpected HTTP %d from %v", resp.StatusCode, s.url)
	}
}

// Open returns the reader of a range of the file: from start to end. An end
// below zero means "until the end of the file".
func (s *httpSource) Open(ctx context.Context, start int64, end int64) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, err
	}
	if end >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	} else {
		// The size of the file is unknown: ask for everything from the byte
		// the part stopped at.
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	// A part that covers the whole file also accepts a plain 200 answer: some
	// servers ignore the Range header when the range is the whole file.
	wholeFile := start == 0 && (end < 0 || end >= s.total-1)
	if resp.StatusCode != http.StatusPartialContent && !(resp.StatusCode == http.StatusOK && wholeFile) {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d from the file server", resp.StatusCode)
	}
	return resp.Body, nil
}

// redactURL hides the password of an ftp URL: the URL is shown in the app
// and written in the log of the server.
func redactURL(u *url.URL) string {
	if u.User == nil {
		return u.String()
	}
	if _, ok := u.User.Password(); !ok {
		return u.String()
	}
	clone := *u
	clone.User = url.UserPassword(u.User.Username(), "***")
	return clone.String()
}

// hasCredentials tells whether the URL carries a password.
func hasCredentials(u *url.URL) bool {
	if u.User == nil {
		return false
	}
	_, ok := u.User.Password()
	return ok
}
