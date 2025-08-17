package cer_subject_tools

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cert_vrf"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type TLSClient struct {
	certFile   string
	keyFile    string
	caFile     string
	serverAddr string
	conn       *tls.Conn
	tlsConfig  *tls.Config
	publicKey  *ecdsa.PublicKey
	privateKey *ecdsa.PrivateKey
	VRFManager *cert_vrf.VRFManager
}

func NewTLSClient(certFile, keyFile, caFile, serverAddr string) *TLSClient {
	return &TLSClient{
		certFile:   certFile,
		keyFile:    keyFile,
		caFile:     caFile,
		serverAddr: serverAddr,
		VRFManager: cert_vrf.NewVRFManager(),
	}
}

func (tc *TLSClient) LoadCertificates() error {
	cert, err := tls.LoadX509KeyPair(tc.certFile, tc.keyFile)
	if err != nil {
		return fmt.Errorf("load client certificates %s", err)
	}

	var ok bool
	tc.privateKey, ok = cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("load client certificates private key")
	}
	tc.publicKey = &tc.privateKey.PublicKey

	caCert, err := os.ReadFile(tc.caFile)
	if err != nil {
		return fmt.Errorf("load CA certificates %s", err)
	}

	caCetPool := x509.NewCertPool()
	if !caCetPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("load CA certificates failed")
	}

	tc.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCetPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		InsecureSkipVerify: false,
	}
	return nil
}

func (tc *TLSClient) Connect() error {
	log.Printf("Connecting to %s", tc.serverAddr)

	conn, err := tls.Dial("tcp", tc.serverAddr, tc.tlsConfig)
	if err != nil {
		return fmt.Errorf("connect to server %s failed", err)
	}

	tc.conn = conn

	state := conn.ConnectionState()
	log.Printf("Connected to server %s success", tc.serverAddr)
	log.Printf("TLS Version %s", state.Version)

	if len(state.PeerCertificates) > 0 {
		serverCert := state.PeerCertificates[0]
		log.Printf("Server Certificate: %s", serverCert.Subject)
		log.Printf("Server Certificate Issuer: %s", serverCert.Issuer)
	}

	tc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 4096)
	n, err := tc.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("read from server %s failed", tc.serverAddr)
	}

	welcomeMsg := string(buffer[:n])
	log.Printf("Welcome message: %s", welcomeMsg)

	return nil
}

func (tc *TLSClient) RequestVRFChallenge(sessionID string) (*cert_vrf.Challenge, error) {
	if tc.conn == nil {
		return nil, fmt.Errorf("no connection")
	}

	requestMsm := &cert_vrf.VRFMessage{
		Type:      "challenge_request",
		SessionID: sessionID,
	}

	requestJSON, err := json.Marshal(requestMsm)
	if err != nil {
		return nil, fmt.Errorf("marshal request msm %s", err)
	}

	_, err = tc.conn.Write(requestJSON)
	if err != nil {
		return nil, fmt.Errorf("write request msm %s", err)
	}

	tc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 4096)
	n, err := tc.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("read response msm %s", err)
	}

	var responseMsg cert_vrf.VRFMessage
	responseMsg.Success = true
	err = json.Unmarshal(buffer[:n], &responseMsg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response msm %s", err)
	}

	if !responseMsg.Success {
		return nil, fmt.Errorf("response msm %s failed", responseMsg.SessionID)
	}

	log.Printf("Response msm %s", responseMsg.Message)
	return responseMsg.Challenge, nil
}

func (tc *TLSClient) SubmitVRFProof(sessionID string, challenge *cert_vrf.Challenge) (bool, error) {
	if tc.conn == nil {
		return false, fmt.Errorf("no connection")
	}

	vrfKeyPair := &cert_vrf.VRFKeyPair{
		PublicKey:  tc.publicKey,
		PrivateKey: tc.privateKey,
	}

	proof, err := tc.VRFManager.GenerateVRFProof(vrfKeyPair, challenge)
	if err != nil {
		return false, fmt.Errorf("generate VRFProof %s", err)
	}

	log.Printf("Generated VRFProof %s", proof)

	proofMsg := &cert_vrf.VRFMessage{
		Type:      "proof_submission",
		SessionID: sessionID,
		Proof:     proof,
	}

	proofJSON, err := json.Marshal(proofMsg)
	if err != nil {
		return false, fmt.Errorf("marshal proof msm %s", err)
	}

	_, err = tc.conn.Write(proofJSON)
	if err != nil {
		return false, fmt.Errorf("write proof msm %s", err)
	}

	tc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 4096)
	n, err := tc.conn.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("read verification result msm %s", err)
	}

	var responseMsg cert_vrf.VRFMessage
	err = json.Unmarshal(buffer[:n], &responseMsg)
	if err != nil {
		return false, fmt.Errorf("unmarshal verification result msm %s", err)
	}

	log.Printf("VRF verification result: %s", responseMsg.Message)
	return responseMsg.Success, nil
}

func (tc *TLSClient) PerformVRFAuthentication(sessionID string) error {
	challenge, err := tc.RequestVRFChallenge(sessionID)
	if err != nil {
		return fmt.Errorf("perform VRF challenge %s", err)
	}

	verified, err := tc.SubmitVRFProof(sessionID, challenge)
	if err != nil {
		return fmt.Errorf("perform VRF challenge %s", err)
	}

	if verified {
		log.Printf("VRF challenge verified")
	} else {
		log.Printf("VRF challenge verification failed")
	}

	return nil
}

func (tc *TLSClient) StartInteractiveSession() error {
	if tc.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	log.Printf("enter interactive session!")

	go tc.readServerMessages()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			log.Printf("exiting interactive session!")
			break
		}

		_, err := tc.conn.Write([]byte(input + "\n"))
		if err != nil {
			log.Printf("send data to server %s failed", err)
			break
		}
	}
	return nil
}

func (tc *TLSClient) readServerMessages() {
	buffer := make([]byte, 4096)
	for {
		tc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := tc.conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("Server is close")
			} else {
				log.Printf("readServerMessages failed %s", err)
			}
			break
		}

		message := strings.TrimSpace(string(buffer[:n]))
		if message != "" {
			fmt.Printf("\n Server: %s\n", message)
		}
	}
}

func (tc *TLSClient) VerifyServerCertificate() error {
	if tc.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	state := tc.conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("server certificate is empty")
	}

	serverCert := state.PeerCertificates[0]

	now := time.Now()
	if now.Before(serverCert.NotBefore) {
		return fmt.Errorf("server certificate is not valid yet")
	}
	if now.After(serverCert.NotAfter) {
		return fmt.Errorf("server certificate is expired")
	}

	return nil
}

func (tc *TLSClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}
