/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package certmgr

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
)

func GenCACertSHA256(k8sca []byte) (string, error) {
	block, _ := pem.Decode(k8sca)
	var cert *x509.Certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Errorf("Failed to parse cert %v", err)
		return "", err
	}
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	x509.MarshalPKCS1PublicKey(rsaPublicKey)
	sha256bytes := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(sha256bytes[:]), nil
}

func GenCACertSHA512(k8sca []byte) (string, error) {
	block, _ := pem.Decode(k8sca)
	var cert *x509.Certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Errorf("Failed to parse cert %v", err)
		return "", err
	}
	ecdsaPublicKey := cert.PublicKey.(*ecdsa.PublicKey)
	_, err = x509.MarshalPKIXPublicKey(ecdsaPublicKey)
	if err != nil {
		log.Errorf("Failed to convert public key %v", err)
		return "", err
	}
	sha512bytes := sha512.Sum512(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(sha512bytes[:]), nil
}
