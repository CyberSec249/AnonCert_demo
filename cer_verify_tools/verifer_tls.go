package ca_verifier_tools

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cert_vrf"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type VRFSession struct {
	SessionID  string              `json:"session_id"`
	Challenge  *cert_vrf.Challenge `json:"challenge,omitempty"`
	ClientPK   *ecdsa.PublicKey    `json:"client_pk,omitempty"`
	IsVerified bool                `json:"is_verified"`
	CreateAT   time.Time           `json:"create_at"`
}

type VerifierManager struct {
	certFile    string
	keyFile     string
	caFile      string
	port        string
	listener    net.Listener
	tlsConfig   *tls.Config
	VRFManager  *cert_vrf.VRFManager
	vrfSessions map[string]*VRFSession
}

func NewVerifierManager(certFile, keyFile, caFile, port string) *VerifierManager {
	return &VerifierManager{
		certFile:    certFile,
		keyFile:     keyFile,
		caFile:      caFile,
		port:        port,
		VRFManager:  cert_vrf.NewVRFManager(),
		vrfSessions: make(map[string]*VRFSession),
	}
}

func (vm *VerifierManager) LoadCertificates() error {
	cert, err := tls.LoadX509KeyPair(vm.certFile, vm.keyFile)
	if err != nil {
		return fmt.Errorf("error loading server certificate: %v", err)
	}

	caCert, err := os.ReadFile(vm.caFile)
	if err != nil {
		return fmt.Errorf("error loading CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("error appending CA certificate")
	}

	vm.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}
	return nil
}

func (vm *VerifierManager) StartServer() error {
	listener, err := tls.Listen("tcp", ":"+vm.port, vm.tlsConfig)
	if err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}

	vm.listener = listener
	log.Printf("Listening on port %s", vm.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go vm.handleConnection(conn)
	}
}

func (vm *VerifierManager) handleConnection(conn net.Conn) {
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	err := tlsConn.Handshake()
	if err != nil {
		log.Printf("Error handshake connection: %v", err)
		return
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		log.Printf("No client certificate found")
		return
	}

	clientCert := state.PeerCertificates[0]
	log.Printf("%s certificate found", clientCert.Subject)

	welcomeMsg := "TLS connection successful, Server is ready. \n"
	_, err = tlsConn.Write([]byte(welcomeMsg))
	if err != nil {
		log.Printf("Error writing welcome message: %v", err)
		return
	}

	vm.handleDataTransfer(tlsConn)
}

func (vm *VerifierManager) handleDataTransfer(conn *tls.Conn) {
	buffer := make([]byte, 4096)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by remote host")
			} else {
				log.Printf("Error reading data: %v", err)
			}
			break
		}

		receiveData := string(buffer[:n])
		log.Printf("Data received: %s", receiveData)

		startTLSVerifier := time.Now()
		response := vm.processVRFMessage(receiveData, conn)
		endTLSVerifier := time.Since(startTLSVerifier)
		fmt.Println("Verifier VRF Time:", endTLSVerifier)

		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("Error writing data: %v", err)
			break
		}
	}
}

func (vm *VerifierManager) processVRFMessage(receiveData string, conn *tls.Conn) string {
	var vrfMsg cert_vrf.VRFMessage
	err := json.Unmarshal([]byte(receiveData), &vrfMsg)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		return vm.createErrorResponse("invalid JSON format")
	}

	switch vrfMsg.Type {
	case "challenge_request":
		return vm.handleChallengeRequest(vrfMsg, conn)
	case "proof_submission":
		return vm.handleProofSubmission(vrfMsg)
	case "ping":
		return vm.createSimpleResponse("pong")
	case "quit", "exit":
		return vm.createSimpleResponse("goodbye")
	default:
		return vm.createSimpleResponse(fmt.Sprintf("unknown command: %s", vrfMsg.Type))
	}
}

func (vm *VerifierManager) handleChallengeRequest(vrfMsg cert_vrf.VRFMessage, conn *tls.Conn) string {
	sessionID := vrfMsg.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	challenge, err := vm.VRFManager.GenerateVRFChallenge(sessionID)
	if err != nil {
		log.Printf("Error generating challenge: %v", err)
		return vm.createErrorResponse("Error generating challenge")
	}
	log.Printf("Created challenge: %s", challenge)

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return vm.createErrorResponse("No client certificate found")
	}

	clientCert := state.PeerCertificates[0]
	clientPK, ok := clientCert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return vm.createErrorResponse("Invalid client public key")
	}

	session := &VRFSession{
		SessionID:  sessionID,
		Challenge:  challenge,
		ClientPK:   clientPK,
		IsVerified: false,
		CreateAT:   time.Now(),
	}
	vm.vrfSessions[sessionID] = session

	log.Printf("Created session: %s", sessionID)

	responseJSON, err := json.Marshal(session)
	if err != nil {
		return vm.createErrorResponse("Error marshalling response")
	}
	return string(responseJSON)
}

func (vm *VerifierManager) handleProofSubmission(vrfMsg cert_vrf.VRFMessage) string {
	session, exists := vm.vrfSessions[vrfMsg.SessionID]
	if !exists {
		return vm.createErrorResponse("No session found")
	}

	if vrfMsg.Proof == nil {
		return vm.createErrorResponse("No proof found")
	}

	isValid, err := vm.VRFManager.VerifyVRFProof(session.ClientPK, session.Challenge, vrfMsg.Proof)
	if err != nil {
		log.Printf("Error verifying proof: %v", err)
		return vm.createErrorResponse("Error verifying proof")
	}

	session.IsVerified = true

	var message string
	if isValid {
		message = "Verified successfully"
		log.Printf("Verified successfully: %s", vrfMsg.SessionID)
	} else {
		message = "Invalid proof"
		log.Printf("Invalid proof: %s", vrfMsg.SessionID)
	}

	response := &cert_vrf.VRFMessage{
		Type:      "verification_result",
		SessionID: vrfMsg.SessionID,
		Success:   isValid,
		Message:   message,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return vm.createErrorResponse("Error marshalling response")
	}
	return string(responseJSON)
}

func (vm *VerifierManager) createSimpleResponse(message string) string {
	response := &cert_vrf.VRFMessage{
		Type:    "simple_response",
		Success: true,
		Message: message,
	}
	responseJSON, _ := json.Marshal(response)
	return string(responseJSON)
}

func (vm *VerifierManager) createErrorResponse(message string) string {
	response := &cert_vrf.VRFMessage{
		Type:    "error",
		Success: false,
		Message: message,
	}
	responseJSON, _ := json.Marshal(response)
	return string(responseJSON)
}

func (vm *VerifierManager) Close() error {
	if vm.listener != nil {
		return vm.listener.Close()
	}
	return nil
}
