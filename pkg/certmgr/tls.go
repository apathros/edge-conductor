/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package certmgr

import (
	"crypto/tls"
	"crypto/x509"
	cmapi "ep/pkg/api/certmgr"
	"ep/pkg/eputils"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	RUNTIMECFGDIR = "runtime/config"
)

type TLSCertConfig struct {
	CAFile        string
	CertFile      string
	KeyFile       string
	ServerAddress string
	IsServer      bool
}

func getCertCfgFileFromName(cname, ctype string) (string, CertType) {
	certFile := RUNTIMECFGDIR + "/" + cname + "-cert.yaml"
	var cType CertType
	switch ctype {
	case "server":
		cType = SERVERCERT
	case "client":
		cType = CLIENTCERT
	case "ca":
		cType = CACERT
	default:
		cType = UNKNOWNCERT
	}
	return certFile, cType
}

func GetCertBundleByName(cname, ctype string) (*cmapi.Certificate, CertType, error) {
	certCfgFile, ct := getCertCfgFileFromName(cname, ctype)
	if _, err := os.Stat(certCfgFile); os.IsNotExist(err) {
		log.Errorf("Cert config file %s not found!", certCfgFile)
		return nil, UNKNOWNCERT, err
	}
	certBundle, err := getCertBundleFromConfigFile(certCfgFile)
	if err != nil {
		return nil, UNKNOWNCERT, err
	}
	return certBundle, ct, nil
}

func GetTLSConfigByName(cname, ctype, serverAddr string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	certBundle, certType, err := GetCertBundleByName(cname, ctype)
	if err != nil {
		return nil, err
	}
	tlsCertCfg := &TLSCertConfig{}
	tlsCertCfg.CAFile = certBundle.Ca.Cert
	if certType == SERVERCERT {
		tlsCertCfg.CertFile = certBundle.Server.Cert
		tlsCertCfg.KeyFile = certBundle.Server.Key
		tlsCertCfg.ServerAddress = serverAddr
		tlsCertCfg.IsServer = true
	} else if certType == CLIENTCERT {
		tlsCertCfg.CertFile = certBundle.Client.Cert
		tlsCertCfg.KeyFile = certBundle.Client.Key
		tlsCertCfg.ServerAddress = ""
		tlsCertCfg.IsServer = false
	} else {
		return nil, nil
	}
	err = SetupTLSConfig(*tlsCertCfg, tlsConfig)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func SetupTLSConfig(certCfg TLSCertConfig, tlsConfig *tls.Config) error {
	var err error
	if certCfg.CAFile != "" {
		rootPEM, err := ioutil.ReadFile(certCfg.CAFile)
		if err != nil {
			return err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(rootPEM)
		if !ok {
			log.Errorf("failed to parse root certificate: %q", certCfg.CAFile)
			return eputils.GetError("errRootCert")
		}
		tlsConfig.RootCAs = roots
		tlsConfig.MinVersion = tls.VersionTLS13
		tlsConfig.CurvePreferences = []tls.CurveID{tls.CurveP521}
		tlsConfig.PreferServerCipherSuites = true
		if certCfg.IsServer {
			tlsConfig.ClientCAs = roots
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ServerName = certCfg.ServerAddress
		}
	}
	if certCfg.CertFile != "" && certCfg.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(
			certCfg.CertFile,
			certCfg.KeyFile,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
