package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func PrivateRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	keyDER := x509.MarshalPKCS1PrivateKey(key)
	encodeAsPEM(keyDER, "server.key", "RSA PRIVATE KEY")
	log.Println("Private RSA Key successfully created")
	return key
}

func SelfSignedCertificate(key *rsa.PrivateKey) {
	tmpl := &x509.Certificate{
		SerialNumber: randomBigInt(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 1, 14),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		panic(err)
	}
	encodeAsPEM(certDER, "server.crt", "CERTIFICATE")
	log.Println("x509 Certificate successfully created")
}

// Temporary

func TemporaryPrivateRSAKey() (*rsa.PrivateKey, *os.File) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	keyDER := x509.MarshalPKCS1PrivateKey(key)
	file := encodeAsTempPEM(keyDER, "/etc/ssl/private", "server.key", "RSA PRIVATE KEY")
	log.Println("Private RSA Key successfully created")
	return key, file
}

func TemporarySelfSignedCertificate(key *rsa.PrivateKey) *os.File {
	tmpl := &x509.Certificate{
		SerialNumber: randomBigInt(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 1, 14),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		panic(err)
	}
	file := encodeAsTempPEM(certDER, "/etc/ssl/certs", "server.crt", "CERTIFICATE")
	log.Println("x509 Certificate successfully created")
	return file
}

// Self Signed TLS Orchestrator

type SelfSignedTLS struct {
	keyFile  *os.File
	certFile *os.File
}

func NewSelfSignedTLS() (*SelfSignedTLS, func() error) {
	sstls := new(SelfSignedTLS)
	key, keyFile := TemporaryPrivateRSAKey()
	certFile := TemporarySelfSignedCertificate(key)
	sstls.keyFile = keyFile
	sstls.certFile = certFile

	return sstls, sstls.Close

}

func (s SelfSignedTLS) KeyFilename() string {
	return s.keyFile.Name()
}

func (s SelfSignedTLS) CertFilename() string {
	return s.certFile.Name()
}

func (s *SelfSignedTLS) Close() error {
	err := s.keyFile.Close()
	if err != nil {
		return err
	}
	err = os.Remove(s.keyFile.Name())
	if err != nil {
		return err
	}
	log.Printf("temporary keyfile removed: %s", s.keyFile.Name())
	err = s.certFile.Close()
	if err != nil {
		return err
	}
	err = os.Remove(s.certFile.Name())
	if err != nil {
		return err
	}
	log.Printf("temporary certfile removed: %s", s.certFile.Name())

	return nil
}

// Helpers

func randomBigInt() *big.Int {
	//Max random value, a 130-bits integer, i.e 2^130 - 1
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	//Generate cryptographically strong pseudo-random between 0 - max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return n
}

func encodeAsPEM(crypto []byte, filename, pemType string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(file, &pem.Block{
		Type:  pemType,
		Bytes: crypto,
	})
	if err != nil {
		panic(err)
	}
}

func encodeAsTempPEM(crypto []byte, dir, pattern, pemType string) *os.File {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(file, &pem.Block{
		Type:  pemType,
		Bytes: crypto,
	})
	if err != nil {
		panic(err)
	}
	return file
}
