package edgeproxy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// BuildTLSConfig returns a production-hardened TLS configuration for the edge proxy.
//
// Security properties:
//   - Minimum TLS version: TLS 1.2 (rejects TLS 1.0 and 1.1)
//   - Preferred: TLS 1.3 (negotiated automatically when both sides support it)
//   - ALPN protocols: h2, http/1.1 — enables HTTP/2 negotiation
//   - Preferred curves: X25519, P-256 (modern, fast ECDH)
//   - Cipher suites: restricted to AEAD ciphers only
//
// If certFile and keyFile are both empty, a self-signed certificate is
// auto-generated (suitable for development and testing).
func BuildTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	var cert tls.Certificate
	var err error

	if certFile == "" && keyFile == "" {
		// Auto-generate a self-signed certificate for development.
		cert, err = generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("edgeproxy: generating self-signed cert: %w", err)
		}
	} else {
		cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("edgeproxy: loading TLS certificate: %w", err)
		}
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,

		// ALPN: advertise HTTP/2 and HTTP/1.1 so clients can negotiate h2.
		// golang.org/x/net/http2 sets this automatically when the server uses
		// http2.ConfigureServer(), but we set it here explicitly for clarity.
		NextProtos: []string{"h2", "http/1.1"},

		// Preferred elliptic curves — X25519 is fastest, P-256 as fallback.
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},

		// Restrict to AEAD cipher suites only (no RC4, no CBC with SHA-1).
		// TLS 1.3 cipher suites are not configurable — this only affects TLS 1.2.
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},

		// Prefer server cipher order so we control the negotiated cipher.
		PreferServerCipherSuites: true,
	}

	return cfg, nil
}

// generateSelfSignedCert creates an ECDSA P-256 self-signed certificate
// valid for localhost and 127.0.0.1, good for 10 years.
// The private key is generated in-memory and never written to disk.
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generating ECDSA key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generating serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{"AgentOS Edge (dev)"},
			OrganizationalUnit: []string{"AegisFlow"},
			CommonName:         "localhost",
		},
		NotBefore: time.Now().Add(-time.Minute), // small backdate for clock skew
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		DNSNames:    []string{"localhost", "aegisflow", "agentOS"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},

		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshaling private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}
