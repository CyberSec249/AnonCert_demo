package cert_vrf

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

type VRFKeyPair struct {
	PublicKey  *ecdsa.PublicKey  `json:"public_key"`
	PrivateKey *ecdsa.PrivateKey `json:"private_key"`
}

type VRFProof struct {
	Gamma *ECPoint `json:"gamma"`
	C     *big.Int `json:"c"`
	S     *big.Int `json:"s"`
	Beta  []byte   `json:"beta"`
}

type ECPoint struct {
	X *big.Int `json:"x"`
	Y *big.Int `json:"y"`
}

type Challenge struct {
	RandomSeed string    `json:"random_seed"`
	SessionID  string    `json:"session_id"`
	Timestamp  time.Time `json:"timestamp"`
	FinalHash  []byte    `json:"final_hash"`
}

type VRFMessage struct {
	Type      string     `json:"type"`
	SessionID string     `json:"session_id"`
	Challenge *Challenge `json:"challenge,omitempty"`
	Proof     *VRFProof  `json:"proof,omitempty"`
	Success   bool       `json:"success,omitempty"`
	Message   string     `json:"message,omitempty"`
}

type VRFManager struct {
	curve elliptic.Curve
}

func NewVRFManager() *VRFManager {
	return &VRFManager{
		curve: elliptic.P384(),
	}
}

func (vm *VRFManager) GenerateVRFKeyPair() (*VRFKeyPair, error) {
	priv, err := ecdsa.GenerateKey(vm.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VRF key pair: %w", err)
	}
	return &VRFKeyPair{
		PrivateKey: priv,
		PublicKey:  &priv.PublicKey,
	}, nil
}

func (vm *VRFManager) GenerateVRFChallenge(SessionID string) (*Challenge, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomSeed := hex.EncodeToString(randomBytes)
	timestamp := time.Now()

	hasher := sha256.New()
	hasher.Write([]byte(randomSeed))
	hasher.Write([]byte(timestamp.Format(time.RFC3339Nano)))
	hasher.Write([]byte(SessionID))
	finalHash := hasher.Sum(nil)
	return &Challenge{
		RandomSeed: randomSeed,
		SessionID:  SessionID,
		Timestamp:  timestamp,
		FinalHash:  finalHash,
	}, nil
}

func (vm *VRFManager) GenerateVRFProof(vrfKeyPair *VRFKeyPair, challenge *Challenge) (*VRFProof, error) {
	if vrfKeyPair.PrivateKey == nil {
		return nil, fmt.Errorf("VRF private pair is null")
	}
	alpha := challenge.FinalHash

	h := vm.hashToCurve(alpha)
	if h == nil || h.X == nil || h.Y == nil {
		return nil, fmt.Errorf("failed to hash to curve")
	}

	gammaX, gammaY := vm.curve.ScalarMult(h.X, h.Y, vrfKeyPair.PrivateKey.D.Bytes())
	gamma := &ECPoint{X: gammaX, Y: gammaY}

	beta := vm.hashPoint(gamma)

	proof, err := vm.generateZKProof(vrfKeyPair.PrivateKey, h, gamma, alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	return &VRFProof{
		Gamma: gamma,
		C:     proof.C,
		S:     proof.S,
		Beta:  beta,
	}, nil
}

func (vm *VRFManager) hashToCurve(data []byte) *ECPoint {
	hasher := sha256.New()
	hasher.Write(data)

	for i := 0; i < 256; i++ {
		hasher.Write([]byte{byte(i)})
		hash := hasher.Sum(nil)

		x := new(big.Int).SetBytes(hash[:32])
		if x.Cmp(vm.curve.Params().P) >= 0 {
			hasher.Reset()
			hasher.Write(data)
			continue
		}

		x3 := new(big.Int).Mul(x, x)
		x3.Mul(x3, x)
		threeX := new(big.Int).Mul(big.NewInt(3), x)
		threeX.Mod(threeX, vm.curve.Params().P)

		y2 := new(big.Int).Sub(x3, threeX)
		y2.Add(y2, vm.curve.Params().B)
		y2.Add(y2, vm.curve.Params().P)

		y := vm.modSqrt(y2, vm.curve.Params().P)
		if y != nil && vm.curve.IsOnCurve(x, y) {
			return &ECPoint{X: x, Y: y}
		}

		hasher.Reset()
		hasher.Write(data)
	}
	gx, gy := vm.curve.Params().Gx, vm.curve.Params().Gy
	if gx == nil || gy == nil {
		gx = big.NewInt(1)
		gy = big.NewInt(1)
		if !vm.curve.IsOnCurve(gx, gy) {
			gx, gy = vm.curve.ScalarBaseMult([]byte{1})
		}
	}
	return &ECPoint{X: new(big.Int).Set(gx), Y: new(big.Int).Set(gy)}
}

func (vm *VRFManager) hashPoint(point *ECPoint) []byte {
	hasher := sha256.New()
	hasher.Write(point.X.Bytes())
	hasher.Write(point.Y.Bytes())
	return hasher.Sum(nil)
}

func (vm *VRFManager) modSqrt(a, p *big.Int) *big.Int {
	if a == nil || p == nil {
		return nil
	}
	if a.Sign() == 0 {
		return big.NewInt(0)
	}
	mod4 := new(big.Int).Mod(p, big.NewInt(4))
	if mod4.Cmp(big.NewInt(3)) == 0 {
		exp := new(big.Int).Add(p, big.NewInt(1))
		exp.Div(exp, big.NewInt(4))
		result := new(big.Int).Exp(a, exp, p)

		test := new(big.Int).Mul(result, result)
		test.Mod(test, p)
		if test.Cmp(a) == 0 {
			return result
		}
		neg := new(big.Int).Sub(p, result)
		test2 := new(big.Int).Mul(neg, neg)
		test.Mod(test2, p)
		if test2.Cmp(a) == 0 {
			return neg
		}
	}
	return vm.tonelliShanks(a, p)
}

func (vm *VRFManager) tonelliShanks(a, p *big.Int) *big.Int {
	if a == nil || p == nil {
		return nil
	}

	exp := new(big.Int).Sub(p, big.NewInt(1))
	exp.Div(exp, big.NewInt(2))
	legendre := new(big.Int).Exp(a, exp, p)

	if legendre.Cmp(big.NewInt(1)) != 0 {
		return nil
	}

	mod4 := new(big.Int).Mod(p, big.NewInt(4))
	if mod4.Cmp(big.NewInt(3)) == 0 {
		exp := new(big.Int).Add(p, big.NewInt(1))
		exp.Div(exp, big.NewInt(4))
		return new(big.Int).Exp(a, exp, p)
	}

	Q := new(big.Int).Sub(p, big.NewInt(1))
	S := 0
	for Q.Bit(0) == 0 {
		Q.Rsh(Q, 1)
		S++
	}

	z := big.NewInt(2)
	for {
		zLegendre := new(big.Int).Exp(z, exp, p)
		if zLegendre.Cmp(new(big.Int).Sub(p, big.NewInt(1))) == 0 {
			break
		}
		z.Add(z, big.NewInt(1))
	}

	M := big.NewInt(int64(S))
	c := new(big.Int).Exp(z, Q, p)
	t := new(big.Int).Exp(a, Q, p)
	R := new(big.Int).Exp(a, new(big.Int).Add(Q, big.NewInt(1)).Rsh(new(big.Int).Add(Q, big.NewInt(1)), 1), p)

	for {
		if t.Cmp(big.NewInt(1)) == 0 {
			return R
		}
		i := 1
		temp := new(big.Int).Mul(t, t)
		temp.Mod(temp, p)
		for temp.Cmp(big.NewInt(1)) != 0 && i < int(M.Int64()) {
			temp.Mul(temp, temp)
			temp.Mod(temp, p)
			i++
		}

		b := new(big.Int).Set(c)
		for j := 0; j < int(M.Int64())-i-1; j++ {
			b.Mul(b, b)
			b.Mod(b, p)
		}

		M = big.NewInt(int64(i))
		c.Mul(b, b)
		c.Mod(c, p)
		t.Mul(t, c)
		t.Mod(t, p)
		R.Mul(R, b)
		R.Mod(R, p)
	}
}

func (vm *VRFManager) generateZKProof(privateKey *ecdsa.PrivateKey, h *ECPoint, gamma *ECPoint, alpha []byte) (*struct{ C, S *big.Int }, error) {
	k, err := rand.Int(rand.Reader, vm.curve.Params().N)
	if err != nil {
		return nil, fmt.Errorf("failed to generate zk proof: %w", err)
	}

	r1x, r1y := vm.curve.ScalarBaseMult(k.Bytes())
	r2x, r2y := vm.curve.ScalarMult(h.X, h.Y, k.Bytes())

	hasher := sha256.New()
	hasher.Write(vm.curve.Params().Gx.Bytes())
	hasher.Write(vm.curve.Params().Gy.Bytes())
	hasher.Write(h.X.Bytes())
	hasher.Write(h.Y.Bytes())
	hasher.Write(privateKey.PublicKey.X.Bytes())
	hasher.Write(privateKey.PublicKey.Y.Bytes())
	hasher.Write(gamma.X.Bytes())
	hasher.Write(gamma.Y.Bytes())
	hasher.Write(r1x.Bytes())
	hasher.Write(r1y.Bytes())
	hasher.Write(r2x.Bytes())
	hasher.Write(r2y.Bytes())
	hasher.Write(alpha)

	c := new(big.Int).SetBytes(hasher.Sum(nil))
	c.Mod(c, vm.curve.Params().N)

	s := new(big.Int).Mul(c, privateKey.D)
	s.Add(s, k)
	s.Mod(s, vm.curve.Params().N)

	return &struct{ C, S *big.Int }{C: c, S: s}, nil
}

func (vm *VRFManager) VerifyZKProof(vrfPK *ecdsa.PublicKey, h *ECPoint, gamma *ECPoint, alpha []byte, c, s *big.Int) bool {
	sx, sy := vm.curve.ScalarBaseMult(s.Bytes())
	cx, cy := vm.curve.ScalarMult(vrfPK.X, vrfPK.Y, c.Bytes())

	cy.Sub(vm.curve.Params().P, cy)
	r1x, r1y := vm.curve.Add(sx, sy, cx, cy)

	sx2, sy2 := vm.curve.ScalarMult(h.X, h.Y, s.Bytes())
	cx2, cy2 := vm.curve.ScalarMult(gamma.X, gamma.Y, c.Bytes())

	cy2.Sub(vm.curve.Params().P, cy2)
	r2x, r2y := vm.curve.Add(sx2, sy2, cx2, cy2)

	hasher := sha256.New()
	hasher.Write(vm.curve.Params().Gx.Bytes())
	hasher.Write(vm.curve.Params().Gy.Bytes())
	hasher.Write(h.X.Bytes())
	hasher.Write(h.Y.Bytes())
	hasher.Write(vrfPK.X.Bytes())
	hasher.Write(vrfPK.Y.Bytes())
	hasher.Write(gamma.X.Bytes())
	hasher.Write(gamma.Y.Bytes())
	hasher.Write(r1x.Bytes())
	hasher.Write(r1y.Bytes())
	hasher.Write(r2x.Bytes())
	hasher.Write(r2y.Bytes())
	hasher.Write(alpha)

	expectedC := new(big.Int).SetBytes(hasher.Sum(nil))
	expectedC.Mod(expectedC, vm.curve.Params().N)

	return c.Cmp(expectedC) == 0
}

func (vm *VRFManager) VerifyVRFProof(vrfPK *ecdsa.PublicKey, challenge *Challenge, proof *VRFProof) (bool, error) {
	if vrfPK == nil || challenge == nil || proof == nil {
		return false, fmt.Errorf("invalid VRF verification parameters")
	}
	alpha := challenge.FinalHash

	h := vm.hashToCurve(alpha)

	isValid := vm.VerifyZKProof(vrfPK, h, proof.Gamma, alpha, proof.C, proof.S)
	if !isValid {
		return false, fmt.Errorf("invalid VRF proof")
	}

	expectedBeta := vm.hashPoint(proof.Gamma)
	if !vm.bytesEqual(expectedBeta, proof.Beta) {
		return false, fmt.Errorf("invalid VRF proof")
	}
	return true, nil
}

func (vm *VRFManager) DeserializeVRFPK(data []byte) (*ecdsa.PublicKey, error) {
	pkInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	pk, ok := pkInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key")
	}
	return pk, nil
}

func (vm *VRFManager) bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
