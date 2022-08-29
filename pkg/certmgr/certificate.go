/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package certmgr

//go:generate mockgen -destination=./mock/certificate_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt ep/pkg/certmgr CertificateWrapper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	cmapi "ep/pkg/api/certmgr"
	eputils "ep/pkg/eputils"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type CertificateWrapper interface {
	GenCertAndConfig(certbundle cmapi.Certificate, hosts string) error
}

var (
	RUNTIMEPKIDIR = "cert/pki"
)

const (
	CACSR       = "config/certificate/ca-csr.json"
	WFSERVERCSR = "config/certificate/workflow/server-csr.json"
	WFCLIENTCSR = "config/certificate/workflow/client-csr.json"
	REGISTRYCSR = "config/certificate/registry/registry-csr.json"
)

type CertType int32

const (
	CACERT      CertType = 0
	SERVERCERT  CertType = 1
	CLIENTCERT  CertType = 2
	UNKNOWNCERT CertType = 99
)

type Certcsr struct {
	Cn    string         `json:"CN"`
	Hosts []string       `json:"hosts"`
	Key   Certcsrkey     `json:"key"`
	Names []Certcsrnames `json:"names"`
}

type Certcsrkey struct {
	Algo string `json:"algo"`
	Size int    `json:"size"`
}

type Certcsrnames struct {
	C  string `json:"C"`
	L  string `json:"L"`
	ST string `json:"ST"`
	O  string `json:"O"`
	OU string `json:"OU"`
}

const (
	ROOTCAKEYFILE     = "cert/pki/ca-key.pem"
	WFSERVERCERTFILE  = "cert/pki/workflow/server.pem"
	WFSERVERKEYFILE   = "cert/pki/workflow/server-key.pem"
	WFCLIENTCERTFILE  = "cert/pki/workflow/client.pem"
	WFCLIENTKEYFILE   = "cert/pki/workflow/client-key.pem"
	REGSERVERCERTFILE = "cert/pki/registry/registry.pem"
	REGSERVERKEYFILE  = "cert/pki/registry/registry-key.pem"
)

func GenerateCertBundle(cb *cmapi.Certificate, ctype CertType, hosts string) error {
	var err error
	cbundle := *cb

	// Generate a private key for cert
	priv, err := generateECDSAPrivKey(cb, ctype)
	if err != nil {
		log.Errorf("Failed to generate ECDSA private key: %v", err)
		return err
	}

	// Prepare certificate template
	template, err := prepareCertTemplate(cb, ctype, hosts)
	if err != nil {
		log.Errorf("Failed to prepare certificate template: %v", err)
		return err
	}

	// Create certificates
	var derBytes []byte
	if ctype == CACERT {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		derBytes, err = x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
		if err != nil {
			log.Errorf("Failed to create certificate: %v", err)
			return err
		}
	} else {
		caraw, cerr := ioutil.ReadFile(cbundle.Ca.Cert)
		if cerr != nil {
			log.Errorf("Failed to get raw Cert: %v", cerr)
			return cerr
		}
		capem, _ := pem.Decode(caraw)
		if capem == nil {
			log.Errorf("Failed to decode Cert")
			return eputils.GetError("errCertDecodeFail")
		}
		cacert, err := x509.ParseCertificate(capem.Bytes)
		if err != nil {
			log.Errorf("Failed to parse ca certificate: %v", err)
			return err
		}
		capriv, err := ioutil.ReadFile(cbundle.Ca.Key)
		if err != nil {
			log.Errorf("Failed to get ca key: %v", err)
			return err
		}
		caprivPem, _ := pem.Decode(capriv)
		ecdsaCApriv, err := x509.ParsePKCS8PrivateKey(caprivPem.Bytes)
		if err != nil {
			log.Errorf("Failed to parse ca private key: %v", err)
			return err
		}
		derBytes, err = x509.CreateCertificate(rand.Reader, template, cacert, &priv.PublicKey, ecdsaCApriv)
		if err != nil {
			log.Errorf("Failed to create certificate: %v", err)
			return err
		}
	}

	// Encode cert and key to file
	var certFile string
	var keyFile string
	if ctype == CACERT {
		certFile = cbundle.Ca.Cert
		keyFile = cbundle.Ca.Key
	} else if ctype == SERVERCERT {
		certFile = cbundle.Server.Cert
		keyFile = cbundle.Server.Key
	} else if ctype == CLIENTCERT {
		certFile = cbundle.Client.Cert
		keyFile = cbundle.Client.Key
	}
	if eputils.FileExists(certFile) {
		valid := eputils.IsValidFile(certFile)
		if !valid {
			return eputils.GetError("errInvalidFile")
		}
	}

	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Errorf("Failed to open %v for writing: %v", certFile, err)
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Errorf("Failed to write data to %v: %v", certFile, err)
		return err
	}
	if err := certOut.Close(); err != nil {
		log.Errorf("Error closing %v: %v", certFile, err)
		return err
	}

	if eputils.FileExists(keyFile) {
		valid := eputils.IsValidFile(keyFile)
		if !valid {
			return eputils.GetError("errInvalidFile")
		}
	}
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Errorf("Failed to open %v for writing: %v", keyFile, err)
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Errorf("Unable to marshal private key: %v", err)
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Errorf("Failed to write data to %v: %v", keyFile, err)
		return err
	}
	if err := keyOut.Close(); err != nil {
		log.Errorf("Error closing %v: %v", keyFile, err)
		return err
	}
	return nil
}

func validateCertbundle(certbundle cmapi.Certificate) error {
	if certbundle.Ca != nil {
		if directoryFlag := eputils.IsDirectory(certbundle.Ca.Cert); directoryFlag {
			return eputils.GetError("errCaCertIsDir")
		}

		if directoryFlag := eputils.IsDirectory(certbundle.Ca.Key); directoryFlag {
			return eputils.GetError("errCaKeyIsDir")
		}
	}

	if certbundle.Server != nil {
		if directoryFlag := eputils.IsDirectory(certbundle.Server.Cert); directoryFlag {
			return eputils.GetError("errServerCertIsDir")
		}

		if directoryFlag := eputils.IsDirectory(certbundle.Server.Key); directoryFlag {
			return eputils.GetError("errServerKeyIsDir")
		}
	}

	if certbundle.Client != nil {
		if directoryFlag := eputils.IsDirectory(certbundle.Client.Cert); directoryFlag {
			return eputils.GetError("errClientCertIsDir")
		}

		if directoryFlag := eputils.IsDirectory(certbundle.Client.Key); directoryFlag {
			return eputils.GetError("errClientKeyIsDir")
		}
	}

	return nil
}

func GenCertAndConfig(certbundle cmapi.Certificate, hosts string) error {
	var err error
	if err := validateCertbundle(certbundle); err != nil {
		return err
	}

	targetPkiDir := RUNTIMEPKIDIR + "/" + certbundle.Name
	if _, err := os.Stat(targetPkiDir); os.IsNotExist(err) {
		if err = os.MkdirAll(targetPkiDir, os.ModePerm); err != nil {
			log.Errorf("Failed to create dir: %v", err)
			return err
		}
	}
	// Generate self-signed root CA, server cert/key pair and
	// client cert/key pair
	if certbundle.Ca.Cert == "" || certbundle.Ca.Key == "" {
		log.Errorf("Cert path or Key path is nil")
		return eputils.GetError("errCertNil")
	}
	_, cacerterr := os.Stat(certbundle.Ca.Cert)
	if cacerterr != nil && !os.IsNotExist(cacerterr) {
		log.Errorf("Failed to get ca cert: %v", cacerterr)
		return cacerterr
	}
	_, cakeyerr := os.Stat(certbundle.Ca.Key)
	if cakeyerr != nil && !os.IsNotExist(cakeyerr) {
		log.Errorf("Failed to get ca key: %v", cakeyerr)
		return cakeyerr
	}

	if os.IsNotExist(cacerterr) && os.IsNotExist(cakeyerr) {
		// generate root ca
		generr := GenerateCertBundle(&certbundle, CACERT, "")
		if generr != nil {
			log.Errorf("Failed to generate root ca cert: %v", generr)
			return generr
		}
	}
	// sign server cert and client cert if not provided
	if certbundle.Server != nil {
		_, bundlecacerterr := os.Stat(certbundle.Ca.Cert)
		if bundlecacerterr != nil {
			log.Errorln("Ca Cert Failed:", bundlecacerterr)
		}

		_, cakeyerr = os.Stat(certbundle.Ca.Key)
		_, servercerterr := os.Stat(certbundle.Server.Cert)
		if os.IsNotExist(servercerterr) {
			if cakeyerr == nil {
				generr := GenerateCertBundle(&certbundle, SERVERCERT, hosts)
				if generr != nil {
					return generr
				}
			} else {
				log.Errorf("Failed to sign server certificate: %v", cakeyerr)
				return cakeyerr
			}
		}
	}
	if certbundle.Client != nil {
		if certbundle.Client.Cert != "" {
			_, clientcerterr := os.Stat(certbundle.Client.Cert)
			if os.IsNotExist(clientcerterr) {
				if cakeyerr == nil {
					generr := GenerateCertBundle(&certbundle, CLIENTCERT, "")
					if generr != nil {
						return generr
					}
				} else {
					log.Errorf("Failed to sign client certificate: %v", cakeyerr)
					return cakeyerr
				}
			}
		}
	}

	err = writeCertBundleToConfigFile(certbundle)
	if err != nil {
		return err
	}
	return nil
}

func generateECDSAPrivKey(cb *cmapi.Certificate, ctype CertType) (*ecdsa.PrivateKey, error) {
	// Set default private algo
	var pkeyAlgo = "ecdsa-p384"

	switch ctype {
	case CACERT:
		// Check if csr file exists
		if _, err := os.Stat(cb.Ca.Csr); err == nil {
			cacsr := &Certcsr{}

			if err := eputils.LoadJsonFromFile(cb.Ca.Csr, cacsr); err != nil {
				log.Error(err)
				return nil, err
			}
			if cacsr.Key.Algo != "" {
				pkeyAlgo = cacsr.Key.Algo
			}
		} else {
			return nil, err
		}
	case SERVERCERT:
		// Check if csr file exists
		if _, err := os.Stat(cb.Server.Csr); err == nil {
			servercsr := &Certcsr{}
			if err := eputils.LoadJsonFromFile(cb.Server.Csr, servercsr); err != nil {
				log.Error(err)
				return nil, err
			}
			if servercsr.Key.Algo != "" {
				pkeyAlgo = servercsr.Key.Algo
			}
		} else {
			return nil, err
		}
	case CLIENTCERT:
		// Check if csr file exists
		if _, err := os.Stat(cb.Client.Csr); err == nil {
			clientcsr := &Certcsr{}
			if err := eputils.LoadJsonFromFile(cb.Client.Csr, clientcsr); err != nil {
				log.Error(err)
				return nil, err
			}
			if clientcsr.Key.Algo != "" {
				pkeyAlgo = clientcsr.Key.Algo
			}
		} else {
			return nil, err
		}
	default:
		log.Errorf("Unsupported certificate type: %v", ctype)
		return nil, eputils.GetError("errCertType")
	}
	switch pkeyAlgo {
	case "ecdsa-p521":
		priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			log.Errorf("Failed to generate private key: %v", err)
		}
		return priv, err
	case "ecdsa-p384":
		priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			log.Errorf("Failed to generate private key: %v", err)
		}
		return priv, err
	default:
		log.Errorf("Unsupported key algo: " + pkeyAlgo)
		return nil, eputils.GetError("errKeyAlgo")
	}
}

func prepareCertTemplate(cb *cmapi.Certificate, ctype CertType, hosts string) (*x509.Certificate, error) {
	keyUsage := x509.KeyUsageDigitalSignature
	keyUsage |= x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Errorf("Failed to generate serial number: %v", err)
		return nil, err
	}

	var subname string
	var extKeyu []x509.ExtKeyUsage
	usercsr := new(Certcsr)
	if ctype == CACERT {
		if _, err := os.Stat(cb.Ca.Csr); err == nil {
			if err := eputils.LoadJsonFromFile(cb.Ca.Csr, usercsr); err != nil {
				log.Error(err)
				return nil, err
			}
			subname = usercsr.Cn
		} else {
			return nil, err
		}
	} else if ctype == SERVERCERT {
		if _, err := os.Stat(cb.Server.Csr); err == nil {
			if err := eputils.LoadJsonFromFile(cb.Server.Csr, usercsr); err != nil {
				log.Error(err)
				return nil, err
			}
			subname = usercsr.Cn
		} else {
			return nil, err
		}
		extKeyu = append(extKeyu, x509.ExtKeyUsageServerAuth)
		// Use serving cert for client auth if the same csr file
		// is used for server and client
		if cb.Client != nil && cb.Server.Csr == cb.Client.Csr {
			extKeyu = append(extKeyu, x509.ExtKeyUsageClientAuth)
		}
	} else if ctype == CLIENTCERT {
		if _, err := os.Stat(cb.Client.Csr); err == nil {
			if err := eputils.LoadJsonFromFile(cb.Client.Csr, usercsr); err != nil {
				log.Error(err)
				return nil, err
			}
			subname = usercsr.Cn
		} else {
			return nil, err
		}
		extKeyu = append(extKeyu, x509.ExtKeyUsageClientAuth)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: subname,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		SignatureAlgorithm:    x509.ECDSAWithSHA384,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyu,
		BasicConstraintsValid: true,
	}

	if ctype == CACERT {
		template.SignatureAlgorithm = x509.ECDSAWithSHA512
	} else if ctype == SERVERCERT {
		certHosts := append(strings.Split(hosts, ","), usercsr.Hosts...)
		for _, h := range certHosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	}
	return &template, nil
}

func writeCertBundleToConfigFile(certbundle cmapi.Certificate) error {
	if _, err := os.Stat(RUNTIMECFGDIR); os.IsNotExist(err) {
		if err = os.MkdirAll(RUNTIMECFGDIR, os.ModePerm); err != nil {
			log.Error(err)
			return err
		}
	}
	filename := RUNTIMECFGDIR + "/" + certbundle.Name + "-cert.yaml"
	err := eputils.SaveSchemaStructToYamlFile(&certbundle, filename)
	if err != nil {
		log.Errorf("Write file %s err: %s", filename, err)
		return err
	}
	return nil
}

func getCertBundleFromConfigFile(certCfgFile string) (*cmapi.Certificate, error) {
	certbundle := &cmapi.Certificate{}
	err := eputils.LoadSchemaStructFromYamlFile(certbundle, certCfgFile)
	if err != nil {
		log.Errorf("Read file %s err: %s", certCfgFile, err)
		return nil, err
	}
	return certbundle, err
}
