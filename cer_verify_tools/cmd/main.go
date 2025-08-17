package main

import (
	ca_verifier_tools "github.com/FISCO-BCOS/go-sdk/cer_verify_tools"
	"log"
	"os"
)

func main() {
	currentDir, _ := os.Getwd()
	certFile := currentDir + "/certs/tls_server.crt"
	keyFile := currentDir + "/certs/tls_server.key"
	caFile := currentDir + "/certs/tls_ca.crt"
	port := "8443"

	verifier := ca_verifier_tools.NewVerifierManager(certFile, keyFile, caFile, port)

	err := verifier.LoadCertificates()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting server...")
	err = verifier.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}
