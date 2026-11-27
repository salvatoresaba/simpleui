// ftp.go — a minimal FTP client for the download manager.
//
// The manager splits a file in parts, so the client needs two things: the size
// of the file (SIZE) and the bytes of a range of it (REST + RETR). The server
// must support the passive mode (PASV) and the restart (REST); a server
// without them gets the whole file in one part.
//
// The client is written here, without any dependency: the example must stay a
// module with no external requirement.

package downloadmanager

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ftpSource reads a file from an FTP server. Every part opens its own
// connection: the data connection of a transfer cannot be shared.
type ftpSource struct {
	host string // the host of the URL
	port string // the port of the control connection
	path string
	user string
	pass string
}

// newFTPSource reads the address of an ftp:// URL. The credentials are
// optional: without them the client logs in as anonymous.
func newFTPSource(u *url.URL) (*ftpSource, error) {
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("the ftp url has no host")
	}
	port := u.Port()
	if port == "" {
		port = "21"
	}
	user, pass := "anonymous", "anonymous@"
	if u.User != nil {
		if name := u.User.Username(); name != "" {
			user = name
		}
		if p, ok := u.User.Password(); ok {
			pass = p
		}
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	return &ftpSource{host: host, port: port, path: path, user: user, pass: pass}, nil
}

// addr is the address of the control connection.
func (s *ftpSource) addr() string {
	return net.JoinHostPort(s.host, s.port)
}

// ftpConn is a control connection of the FTP protocol.
type ftpConn struct {
	conn net.Conn
	r    *bufio.Reader
}

// dial opens the control connection and reads the greeting of the server.
func (s *ftpSource) dial(ctx context.Context) (*ftpConn, error) {
	d := net.Dialer{}
	c, err := d.DialContext(ctx, "tcp", s.addr())
	if err != nil {
		return nil, err
	}
	fc := &ftpConn{conn: c, r: bufio.NewReader(c)}
	if _, _, err := fc.readReply(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("ftp: %v", err)
	}
	return fc, nil
}

func (fc *ftpConn) close() {
	_ = fc.conn.Close()
}

// readReply reads one reply of the server: the code and the text.
func (fc *ftpConn) readReply() (int, string, error) {
	line, err := fc.r.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) < 3 {
		return 0, line, fmt.Errorf("bad reply %q", line)
	}
	code, cerr := strconv.Atoi(line[:3])
	if cerr != nil {
		return 0, line, fmt.Errorf("bad reply %q", line)
	}
	if len(line) > 3 && line[3] == '-' {
		// A reply over more lines ends with the code and a space.
		for {
			next, err := fc.r.ReadString('\n')
			if err != nil {
				return code, line, nil
			}
			next = strings.TrimRight(next, "\r\n")
			if len(next) >= 4 && next[:3] == line[:3] && next[3] == ' ' {
				break
			}
		}
	}
	return code, line, nil
}

// send writes a command and reads the reply.
func (fc *ftpConn) send(format string, args ...any) (int, string, error) {
	if _, err := fmt.Fprintf(fc.conn, format+"\r\n", args...); err != nil {
		return 0, "", err
	}
	return fc.readReply()
}

// mustSend writes a command that has to succeed.
func (fc *ftpConn) mustSend(format string, args ...any) (string, error) {
	code, reply, err := fc.send(format, args...)
	if err != nil {
		return reply, err
	}
	if code < 200 || code >= 300 {
		return reply, fmt.Errorf("ftp: %s", strings.TrimSpace(reply))
	}
	return reply, nil
}

// login opens the session: the user and the password, then the binary mode.
func (s *ftpSource) login(fc *ftpConn) error {
	code, reply, err := fc.send("USER %s", s.user)
	if err != nil {
		return err
	}
	if code == 331 {
		code, reply, err = fc.send("PASS %s", s.pass)
		if err != nil {
			return err
		}
	}
	if code != 230 && code != 202 {
		return fmt.Errorf("ftp: the login was refused (%s)", strings.TrimSpace(reply))
	}
	if _, err := fc.mustSend("TYPE I"); err != nil {
		return err
	}
	return nil
}

// Size returns the size of the file and tells that the server can read a
// range of it (the restart command).
func (s *ftpSource) Size(ctx context.Context) (int64, bool, error) {
	fc, err := s.dial(ctx)
	if err != nil {
		return 0, false, err
	}
	defer fc.close()
	if err := s.login(fc); err != nil {
		return 0, false, err
	}
	reply, err := fc.mustSend("SIZE %s", s.path)
	if err != nil {
		return 0, false, fmt.Errorf("ftp: the server does not know the size of %v: %v", s.path, err)
	}
	idx := strings.Index(reply, " ")
	if idx < 0 {
		return 0, false, fmt.Errorf("ftp: bad size reply %q", strings.TrimSpace(reply))
	}
	size, err := strconv.ParseInt(strings.TrimSpace(reply[idx+1:]), 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("ftp: bad size reply %q", strings.TrimSpace(reply))
	}
	return size, true, nil
}

// ProbeRange tells whether the server can start a transfer in the middle of
// the file: the restart command (REST). A server that refuses it gets the
// whole file in one single part.
func (s *ftpSource) ProbeRange(ctx context.Context) (bool, error) {
	fc, err := s.dial(ctx)
	if err != nil {
		return false, err
	}
	defer fc.close()
	if err := s.login(fc); err != nil {
		return false, err
	}
	code, reply, err := fc.send("REST 1")
	if err != nil {
		return false, err
	}
	switch {
	case code == 350:
		// The server accepted the restart point.
		return true, nil
	case code >= 500:
		// The server does not know the command: the file is one part.
		return false, nil
	default:
		return false, fmt.Errorf("ftp: %s", strings.TrimSpace(reply))
	}
}

// Open returns the reader of a range of the file: from start to end. An end
// below zero means "until the end of the file".
func (s *ftpSource) Open(ctx context.Context, start int64, end int64) (io.ReadCloser, error) {
	fc, err := s.dial(ctx)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			fc.close()
		}
	}()
	if err := s.login(fc); err != nil {
		return nil, err
	}
	if start > 0 {
		// The restart command tells the server where the transfer starts.
		code, reply, err := fc.send("REST %d", start)
		if err != nil {
			return nil, err
		}
		if code != 350 {
			return nil, fmt.Errorf("ftp: the server does not support the restart (%s)", strings.TrimSpace(reply))
		}
	}
	port, err := ftpPassive(fc)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	// The address of the data connection is the one of the control
	// connection: the host the server writes in the reply is often wrong
	// behind a router.
	data, err := d.DialContext(ctx, "tcp", net.JoinHostPort(s.host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	code, reply, err := fc.send("RETR %s", s.path)
	if err != nil {
		_ = data.Close()
		return nil, err
	}
	if code < 100 || code >= 300 {
		_ = data.Close()
		return nil, fmt.Errorf("ftp: %s", strings.TrimSpace(reply))
	}
	ok = true
	var r io.Reader = data
	if end >= 0 {
		// The transfer goes to the end of the file: only the bytes of the
		// part are read.
		r = io.LimitReader(data, end-start+1)
	}
	return &ftpReader{r: r, data: data, ctrl: fc}, nil
}

// ftpReader is the reader of a transfer: it closes the data connection and
// the control connection together.
type ftpReader struct {
	r    io.Reader
	data net.Conn
	ctrl *ftpConn
}

func (f *ftpReader) Read(p []byte) (int, error) { return f.r.Read(p) }

func (f *ftpReader) Close() error {
	_ = f.data.Close()
	f.ctrl.close()
	return nil
}

// ftpPassive asks the server for the port of the data connection and returns
// it. The host of the data connection is the one of the control connection.
func ftpPassive(fc *ftpConn) (int, error) {
	reply, err := fc.mustSend("PASV")
	if err != nil {
		return 0, err
	}
	start := strings.Index(reply, "(")
	end := strings.Index(reply, ")")
	if start < 0 || end < start {
		return 0, fmt.Errorf("ftp: bad passive reply %q", strings.TrimSpace(reply))
	}
	nums := strings.Split(reply[start+1:end], ",")
	if len(nums) < 6 {
		return 0, fmt.Errorf("ftp: bad passive reply %q", strings.TrimSpace(reply))
	}
	values := make([]int, 6)
	for i := 0; i < 6; i++ {
		v, err := strconv.Atoi(strings.TrimSpace(nums[i]))
		if err != nil {
			return 0, fmt.Errorf("ftp: bad passive reply %q", strings.TrimSpace(reply))
		}
		values[i] = v
	}
	return values[4]*256 + values[5], nil
}
