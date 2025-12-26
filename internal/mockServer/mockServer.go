package mockServer

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Method string

const (
	DELETE Method = "DELETE"
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
)

func (m Method) String() string { return string(m) }

type Path string

func (p Path) String() string { return string(p) }

func New(t *testing.T) *Server {

	server := Server{
		config: &config{},
	}

	// --- HTTPS part ---
	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		assert.FailNow(t, "Failed to generate TLS certificate: "+err.Error())
	}

	// Load certificate
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		assert.FailNow(t, "Failed to load TLS key pair: "+err.Error())
	}

	server.listener, err = tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		assert.FailNow(t, "Uncaught error starting mock server: "+err.Error())
	}
	go func() { // http.Serve blocks, but only this goroutine will wait
		_ = http.Serve(server.listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			server.config.handle(w, r)
		}))
	}()

	return &server
}

func generateSelfSignedCert() (certPEM []byte, keyPEM []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	tmpl := x509.Certificate{
		DNSNames:     []string{"localhost"},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		NotAfter:     time.Now().Add(1 * time.Hour),
		NotBefore:    time.Now(),
		SerialNumber: big.NewInt(1)}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certBuf := &bytes.Buffer{}
	pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyBuf := &bytes.Buffer{}
	pem.Encode(keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}
