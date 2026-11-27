package downloadmanager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ftpTestServer is a small FTP server for the tests: it knows the login, the
// files it serves and the restart command. It is enough to exercise the
// client without a real server on the network.
type ftpTestServer struct {
	ln        net.Listener
	files     map[string][]byte
	username  string
	password  string
	noRestart bool
}

func newFTPTestServer(t *testing.T, files map[string][]byte) *ftpTestServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &ftpTestServer{
		ln:       ln,
		files:    files,
		username: "tester",
		password: "secret",
	}
	go s.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

// addr is the address of the server: host and port.
func (s *ftpTestServer) addr() string {
	return s.ln.Addr().String()
}

// urlOf returns the ftp URL of a file of the server.
func (s *ftpTestServer) urlOf(path string) string {
	return "ftp://" + s.username + ":" + s.password + "@" + s.addr() + path
}

func (s *ftpTestServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

// handle serves one control connection.
func (s *ftpTestServer) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	reply := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\r\n", args...)
		_ = w.Flush()
	}
	reply("220 the test ftp server is ready")

	rest := int64(0)
	var dataLn net.Listener
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		fields := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(fields[0])
		arg := ""
		if len(fields) > 1 {
			arg = fields[1]
		}
		switch cmd {
		case "USER":
			if arg != s.username {
				reply("530 Login incorrect")
				continue
			}
			reply("331 Password required")
		case "PASS":
			if arg != s.password {
				reply("530 Login incorrect")
				continue
			}
			reply("230 Login successful")
		case "TYPE":
			reply("200 Type set to %s", arg)
		case "SIZE":
			data, ok := s.files[arg]
			if !ok {
				reply("550 %s: no such file", arg)
				continue
			}
			reply("213 %d", len(data))
		case "REST":
			if s.noRestart {
				reply("502 Command not implemented")
				continue
			}
			point, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
			if err != nil {
				reply("501 Bad restart point")
				continue
			}
			rest = point
			reply("350 Restarting at %d", point)
		case "PASV":
			if dataLn != nil {
				_ = dataLn.Close()
			}
			dl, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				reply("425 Cannot open the data connection")
				continue
			}
			dataLn = dl
			port := dl.Addr().(*net.TCPAddr).Port
			reply("227 Entering Passive Mode (127,0,0,1,%d,%d)", port/256, port%256)
		case "RETR":
			data, ok := s.files[arg]
			if !ok {
				reply("550 %s: no such file", arg)
				continue
			}
			if dataLn == nil {
				reply("425 Use PASV first")
				continue
			}
			reply("150 Opening the data connection")
			dl := dataLn
			dataLn = nil
			go func() {
				defer dl.Close()
				conn, err := dl.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				if rest < int64(len(data)) {
					_, _ = conn.Write(data[rest:])
				}
			}()
			reply("226 Transfer complete")
		case "QUIT":
			reply("221 Bye")
			return
		default:
			reply("502 Command not implemented")
		}
	}
}

// The client asks the server for the size of the file and reads a range of
// it: the bytes are the ones the server holds.
func TestTheFTPSourceReadsTheSizeOfTheFile(t *testing.T) {
	payload := bytes.Repeat([]byte("0123456789"), 100)
	srv := newFTPTestServer(t, map[string][]byte{"/pub/file.bin": payload})

	u, err := url.Parse(srv.urlOf("/pub/file.bin"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	src, err := newFTPSource(u)
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	size, _, err := src.Size(context.Background())
	if err != nil {
		t.Fatalf("size: %v", err)
	}
	if size != int64(len(payload)) {
		t.Fatalf("wrong size: %d", size)
	}
}

// A part is read with the restart command: the bytes are the ones of the
// range, not the ones of the beginning of the file.
func TestTheFTPSourceReadsAPartOfTheFile(t *testing.T) {
	payload := []byte("0123456789abcdefghij")
	srv := newFTPTestServer(t, map[string][]byte{"/pub/file.bin": payload})

	u, err := url.Parse(srv.urlOf("/pub/file.bin"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	src, err := newFTPSource(u)
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	reader, err := src.Open(context.Background(), 10, 14)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer reader.Close()
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "abcde" {
		t.Fatalf("wrong bytes: %q", got)
	}
}

// A server that does not know the restart command cannot split the file: the
// manager downloads it in one single part.
func TestTheFTPSourceWithoutTheRestartGivesOnePart(t *testing.T) {
	srv := newFTPTestServer(t, map[string][]byte{"/pub/file.bin": []byte("0123456789")})
	srv.noRestart = true

	u, err := url.Parse(srv.urlOf("/pub/file.bin"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	src, err := newFTPSource(u)
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	canRange, err := src.ProbeRange(context.Background())
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if canRange {
		t.Fatal("the server does not support the restart")
	}
}

// The manager downloads a file from an ftp server and puts the bytes back
// together: the merged file is the one of the server.
func TestTheManagerDownloadsAFileOverFTP(t *testing.T) {
	payload := bytes.Repeat([]byte("simpleui ftp "), 4096)
	srv := newFTPTestServer(t, map[string][]byte{"/pub/file.bin": payload})

	out := t.TempDir()
	mgr := &Manager{
		jobs:     map[string]*Job{},
		outDir:   out,
		workers:  2,
		maxParts: 256,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
	job := &Job{
		ID:      "job-ftp",
		State:   "running",
		dir:     filepath.Join(out, "job-ftp"),
		workers: 2,
	}
	f, err := mgr.planFile(job, srv.urlOf("/pub/file.bin"), "", 4096, false, prioNormal)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if f.PartCnt < 2 {
		t.Fatalf("the file was not split: %d part(s)", f.PartCnt)
	}
	if !f.Range {
		t.Fatal("the server supports the restart: the file must be split")
	}
	job.addFile(f)
	if err := os.MkdirAll(job.dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	job.start()
	defer job.cancelJob()

	deadline := time.Now().Add(20 * time.Second)
	for {
		job.mu.Lock()
		state := job.State
		job.mu.Unlock()
		if state == "done" || state == "error" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the download did not finish: %v", state)
		}
		time.Sleep(20 * time.Millisecond)
	}

	job.mu.Lock()
	state, merged, verify := job.State, f.Merged, f.Verify
	job.mu.Unlock()
	if state != "done" {
		t.Fatalf("the download failed: %v (%v)", state, f.Error)
	}
	if verify != "size matches the server" {
		t.Fatalf("wrong verification: %v", verify)
	}
	got, err := os.ReadFile(merged)
	if err != nil {
		t.Fatalf("read the merged file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("the merged file is not the one of the server: %d bytes", len(got))
	}
}

// A wrong password is refused: the download fails with the answer of the
// server.
func TestTheFTPLoginIsRefused(t *testing.T) {
	srv := newFTPTestServer(t, map[string][]byte{"/pub/file.bin": []byte("0123456789")})
	raw := "ftp://tester:wrong@" + srv.addr() + "/pub/file.bin"
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	src, err := newFTPSource(u)
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	if _, _, err := src.Size(context.Background()); err == nil {
		t.Fatal("the login should have been refused")
	}
}

// The password of an ftp URL is a login: it is accepted, but it never shows
// in the URL that the app displays.
func TestTheFTPPasswordIsAcceptedAndHidden(t *testing.T) {
	u, err := url.Parse("ftp://tester:secret@ftp.example.com/pub/file.bin")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := validateTarget(u.String(), nil); err != nil {
		t.Fatalf("the ftp url should be accepted: %v", err)
	}
	if !hasCredentials(u) {
		t.Fatal("the url should carry the password")
	}
	shown := redactURL(u)
	if strings.Contains(shown, "secret") {
		t.Fatalf("the password is visible: %v", shown)
	}
	if !strings.Contains(shown, "tester") {
		t.Fatalf("the user is missing: %v", shown)
	}
}

// The credentials of an http URL are still refused.
func TestTheHTTPCredentialsAreRefused(t *testing.T) {
	if _, err := validateTarget("http://user:pass@example.com/file.bin", nil); err == nil {
		t.Fatal("the credentials should have been refused")
	}
}

// The allow-list of the CLI applies to the host of an ftp url too.
func TestTheFTPHostIsCheckedAgainstTheAllowList(t *testing.T) {
	if _, err := validateTarget("ftp://ftp.example.com/file.bin", []string{"other.com"}); err == nil {
		t.Fatal("the host should have been refused")
	}
	if _, err := validateTarget("ftp://ftp.example.com/file.bin", []string{"example.com"}); err != nil {
		t.Fatalf("the host should have been accepted: %v", err)
	}
}
