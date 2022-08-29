/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package service

import (
	cmapi "ep/pkg/api/certmgr"
	epplugins "ep/pkg/api/plugins"
	certmgr "ep/pkg/certmgr"
	"ep/pkg/eputils"
	kubeutils "ep/pkg/eputils/kubeutils"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	SVCTLSEXT  = "service-tls"
	CSRFOLDER  = "config/certificate/component"
	CERTFOLDER = "cert/pki"

	SVCNAME = "service-name"

	SVRCSRFILEKEY = "csr-filename"
	CLTCSRFILEKEY = "client-csr-filename"

	CASECRETNAME  = "ca-secret-name"
	SVRSECRETNAME = "tls-secret-name"
	// #nosec G101
	CLTSECRETNAME = "client-tls-secret-name"
)

func GenSvcTLSCertFromTLSExtension(exts []*epplugins.EpParamsExtensionsItems0, tgtSvc string) error {
	for _, ext := range exts {
		if ext.Name == SVCTLSEXT {
			if ext.Extension == nil {
				return eputils.GetError("errExtNotFound")
			}
			for _, ext_svc := range ext.Extension.Extension {
				if ext_svc.Config == nil {
					return eputils.GetError("errExtCfgNotFound")
				}
				tls_cfg_found := false
				for _, cfg_instance := range ext_svc.Config {
					// Found the target service's tls config
					if cfg_instance.Name == SVCNAME && cfg_instance.Value == tgtSvc {
						tls_cfg_found = true
						break
					}
				}
				if !tls_cfg_found {
					continue
				}

				// Prepare CertBundle config
				wfcert, _, err := certmgr.GetCertBundleByName("workflow", "ca")
				if err != nil {
					return err
				}

				svccerts := cmapi.Certificate{
					Name: ext_svc.Name,
					Ca:   wfcert.Ca,
				}

				svrcsrname_found := false
				for _, ext_svc_cfg := range ext_svc.Config {
					// Server Cert cfg
					if ext_svc_cfg.Name == SVRCSRFILEKEY {
						svrcsrname_found = true
						svccerts.Server = &cmapi.CertificateServer{
							Csr:  filepath.Join(CSRFOLDER, ext_svc_cfg.Value),
							Cert: filepath.Join(CERTFOLDER, ext_svc.Name, ext_svc.Name+".pem"),
							Key:  filepath.Join(CERTFOLDER, ext_svc.Name, ext_svc.Name+"-key.pem"),
						}
					}
					// Client Cert cfg
					if ext_svc_cfg.Name == CLTCSRFILEKEY {
						svccerts.Client = &cmapi.CertificateClient{
							Csr:  filepath.Join(CSRFOLDER, ext_svc_cfg.Value),
							Cert: filepath.Join(CERTFOLDER, ext_svc.Name, ext_svc.Name+"-client.pem"),
							Key:  filepath.Join(CERTFOLDER, ext_svc.Name, ext_svc.Name+"-client-key.pem"),
						}
					}
				}
				if !svrcsrname_found {
					log.Errorf("Service TLS " + ext_svc.Name + " error: CSR filename not found!")
					return eputils.GetError("errCSRFileNotFound")
				}
				// Gen Certs
				if err := certmgr.GenCertAndConfig(svccerts, ""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func GenSvcSecretFromTLSExtension(exts []*epplugins.EpParamsExtensionsItems0, tgtSvc, ns, kubeconfig string) error {
	for _, ext := range exts {
		if ext.Name == SVCTLSEXT {
			if ext.Extension == nil {
				return eputils.GetError("errExtNotFound")
			}
			for _, ext_svc := range ext.Extension.Extension {
				if ext_svc.Config == nil {
					return eputils.GetError("errExtCfgNotFound")
				}
				tls_cfg_found := false
				for _, cfg_instance := range ext_svc.Config {
					// Found the target service's tls config
					if cfg_instance.Name == SVCNAME && cfg_instance.Value == tgtSvc {
						tls_cfg_found = true
						break
					}
				}
				if !tls_cfg_found {
					continue
				}

				log.Infof("Generating secret for %s", ext_svc.Name)
				// Get svc CertBundle config
				cb, _, err := certmgr.GetCertBundleByName(ext_svc.Name, "")
				if err != nil {
					return err
				}
				for _, ext_svc_cfg := range ext_svc.Config {
					secretName := ""
					tlscertf := ""
					tlskeyf := ""
					cacertf := cb.Ca.Cert
					switch ext_svc_cfg.Name {
					case CASECRETNAME:
						secretName = ext_svc_cfg.Value
						tlscertf = cb.Ca.Cert
						tlskeyf = cb.Ca.Key
					case SVRSECRETNAME:
						secretName = ext_svc_cfg.Value
						tlscertf = cb.Server.Cert
						tlskeyf = cb.Server.Key
					case CLTSECRETNAME:
						secretName = ext_svc_cfg.Value
						tlscertf = cb.Client.Cert
						tlskeyf = cb.Client.Key
					}
					if secretName == "" || tlscertf == "" || tlskeyf == "" || cacertf == "" {
						continue
					}

					cacert, err := ioutil.ReadFile(cacertf)
					if err != nil {
						log.Errorf("Failed to open: %s", cacertf)
						return err
					}

					tlscert, err := ioutil.ReadFile(tlscertf)
					if err != nil {
						log.Errorf("Failed to open: %s", tlscertf)
						return err
					}

					tlskey, err := ioutil.ReadFile(tlskeyf)
					if err != nil {
						log.Errorf("Failed to open: %s", tlskeyf)
						return err
					}

					tlsSecret, err := kubeutils.NewSecret(ns, secretName, "", kubeconfig)
					if err != nil {
						return err
					}

					err = tlsSecret.New()
					if err != nil {
						return err
					}

					err = tlsSecret.RenewStringData("ca.crt", string(cacert))
					if err != nil {
						return err
					}

					err = tlsSecret.RenewStringData("tls.crt", string(tlscert))
					if err != nil {
						return err
					}

					err = tlsSecret.RenewStringData("tls.key", string(tlskey))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
