package d0

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// D0BlindIDKeyGen is the default location standalone verification binary.
const D0BlindIDKeyGen = "/usr/local/bin/crypto-keygen-standalone"

// D0BlindIDPubKey is the default location of the d0 public key.
const D0BlindIDPubKey = "~/key_0.d0pk"

// VerifyResult is the result of a d0_blind_id verification
type VerifyResult struct {
	IDFP     string
	CAStatus bool
}

// Verify checks if the given request data is verified via the d0_blind_id library
// via its command line executable.
func Verify(keygen, pubkey, signature, queryString, data string) (*VerifyResult, error) {
	if signature == "" {
		return nil, fmt.Errorf("missing signature")
	}

	var input string
	if data == "" {
		input = queryString
	} else {
		input = fmt.Sprintf("%s\x00%s", data, queryString)
	}

	// Create the data file for processing.
	dFile, err := ioutil.TempFile("", "d0verify_d.*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temporary data file for d0 verification")
	}

	dFile.Write([]byte(input))
	dFile.Close()
	defer os.Remove(dFile.Name())

	// Create the signature file for processing.
	sFile, err := ioutil.TempFile("", "d0verify_s.*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temporary signature file for d0 verification")
	}

	sData64, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, err
	}

	sFile.Write([]byte(sData64))
	sFile.Close()
	defer os.Remove(sFile.Name())

	cmd := exec.Command(keygen, "-p", pubkey, "-d", dFile.Name(), "-s", sFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(output), "\n")
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected output format from %s", keygen)
	}

	caStatus := parts[0] == "1"
	result := VerifyResult{IDFP: parts[1], CAStatus: caStatus}

	return &result, nil
}
