package d0

import "fmt"

// BlindIDKeygen is the location of the d0_blind_id keygen executable
const BlindIDKeygen = "./crypto-keygen-standalone"

// BlindIDD0pk is the location of the d0_blind_id public key
const BlindIDD0pk = "key_0.d0pk"

// VerifyResult is the result of a d0_blind_id verification
type VerifyResult struct {
	IDFP     string
	CAStatus bool
}

// Verify checks if the given request data is verified via the d0_blind_id library
// via its command line executable.
func Verify(signature, queryString, data string) (*VerifyResult, error) {
	return nil, fmt.Errorf("unable to verify the request data")
}
