// package tpmt contains TPM 2.0 structures prefixed with "TPMT_"
package tpmt

import "github.com/google/go-tpm/direct/structures/internal"

// HA represents a TPMT_HA.
// See definition in Part 2: Structures, section 10.3.2.
type HA = internal.TPMTHA

// TKCreation represents a TPMT_TK_CREATION.
// See definition in Part 2: Structures, section 10.7.3.
type TKCreation = internal.TPMTTKCreation

// TVerified represents a TPMT_TK_Verified.
// See definition in Part 2: Structures, section 10.7.4.
type TKVerified = internal.TPMTTKVerified

// TKAuth represents a TPMT_TK_AUTH.
// See definition in Part 2: Structures, section 10.7.5.
type TKAuth = internal.TPMTTKAuth

// TPMTTKHashCheck represents a TPMT_TK_HASHCHECK.
// See definition in Part 2: Structures, section 10.7.6.
type TKHashCheck = internal.TPMTTKHashCheck

// SymDef represents a TPMT_SYM_DEF.
// See definition in Part 2: Structures, section 11.1.6.
type SymDef = internal.TPMTSymDef

// SymDefObject represents a TPMT_SYM_DEF_OBJECT.
// See definition in Part 2: Structures, section 11.1.7.
type SymDefObject = internal.TPMTSymDefObject

// KeyedHashScheme represents a TPMT_KEYEDHASH_SCHEME.
// See definition in Part 2: Structures, section 11.1.23.
type KeyedHashScheme = internal.TPMTKeyedHashScheme

// SigScheme represents a TPMT_SIG_SCHEME.
// See definition in Part 2: Structures, section 11.2.1.5.
type SigScheme = internal.TPMTSigScheme

// KDFScheme represents a TPMT_KDF_SCHEME.
// See definition in Part 2: Structures, section 11.2.3.3.
type KDFScheme = internal.TPMTKDFScheme

// RSAScheme represents a TPMT_RSA_SCHEME.
// See definition in Part 2: Structures, section 11.2.4.2.
type RSAScheme = internal.TPMTRSAScheme

// ECCScheme represents a TPMT_ECC_SCHEME.
// See definition in Part 2: Structures, section 11.2.5.6.
type ECCScheme = internal.TPMTECCScheme

// Signature represents a TPMT_SIGNATURE.
// See definition in Part 2: Structures, section 11.3.4.
type Signature = internal.TPMTSignature

// Public represents a TPMT_PUBLIC.
// See definition in Part 2: Structures, section 12.2.4.
type Public = internal.TPMTPublic

// Template represents a TPMT_TEMPLATE. It is not defined in the spec.
// It represents the alternate form of TPMT_PUBLIC for TPM2B_TEMPLATE as
// described in Part 2: Structures, 12.2.6.
type Template = internal.TPMTTemplate

// Sensitive represents a TPMT_SENSITIVE.
// See definition in Part 2: Structures, section 12.3.2.4.
type Sensitive = internal.TPMTSensitive