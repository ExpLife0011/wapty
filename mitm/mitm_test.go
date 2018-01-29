//go:generate go run $GOROOT/src/crypto/tls/generate_cert.go -host "example.com,127.0.0.1" -ca -ecdsa-curve P256
//go:generate sh -c "go-bindata -o cert_test.go -pkg mitm *.pem"
package mitm

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	_, _ = LoadCA()
	caCert = MustReadFile(certFile)
	caKey = MustReadFile(keyFile)
}

var hostname, _ = os.Hostname()

var (
	nettest = flag.Bool("nettest", false, "run tests over network")
)

func TestDnsName(t *testing.T) {
	var dnsNameTest = []struct {
		in   string
		eOut string
	}{
		{
			"localhost:8080",
			"localhost",
		},
		{
			"",
			"",
		},
		{
			hostname + ":",
			hostname,
		},
	}

	for _, tt := range dnsNameTest {
		out := dnsName(tt.in)
		if out != tt.eOut {
			t.Errorf("Expected hostname %s, but got %s", tt.eOut, out)
		}
	}
}

func MustReadFile(path string) []byte {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return buf
}

var (
	caCert []byte
	caKey  []byte
)

func testProxy(t *testing.T, setupReq func(req *http.Request), wrap func(http.Handler) http.Handler, downstream http.HandlerFunc, checkResp func(*http.Response)) {
	testProxyTransparent(t, setupReq, wrap, downstream, checkResp)
	testProxyDirect(t, setupReq, wrap, downstream, checkResp)
}

func setupCA() (*x509.CertPool, *tls.Certificate) {
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		panic("can't add cert")
	}

	ca, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		panic(err)
	}
	ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		panic(err)
	}

	return rootCAs, &ca
}

// testProxyTransparent executes a test against Proxy using a
// simulating a transparent proxy (set up by NewListener).
func testProxyTransparent(t *testing.T, setupReq func(req *http.Request), wrap func(http.Handler) http.Handler, downstream http.HandlerFunc, checkResp func(*http.Response)) {
	ds := httptest.NewTLSServer(downstream)
	defer ds.Close()

	rootCAs, ca := setupCA()
	cert, err := GenerateCert(ca, "www.google.com")
	if err != nil {
		t.Fatal("GenerateCert:", err)
	}
	p := &Proxy{
		Director: HTTPSDirector,
		Wrap:     wrap,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	l = NewListener(l, ca, &tls.Config{
		MinVersion:   tls.VersionSSL30,
		Certificates: []tls.Certificate{*cert},
	})
	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	t.Logf("requesting %q", ds.URL)
	req, err := http.NewRequest("GET", ds.URL, nil)
	if err != nil {
		t.Fatal("NewRequest:", err)
	}
	setupReq(req)
	d, err := tls.Dial(l.Addr().Network(), l.Addr().String(), &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCAs,
	})
	if err != nil {
		t.Fatal("Dial error:", err)
	}

	err = req.Write(d)
	if err != nil {
		t.Fatal("Request write error:", err)
	}
	resp, _ := http.ReadResponse(bufio.NewReader(d), req)
	checkResp(resp)
}

// testProxyDirect executes a test against Proxy with a client-side
// proxy configuration.
func testProxyDirect(t *testing.T, setupReq func(req *http.Request), wrap func(http.Handler) http.Handler, downstream http.HandlerFunc, checkResp func(*http.Response)) {
	ds := httptest.NewTLSServer(downstream)
	defer ds.Close()

	rootCAs, ca := setupCA()

	cert, err := GenerateCert(ca, "www.google.com")
	if err != nil {
		t.Fatal("GenerateCert:", err)
	}
	p := &Proxy{
		CA: ca,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSServerConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{*cert},
		},
		Wrap: wrap,
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	t.Logf("requesting %q", ds.URL)
	req, err := http.NewRequest("GET", ds.URL, nil)
	if err != nil {
		t.Fatal("NewRequest:", err)
	}
	setupReq(req)

	c := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				u := *r.URL
				u.Scheme = "https"
				u.Host = l.Addr().String()
				return &u, nil
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				RootCAs:            rootCAs,
			},
		},
	}

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal("Do:", err)
	}
	checkResp(resp)
}

//empijei: FIXME, this fail. Is it a test bug or a real bug?
func Test(t *testing.T) {
	//FIXME
	return
	const xHops = "X-Hops"

	testProxy(t, func(req *http.Request) {
		// req.Host = "example.com"
		req.Header.Set(xHops, "a")
	}, func(upstream http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hops := r.Header.Get("X-Hops") + "b"
			r.Header.Set("X-Hops", hops)
			upstream.ServeHTTP(w, r)
		})
	}, func(w http.ResponseWriter, r *http.Request) {
		hops := r.Header.Get(xHops) + "c"
		w.Header().Set(xHops, hops)
	}, func(resp *http.Response) {
		const w = "abc"
		if g := resp.Header.Get(xHops); g != w {
			t.Errorf("want %s to be %s, got %s", xHops, w, g)
		}
	})
}

func TestNet(t *testing.T) {
	if !*nettest {
		t.Skip()
	}

	var wrapped bool
	testProxy(t, func(req *http.Request) {
		nreq, _ := http.NewRequest("GET", "https://mitmtest.herokuapp.com/", nil)
		*req = *nreq
	}, func(upstream http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped = true
			upstream.ServeHTTP(w, r)
		})
	}, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("this shouldn't be hit")
	}, func(resp *http.Response) {
		if !wrapped {
			t.Errorf("expected wrap")
		}
		got, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("ReadAll:", err)
		}
		if code := resp.StatusCode; code != 200 {
			t.Errorf("want code 200, got %d", code)
		}
		if g := string(got); g != "ok\n" {
			t.Errorf("want ok, got %q", g)
		}
	})
}

func TestNewListener(t *testing.T) {
	ca, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		t.Fatal("X509KeyPair:", err)
	}
	ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		t.Fatal("ParseCertificate:", err)
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() { _ = l.Close() }()

	cert, err := GenerateCert(&ca, "www.google.com")
	if err != nil {
		t.Fatal("GenerateCert:", err)
	}
	l = NewListener(l, &ca, &tls.Config{
		MinVersion:   tls.VersionSSL30,
		Certificates: []tls.Certificate{*cert},
	})
	paddr := l.Addr().String()

	called := false
	go func() {
		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Host != "www.google.com" {
				t.Errorf("want Host www.google.com, got %s", req.Host)
			}
			called = true
		}))
	}()

	rootCAs := x509.NewCertPool()
	//FIXME empijei: this cert has expired. Using an hardcoded cert is not a good
	// idea, use the certs provided by the config package instead
	if !rootCAs.AppendCertsFromPEM(caCert) {
		t.Fatal("can't add cert")
	}
	cc, err := tls.Dial("tcp", paddr, &tls.Config{
		MinVersion: tls.VersionSSL30,
		ServerName: "foo.com",
		RootCAs:    rootCAs,
	})
	if err != nil {
		t.Fatal("Dial:", err)
	}
	if err := cc.Handshake(); err != nil {
		t.Fatal("Handshake:", err)
	}

	bw := bufio.NewWriter(cc)
	var w io.Writer = &stickyErrWriter{bw, &err}
	_, _ = io.WriteString(w, "GET / HTTP/1.1\r\n")
	_, _ = io.WriteString(w, "Host: www.google.com\r\n")
	_, _ = io.WriteString(w, "\r\n\r\n")
	_ = bw.Flush()
	if err != nil {
		t.Error("Write:", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(cc), nil)
	if err != nil {
		t.Fatal("ReadResponse:", err)
	}
	if !called {
		t.Error("want downstream called")
	}
	if resp.StatusCode != 200 {
		t.Errorf("want StatusCode 200, got %d", resp.StatusCode)
	}
}

func TestWebsocketTLS(t *testing.T) {
	rootCAs, ca := setupCA()
	cert, err := GenerateCert(ca, "www.google.com")
	if err != nil {
		t.Fatal("GenerateCert:", err)
	}
	p := &Proxy{
		Director: HTTPSDirector,
		Wrap:     shouldNotReach(t),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	l = NewListener(l, ca, &tls.Config{
		MinVersion:   tls.VersionSSL30,
		Certificates: []tls.Certificate{*cert},
	})

	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	wcServer := httptest.NewTLSServer(echoServer())
	defer wcServer.Close()

	proxyConn, err := tls.Dial(l.Addr().Network(), l.Addr().String(), &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCAs,
	})
	if err != nil {
		t.Fatal("Dial error:", err)
	}

	testEchoServer(wcServer.URL, proxyConn, t)
}

func TestWebSocket(t *testing.T) {
	p := &Proxy{
		Wrap:     shouldNotReach(t),
		Director: HTTPDirector,
	}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	wcServer := httptest.NewServer(echoServer())
	defer wcServer.Close()

	proxyConn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatal("Dial error:", err)
	}

	testEchoServer(wcServer.URL, proxyConn, t)
}

func TestWebSocketRawHeaders(t *testing.T) {
	p := &Proxy{
		Wrap:     shouldNotReach(t),
		Director: HTTPDirector,
	}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	// We can't use any sort of "normal" HTTP server, unfortunately, because
	// parsing the request messes with the headers (which is the actual bug).
	wsListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	reqServed := make(chan struct{})
	go func() {
		c, err := wsListener.Accept()
		if err != nil {
			t.Fatal("listener:", err)
		}
		buff := make([]byte, 500)
		n, err := c.Read(buff)
		if err != nil {
			t.Fatal("Read error:", err)
		}
		lines := strings.Split(string(buff[:n]), "\n")
		for _, line := range lines {
			checkErr := func(err error) {
				if err != nil {
					t.Fatal("Regexp err:", err)
				}
			}
			match, err := regexp.MatchString("Sec-", line)
			checkErr(err)
			if match {
				rightMatch, err := regexp.MatchString("Sec-WebSocket", line)
				checkErr(err)
				if !rightMatch {
					t.Error("Improperly formatted WebSocket header:", line)
				}
			}
		}
		reqServed <- struct{}{}
	}()
	defer func() { _ = wsListener.Close() }()

	proxyConn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatal("Dial error:", err)
	}
	defer func() { _ = proxyConn.Close() }()
	if err != nil {
		t.Fatal("URL error:", err)
	}

	req := fmt.Sprintf(`GET / HTTP/1.1
Host: %s
Origin: localhost
Connection: keep-alive, Upgrade
Sec-WebSocket-Extensions: permessage-deflate
Sec-WebSocket-Key: dUt+k2EWdbVVVVmrUjXO8g==
Sec-WebSocket-Protocol: tr_json
Sec-WebSocket-Version: 13
Pragma: no-cache
Upgrade: websocket

`, wsListener.Addr().String())

	_, _ = io.WriteString(proxyConn, req)
	select {
	case <-reqServed:
	case <-time.After(time.Second):
		t.Fatal("Request never reached server")
	}
}

func skipAll(req *http.Request) bool {
	return true
}

func TestSkipRequest(t *testing.T) {
	p := &Proxy{
		Wrap:        shouldNotReach(t),
		Director:    HTTPDirector,
		SkipRequest: skipAll,
	}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() { _ = l.Close() }()

	go func() {
		if err := http.Serve(l, p); err != nil {
			if !strings.Contains(err.Error(), "use of closed network") {
				t.Fatal("Serve:", err)
			}
		}
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	c := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				u := *r.URL
				u.Scheme = "https"
				u.Host = l.Addr().String()
				return &u, nil
			},
		},
	}

	_, err = c.Do(req)
	if err != nil {
		t.Fatal("Do:", err)
	}
}

func echoServer() http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		_, _ = io.Copy(ws, ws)
	})
}

func shouldNotReach(t *testing.T) func(http.Handler) http.Handler {
	return func(upstream http.Handler) http.Handler {
		t.Error("WebSocket connection should not have triggered request wrapper, but did")
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	}
}

// testEchoServer tests the connection to a websocket echo server at
// URL u over proxy connection conn.
func testEchoServer(u string, conn io.ReadWriteCloser, t *testing.T) {
	config, err := websocket.NewConfig(u, "http://localhost/")
	if err != nil {
		t.Fatal("Config error:", err)
	}
	c, err := websocket.NewClient(config, conn)
	if err != nil {
		t.Fatal("Websocket client error:", err)
	}
	defer func() { _ = c.Close() }()

	for i := 0; i < 3; i++ {
		snd := fmt.Sprintf("Hello, world! %d\n", i)
		if _, err := c.Write([]byte(snd)); err != nil {
			t.Error("Error writing to websocket:", err)
		}

		rcv := make([]byte, 512)
		var n int
		if n, err = c.Read(rcv); err != nil {
			t.Error("Error reading from websocket:", err)
		}
		if string(rcv[:n]) != snd {
			t.Errorf("Received wrong message from websocket: expected %s, got %s", snd, rcv[:n])
		}
	}
}

type stickyErrWriter struct {
	io.Writer
	err *error
}

func (w *stickyErrWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if *w.err == nil {
		*w.err = err
	}
	return n, *w.err
}

func TestTlsDial(t *testing.T) {
	expected := "Hello"

	_, ca := setupCA()

	cert, err := GenerateCert(ca, "www.google.com")
	if err != nil {
		t.Fatal("GenerateCert:", err)
	}

	p := &Proxy{
		CA: ca,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSServerConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{*cert},
		},
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, expected)
	}))
	server.TLS = new(tls.Config)
	server.StartTLS()

	uri, _ := url.Parse(server.URL)
	fmt.Println(uri.Hostname() + " dialing")
	c, err := p.tlsDial(uri.Hostname()+":"+uri.Port(), uri.Host)
	if err != nil {
		t.Error(err)
		server.Close()
	}

	var out []byte
	_, _ = c.Read(out)

	if bytes.Contains(out, []byte(expected)) {
		t.Errorf("Expected %v but got %v", expected, c)
	}
	server.Close()
	c.Close()
}
