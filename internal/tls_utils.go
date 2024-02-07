package util

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
	C_CERT_BASEPATH string = "/dev/shm/"
)

func generateCertificate() error {
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
	defer keyFile.Close()

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
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return err
	}

	return nil
}

func ServeTLS(r *chi.Mux) {

	err := generateCertificate()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Certificate and private key generated successfully.")

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
		/* trunk-ignore(golangci-lint/errcheck) */
		w.Write([]byte("nodelocker"))
	})

	// Create a server with TLS configuration
	server := &http.Server{
		Addr:         "0.0.0.0:8443",
		Handler:      r,
		TLSConfig:    &tls.Config{Certificates: []tls.Certificate{tlsCert}},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start the server
	fmt.Printf("Server is running on %s\n", server.Addr)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error starting server:", err.Error())
	}
}
