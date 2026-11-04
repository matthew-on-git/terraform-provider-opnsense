// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
)

func validateTrustPEM(p map[string]string) error {
	cert, key := p["crt_payload"], p["prv_payload"]
	if cert != "" {
		block, rest := pem.Decode([]byte(cert))
		if block == nil || block.Type != "CERTIFICATE" || strings.TrimSpace(string(rest)) != "" {
			return errors.New("trust certificate must contain one PEM certificate")
		}
		if _, err := x509.ParseCertificate(block.Bytes); err != nil {
			return errors.New("invalid Trust certificate")
		}
	}
	if key != "" {
		block, rest := pem.Decode([]byte(key))
		if block == nil || strings.TrimSpace(string(rest)) != "" {
			return errors.New("trust private key must contain one unencrypted PEM key")
		}
		var parsed any
		var err error
		switch block.Type {
		case "PRIVATE KEY":
			parsed, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		case "RSA PRIVATE KEY":
			parsed, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		case "EC PRIVATE KEY":
			parsed, err = x509.ParseECPrivateKey(block.Bytes)
		default:
			return errors.New("unsupported Trust private key PEM type")
		}
		if err != nil {
			return errors.New("invalid Trust private key")
		}
		if cert != "" {
			if _, err := tls.X509KeyPair([]byte(cert), []byte(key)); err != nil {
				return errors.New("trust private key does not match certificate")
			}
		}
		if _, ok := parsed.(crypto.Signer); !ok {
			return errors.New("unsupported Trust signing key")
		}
	}
	if csr := p["csr_payload"]; csr != "" {
		block, rest := pem.Decode([]byte(csr))
		if block == nil || len(bytes.TrimSpace(rest)) != 0 {
			return errors.New("invalid Trust CSR")
		}
		req, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil || req.CheckSignature() != nil {
			return errors.New("invalid Trust CSR signature")
		}
	}
	return nil
}

// Existing appliance-held material can satisfy a deliberate transition. Never
// consult the retained add-only local key: lookup cannot recover that material.
func hydrateTrustTransition(p, current map[string]string) error {
	switch p["action"] {
	case "reissue":
		if p["prv_payload"] == "" {
			p["prv_payload"] = current["prv_payload"]
		}
		for _, field := range []string{"caref", "commonname", "country", "state", "city", "organization", "organizationalunit", "email", "key_type", "digest", "cert_type", "ocsp_uri", "altnames_dns", "altnames_ip", "altnames_uri", "altnames_email"} {
			if _, configured := p[field]; !configured && current[field] != "" {
				p[field] = current[field]
			}
		}
	case "sign_csr":
		if p["csr_payload"] == "" {
			p["csr_payload"] = current["csr_payload"]
		}
	case "import_csr":
		if p["crt_payload"] == "" {
			return errors.New("import_csr requires a signed certificate")
		}
		if err := validateTrustPEM(map[string]string{"crt_payload": p["crt_payload"], "prv_payload": current["prv_payload"], "csr_payload": current["csr_payload"]}); err != nil {
			return err
		}
		if current["csr_payload"] != "" {
			block, _ := pem.Decode([]byte(p["crt_payload"]))
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return errors.New("invalid signed certificate")
			}
			block, _ = pem.Decode([]byte(current["csr_payload"]))
			csr, err := x509.ParseCertificateRequest(block.Bytes)
			if err != nil {
				return errors.New("invalid existing CSR")
			}
			certKey, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
			if err != nil {
				return errors.New("invalid certificate public key")
			}
			csrKey, err := x509.MarshalPKIXPublicKey(csr.PublicKey)
			if err != nil {
				return errors.New("invalid CSR public key")
			}
			if !bytes.Equal(certKey, csrKey) {
				return errors.New("signed certificate does not match existing CSR")
			}
		}
	}
	return nil
}
