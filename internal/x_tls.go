package x

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
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	C_CERT_BASEPATH string = "/var/lib/nodelocker/certs/"
)

func generateCertificate() error {
	// Ensure certificate directory exists
	if err := os.MkdirAll(C_CERT_BASEPATH, 0700); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate a private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"internal"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create a self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Save private key to a file
	keyFile, err := os.Create(C_CERT_BASEPATH + "private-key.pem")
	if err != nil {
		return err
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			fmt.Printf("Error closing key file: %v\n", err)
		}
	}()

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return err
	}

	// Save certificate to a file
	certFile, err := os.Create(C_CERT_BASEPATH + "certificate.pem")
	if err != nil {
		return err
	}
	defer func() {
		if err := certFile.Close(); err != nil {
			fmt.Printf("Error closing cert file: %v\n", err)
		}
	}()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return err
	}

	return nil
}

func ServeTLS(r *chi.Mux) {

	err := generateCertificate()
	if err != nil {
		fmt.Printf("%s Error generating certificate: %s\n", C_FAILED, err)
		return
	} else {
		fmt.Printf("%s Certificate and private key generated successfully.\n", C_SUCCESS)
	}

	// Load the private key and certificate
	privateKey, err := os.ReadFile(C_CERT_BASEPATH + "private-key.pem")
	if err != nil {
		fmt.Println("Error reading private key:", err)
		return
	}

	cert, err := os.ReadFile(C_CERT_BASEPATH + "certificate.pem")
	if err != nil {
		fmt.Println("Error reading certificate:", err)
		return
	}

	// Create a TLS certificate
	tlsCert, err := tls.X509KeyPair(cert, privateKey)
	if err != nil {
		fmt.Println("Error creating TLS certificate:", err)
		return
	}

	// Configure the Chi router
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("nodelocker")); err != nil {
			fmt.Printf("Error writing response: %v\n", err)
		}
	})

	// Create a server with TLS configuration
	server := &http.Server{
		Addr:         "0.0.0.0:3000",
		Handler:      r,
		TLSConfig:    &tls.Config{Certificates: []tls.Certificate{tlsCert}},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start the server
	fmt.Printf("%s Server is accepting connections on %s\n", C_STARTED, server.Addr)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Printf("%s Error starting server: %s\n", C_FAILED, err.Error())
	}
}
