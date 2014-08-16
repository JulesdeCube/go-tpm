// Copyright (c) 2014, Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tpm

import (
	"encoding/binary"
	"errors"
	"os"
	"strconv"

	"github.com/golang/glog"
)

// submitTPMRequest sends a structure to the TPM device file and gets results
// back, interpreting them as a new provided structure.
func submitTPMRequest(f *os.File, tag uint16, ord uint32, in []interface{}, out []interface{}) (uint32, error) {
	ch := commandHeader{tag, 0, ord}
	inb, err := packWithHeader(ch, in)
	if err != nil {
		return 0, err
	}

	if glog.V(2) {
		glog.Infof("TPM request:\n%x\n", inb)
	}
	if _, err := f.Write(inb); err != nil {
		return 0, err
	}

	// Try to read the whole thing, but handle the case where it's just a
	// ResponseHeader and not the body, since that's what happens in the error
	// case.
	var rh responseHeader
	rhSize := binary.Size(rh)
	outb := make([]byte, maxTPMResponse)
	outlen, err := f.Read(outb)
	if err != nil {
		return 0, err
	}

	// Resize the buffer to match the amount read from the TPM.
	outb = outb[:outlen]
	if glog.V(2) {
		glog.Infof("TPM response:\n%x\n", outb)
	}

	if err := unpack(outb[:rhSize], []interface{}{&rh}); err != nil {
		return 0, err
	}

	// Check success before trying to read the rest of the result.
	// Note that the command tag and its associated response tag differ by 3,
	// e.g., tagRQUCommand == 0x00C1, and tagRSPCommand == 0x00C4.
	if rh.Res != 0 {
		return rh.Res, tpmError(rh.Res)
	}

	if rh.Tag != ch.Tag+3 {
		return 0, errors.New("inconsistent tag returned by TPM. Expected " + strconv.Itoa(int(ch.Tag+3)) + " but got " + strconv.Itoa(int(rh.Tag)))
	}

	if rh.Size > uint32(rhSize) {
		if err := unpack(outb[rhSize:], out); err != nil {
			return 0, err
		}
	}

	return rh.Res, nil
}

// oiap sends an OIAP command to the TPM and gets back an auth value and a
// nonce.
func oiap(f *os.File) (*oiapResponse, error) {
	var resp oiapResponse
	out := []interface{}{&resp}
	// In this case, we don't need to check ret, since all the information is
	// contained in err.
	if _, err := submitTPMRequest(f, tagRQUCommand, ordOIAP, nil, out); err != nil {
		return nil, err
	}

	return &resp, nil
}

// osap sends an OSAPCommand to the TPM and gets back authentication
// information in an OSAPResponse.
func osap(f *os.File, osap *osapCommand) (*osapResponse, error) {
	in := []interface{}{osap}
	var resp osapResponse
	out := []interface{}{&resp}
	// In this case, we don't need to check the ret value, since all the
	// information is contained in err.
	if _, err := submitTPMRequest(f, tagRQUCommand, ordOSAP, in, out); err != nil {
		return nil, err
	}

	return &resp, nil
}

// seal performs a seal operation on the TPM.
func seal(f *os.File, sc *sealCommand, pcrs *pcrInfoLong, data []byte, ca *commandAuth) (*tpmStoredData, *responseAuth, uint32, error) {
	pcrsize := binary.Size(pcrs)
	if pcrsize < 0 {
		return nil, nil, 0, errors.New("couldn't compute the size of a pcrInfoLong")
	}

	// TODO(tmroeder): special-case pcrInfoLong in pack/unpack so we don't have
	// to write out the length explicitly here.
	in := []interface{}{sc, uint32(pcrsize), pcrs, data, ca}

	var tsd tpmStoredData
	var ra responseAuth
	out := []interface{}{&tsd, &ra}
	ret, err := submitTPMRequest(f, tagRQUAuth1Command, ordSeal, in, out)
	if err != nil {
		return nil, nil, 0, err
	}

	return &tsd, &ra, ret, nil
}

// unseal data sealed by the TPM.
func unseal(f *os.File, keyHandle Handle, sealed *tpmStoredData, ca1 *commandAuth, ca2 *commandAuth) ([]byte, *responseAuth, *responseAuth, uint32, error) {
	in := []interface{}{keyHandle, sealed, ca1, ca2}
	var outb []byte
	var ra1 responseAuth
	var ra2 responseAuth
	out := []interface{}{&outb, &ra1, &ra2}
	ret, err := submitTPMRequest(f, tagRQUAuth2Command, ordUnseal, in, out)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	return outb, &ra1, &ra2, ret, nil
}

// flushSpecific removes a handle from the TPM. Note that removing a handle
// doesn't require any authentication.
func flushSpecific(f *os.File, handle Handle, resourceType uint32) error {
	// In this case, all the information is in err, so we don't check the
	// specific return-value details.
	_, err := submitTPMRequest(f, tagRQUCommand, ordFlushSpecific, []interface{}{handle, resourceType}, nil)
	return err
}

// loadKey2 loads a key into the TPM. It's a tagRQUAuth1Command, so it only
// needs one auth parameter.
// TODO(tmroeder): support key12, too.
func loadKey2(f *os.File, k *key, ca *commandAuth) (Handle, *responseAuth, uint32, error) {
	// We always load our keys with the SRK as the parent key.
	in := []interface{}{khSRK, k, ca}
	var keyHandle Handle
	var ra responseAuth
	out := []interface{}{&keyHandle, &ra}
	if glog.V(2) {
		glog.Info("About to submit the TPM request for loadKey2")
	}

	ret, err := submitTPMRequest(f, tagRQUAuth1Command, ordLoadKey2, in, out)
	if err != nil {
		return 0, nil, 0, err
	}

	if glog.V(2) {
		glog.Info("Received a good response for loadKey2")
	}

	return keyHandle, &ra, ret, nil
}

// getPubKey gets a public key from the TPM
func getPubKey(f *os.File, keyHandle Handle, ca *commandAuth) (*pubKey, *responseAuth, uint32, error) {
	in := []interface{}{keyHandle, ca}
	var pk pubKey
	var ra responseAuth
	out := []interface{}{&pk, &ra}
	ret, err := submitTPMRequest(f, tagRQUAuth1Command, ordGetPubKey, in, out)
	if err != nil {
		return nil, nil, 0, err
	}

	return &pk, &ra, ret, nil
}

// quote2 signs arbitrary data under a given set of PCRs and using a key
// specified by keyHandle. It returns information about the PCRs it signed
// under, the signature, auth information, and optionally information about the
// TPM itself. Note that the input to quote2 must be exactly 20 bytes, so it is
// normally the SHA1 hash of the data.
func quote2(f *os.File, keyHandle Handle, hash [20]byte, pcrs *pcrSelection, addVersion byte, ca *commandAuth) (*pcrInfoShort, *capVersionInfo, []byte, []byte, *responseAuth, uint32, error) {
	in := []interface{}{keyHandle, hash, pcrs, addVersion, ca}
	var pcrShort pcrInfoShort
	var capInfo capVersionInfo
	var capBytes []byte
	var sig []byte
	var ra responseAuth
	out := []interface{}{&pcrShort, &capBytes, &sig, &ra}
	ret, err := submitTPMRequest(f, tagRQUAuth1Command, ordQuote2, in, out)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}

	// Deserialize the capInfo, if any.
	if len(capBytes) == 0 {
		return &pcrShort, nil, capBytes, sig, &ra, ret, nil
	}

	size := binary.Size(capInfo.CapVersionFixed)
	capInfo.VendorSpecific = make([]byte, len(capBytes)-size)
	if err := unpack(capBytes[:size], []interface{}{&capInfo.CapVersionFixed}); err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}

	copy(capInfo.VendorSpecific, capBytes[size:])

	return &pcrShort, &capInfo, capBytes, sig, &ra, ret, nil
}