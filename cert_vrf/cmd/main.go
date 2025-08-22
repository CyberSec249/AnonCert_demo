package main

import (
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cert_vrf"
	"log"
)

func main() {
	vrfManager := cert_vrf.NewVRFManager()

	vrfKeyPair, err := vrfManager.GenerateVRFKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	sessionID := "test-session-id"
	challenge, err := vrfManager.GenerateVRFChallenge(sessionID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("challenge: %s\n", challenge)

	vrfProof, err := vrfManager.GenerateVRFProof(vrfKeyPair, challenge)
	if err != nil {
		log.Fatal(err)
	}

	isValid, err := vrfManager.VerifyVRFProof(vrfKeyPair.PublicKey, challenge, vrfProof)
	if err != nil {
		log.Fatal(err)
	}
	if isValid {
		fmt.Println("VRF proof verified")
	} else {
		fmt.Println("VRF proof verification failed")
	}

}
