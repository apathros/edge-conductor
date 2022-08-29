/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
//nolint: dupl
package certmgr_test

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	cmapi "ep/pkg/api/certmgr"
	certmgr "ep/pkg/certmgr"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errForce = errors.New("Force error")
)

func unpatch(m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		Fail(fmt.Sprintf("%s", err))
	}
}

func TestCertService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test certmgr utils")
}

const (
	CACERT      certmgr.CertType = 0
	SERVERCERT  certmgr.CertType = 1
	CLIENTCERT  certmgr.CertType = 2
	UNKNOWNCERT certmgr.CertType = 99
)

const (
	TESTCACSR         = "testdata/context1/ca-csr.json"
	TESTCAKEY         = "testdata/context1/ca.key"
	TESTCACRT         = "testdata/context1/ca.crt"
	WRONGCACRT        = "testdata/context1/wrongca.crt"
	WRONGCSR          = "testdata/context1/wrongcsr.json"
	FAKECERT          = "testdata/fake.crt"
	FAKEKEY           = "testdata/fake.key"
	TESTWFSERVERKEY   = "testdata/context1/server.key"
	TESTWFSERVERCSR   = "testdata/context1/server-csr.json"
	TESTWFSERVCLICSR  = "testdata/context1/servcli-csr.json"
	TESTWFSERVERCRT   = "testdata/context1/server.crt"
	TESTWFCLIENTKEY   = "testdata/context1/client.key"
	TESTWFCLIENTCSR   = "testdata/context1/client-csr.json"
	TESTWFCLIENTCRT   = "testdata/context1/client.crt"
	DUMMYPATH         = "testcertmgrt"
	DUMMYDIR          = "testdata/context1"
	HOSTNAME          = "localhost"
	TESTRUNTIMEPKIDIR = "testdata/cert/pki"
	TESTRUNTIMECFGDIR = "testdata/runtime/config"
	UNAUTHFILE        = "testdata/context1/unauthtext"
)

var _ = Describe("Check certs on root-ca and server", func() {
	var (
		initcerts cmapi.Certificate
	)

	Context("Verify CSR json file", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  TESTWFCLIENTCSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
			By("GenCertAndConfig")
			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).To(BeNil())
		})

		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.Remove(TESTWFCLIENTKEY)
			os.Remove(TESTWFCLIENTCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check if json file exist", func() {
			By("Check if csr file exist")
			Expect(initcerts.Ca.Csr).NotTo(BeNil())
			Expect(initcerts.Server.Csr).NotTo(BeNil())
			Expect(initcerts.Ca.Key).NotTo(BeNil())
			Expect(initcerts.Ca.Cert).NotTo(BeNil())
			Expect(initcerts.Server.Key).NotTo(BeNil())
			Expect(initcerts.Server.Cert).NotTo(BeNil())
		})
	})

	Context("Verify same server and client CSR json file", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVCLICSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  TESTWFSERVCLICSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
			By("GenCertAndConfig")
			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).To(BeNil())
		})

		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.Remove(TESTWFCLIENTKEY)
			os.Remove(TESTWFCLIENTCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check if json file exist", func() {
			By("Check if csr file exist")
			Expect(initcerts.Ca.Csr).NotTo(BeNil())
			Expect(initcerts.Server.Csr).NotTo(BeNil())
			Expect(initcerts.Ca.Key).NotTo(BeNil())
			Expect(initcerts.Ca.Cert).NotTo(BeNil())
			Expect(initcerts.Server.Key).NotTo(BeNil())
			Expect(initcerts.Server.Cert).NotTo(BeNil())
		})
	})

	Context("Verify ca CSR json file existence", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  "",
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  "",
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  "",
					Key:  TESTWFCLIENTKEY,
				},
			}
			By("GenCertAndConfig")
			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).NotTo(BeNil())
		})
	})

	Context("Verify server CSR json file existence", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  "",
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  "",
					Key:  TESTWFCLIENTKEY,
				},
			}
			By("GenCertAndConfig")
			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).NotTo(BeNil())
		})
	})

	Context("Verify client CSR json file existence", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  "",
					Key:  TESTWFCLIENTKEY,
				},
			}
			By("GenCertAndConfig")
			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).NotTo(BeNil())
		})
	})

	Context("Make mkdir fail", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
			}
		})
		It("Check mkdir func logic", func() {
			By("force mkdir error")
			patch, _ := mpatch.PatchMethod(os.MkdirAll, func(name string, perm os.FileMode) error { return errForce })
			defer unpatch(patch)
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			errGenCertAndConfig := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(errGenCertAndConfig).ToNot(BeNil())

			_, err := os.Stat(TESTRUNTIMEPKIDIR)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("check ca crt and key", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca:   &cmapi.CertificateCa{},
			}

		})
		It("check ca crt and key", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give worng path for ca", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: DUMMYPATH,
					Csr:  TESTCACSR,
					Key:  DUMMYPATH,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  TESTWFCLIENTCSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(DUMMYPATH)
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.Remove(TESTWFCLIENTKEY)
			os.Remove(TESTWFCLIENTCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give wrong csr 1", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  WRONGCSR,
					Key:  TESTCAKEY,
				},
			}
		})

		It("Check csr", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give wrong csr 2", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  WRONGCSR,
					Key:  TESTWFSERVERKEY,
				},
			}
		})

		It("Check csr", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give wrong csr 3", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  WRONGCSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
		})

		It("Check csr", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 1", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: DUMMYDIR,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 2", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  DUMMYDIR,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 3", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: DUMMYDIR,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 4", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  DUMMYDIR,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 5", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: DUMMYDIR,
					Csr:  TESTWFCLIENTCSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give dir path 6", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  TESTWFCLIENTCSR,
					Key:  DUMMYDIR,
				},
			}
		})

		It("Check cert/key path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func rand_Int error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  TESTWFSERVERCSR,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  TESTWFCLIENTCSR,
					Key:  TESTWFCLIENTKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.Remove(TESTWFCLIENTKEY)
			os.Remove(TESTWFCLIENTCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(rand.Int, func(rand io.Reader, max *big.Int) (*big.Int, error) { return nil, errForce })
			defer unpatch(patch)
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)
			Expect(err).ToNot(BeNil())

		})
	})

	Context("force func x509_CreateCertificate  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(x509.CreateCertificate, func(rand io.Reader, template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
				return nil, errForce
			})
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func ioutil_ReadFile  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { return nil, errForce })
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func x509_ParsePKCS8PrivateKey  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(x509.ParsePKCS8PrivateKey, func(der []byte) (interface{}, error) { return nil, errForce })
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func os_OpenFile  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) { return nil, errForce })
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func x509_MarshalPKCS8PrivateKey  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(x509.MarshalPKCS8PrivateKey, func(key interface{}) ([]byte, error) { return nil, errForce })
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("force func pem_Encode  error", func() {
		BeforeEach(func() {
			By("Create instances for certs")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  TESTCACSR,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
			}
		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check server ca crt with error func", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			patch, _ := mpatch.PatchMethod(pem.Encode, func(out io.Writer, b *pem.Block) error { return errForce })
			defer unpatch(patch)

			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

	Context("Give worng path for ca's , server's and client's crt", func() {
		BeforeEach(func() {
			By("Give worng path for ca's , server's and client's crt")
			initcerts = cmapi.Certificate{
				Name: "workflow",
				Ca: &cmapi.CertificateCa{
					Cert: TESTCACRT,
					Csr:  DUMMYPATH,
					Key:  TESTCAKEY,
				},
				Server: &cmapi.CertificateServer{
					Cert: TESTWFSERVERCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFSERVERKEY,
				},
				Client: &cmapi.CertificateClient{
					Cert: TESTWFCLIENTCRT,
					Csr:  DUMMYPATH,
					Key:  TESTWFCLIENTKEY,
				},
			}

		})
		AfterEach(func() {
			By("delete the created file during testing")
			os.Remove(DUMMYPATH)
			os.Remove(TESTCAKEY)
			os.Remove(TESTCACRT)
			os.Remove(TESTWFSERVERKEY)
			os.Remove(TESTWFSERVERCRT)
			os.Remove(TESTWFCLIENTKEY)
			os.Remove(TESTWFCLIENTCRT)
			os.RemoveAll(TESTRUNTIMEPKIDIR)
		})

		It("Check crt path", func() {
			By("GenCertAndConfig")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()
			err := certmgr.GenCertAndConfig(initcerts, HOSTNAME)

			Expect(err).ToNot(BeNil())
		})
	})

})

var _ = Describe("converts an RSA public key to PKCS1  form", func() {

	var fakeCertData = []byte(`
-----BEGIN CERTIFICATE-----
MIICmzCCAYMCCQCbDmElUJL8zDANBgkqhkiG9w0BAQsFADAQMQ4wDAYDVQQDDAVD
SEFOVDAeFw0yMTA5MTYwMjAzMTRaFw0yMTEwMTYwMjAzMTRaMA8xDTALBgNVBAMM
BEZBS0UwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDMaSGhK9NYarqG
ilE0VwBdux9QDtyE2ype3Rc9WzYadu0EGHPsyaYS+JFA4AyMHfwH6Io4tJtargzX
LPFKWq6oQgZ4hMQl+TbS64quZCNw/G8dFGT4khyEQvNPaxoZSvuDtr4uzcU5g4M2
9P2+OZH6+LaykMgxE9dmTAFX0s4L4ahd+H46qYC0KOrVHKyeYasSsbNZSGwWwOS6
juQKZJyUFHo0zeMi0TCxXfGBlJ6zynebqzJ1/GikTe8ZjoQoDkyPHboOgdd+gi4b
QhnZrlYppFJIOc3S9ZwvA6ThMgJ6l8mQn5Lnj4lh+jsH+Bno5fvtYd1N5rXEvlSc
MrLLcUrFAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKgZ4v9qFNiR9n9q+8ozfAop
uNwJtoV8Av6BX8uWUHzHCwL/icbvoY/yUxBKN8Dj10k6WN0GaR7OlqoclkHXCJ8P
akwUUOkf9uERF3+7/+rmhhiwDz2Qhid5MRbirihL7Tz4CtIeqBXRAwsD1batf+Z4
f4fhHAYpbzVZWytIrP9nhS9FMj9HbDn+VI6tDtD+yoiHW4FN7PgdpGpB24KjiigI
LHzDbPVzdLMyQI3FOTYbj8YFAR0RY8f+MBUeGufoe0UNAXK7IMeN6oArmqGjUfS7
PCa+pDeuygboJdtxhfSaU5xqvdydrcrHWub/7CoWICmT3xCiI8qib3WljvkfY+4=
-----END CERTIFICATE-----
`)
	Context("load cert and convert an RSA public key to PKCS1  form", func() {
		It("Check hash public key", func() {
			By("Check hash public key")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			CASHA256, _ := certmgr.GenCACertSHA256(fakeCertData)
			Expect(CASHA256).NotTo(BeNil())
		})
	})

	Context("load cert and convert an RSA public key to PKCS1 form", func() {
		It("Check hash public key", func() {
			By("force x509 error")
			patch, _ := mpatch.PatchMethod(x509.ParseCertificate, func(der []byte) (*x509.Certificate, error) { return nil, errForce })
			defer unpatch(patch)
			By("Check hash public key")
			_, err := certmgr.GenCACertSHA256(fakeCertData)
			Expect(err).ToNot(BeNil())
		})
	})

})

var _ = Describe("converts an ECDSA public key to PKIX form", func() {

	var fakeCertData = []byte(`
-----BEGIN CERTIFICATE-----
MIICCjCCAWygAwIBAgIRAPADqzWI5FWISoo2nMoSNMowCgYIKoZIzj0EAwQwITEf
MB0GA1UEAxMWRWRnZSBDb25kdWN0b3IgUm9vdCBDQTAeFw0yMjAzMTEwOTA3NDVa
Fw0yMzAzMTEwOTA3NDVaMCExHzAdBgNVBAMTFkVkZ2UgQ29uZHVjdG9yIFJvb3Qg
Q0EwgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABABS5m7CcoLtyPOfjzexY6M1GSzX
KEluK89IY7RrXGB+Eha5hfbkPPcnBYGg7vt/T/kk5RRMOzDFCLr8p8TlusP9FwA1
FanLvFtrObWHHp+9xx/3MbRFMgXt68DmwitagocsxC9Y1s3cpz+fsnVI+irraLrd
GS5VqWiOO8+nJY9Wc3WiqKNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQF
MAMBAf8wHQYDVR0OBBYEFLGWg7/acDhSSRIzQDy6j0nGHvyLMAoGCCqGSM49BAME
A4GLADCBhwJCAZZkT3CP/RIP/4AvwQBBuV5QpOBm/XwglffPAF9Yl5f7Slatal0P
ahyax6wzXjGN3iiRxbAZWFfj7mFuaxZFLundAkFFRDeQZFgV1zU30n1at4/xXgEq
cTDfXspHiWTooOvN9wVzql2u1knuN+MHUnD+qQwJDDwEYEvyz2AyCOGJpSxXiA==
-----END CERTIFICATE-----
`)
	Context("load cert and convert an ECDSA public key to PKIX  form", func() {
		It("Check hash public key", func() {
			By("Check hash public key")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			CASHA512, _ := certmgr.GenCACertSHA512(fakeCertData)
			Expect(CASHA512).NotTo(BeNil())
		})
	})

	Context("load cert and convert an ECDSA public key to PKIX form", func() {
		It("Check hash public key", func() {
			By("force x509 error")
			patch, _ := mpatch.PatchMethod(x509.ParseCertificate, func(der []byte) (*x509.Certificate, error) { return nil, errForce })
			defer unpatch(patch)
			By("Check hash public key")
			_, err := certmgr.GenCACertSHA512(fakeCertData)
			Expect(err).ToNot(BeNil())
		})
	})

})

var _ = Describe("Handle TLS certification", func() {

	Context("convert an ECDSA public key for server", func() {
		It("Check TLS config", func() {
			By("Check TLS config")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "server", "127.0.0.1")
			Expect(err).To(BeNil())
		})
	})

	Context("convert an ECDSA public key for client ", func() {
		It("Check TLS config on client", func() {
			By("Check TLS config on client")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "client", "127.0.0.1")
			Expect(err).To(BeNil())
		})
	})

	Context("convert an ECDSA public key for otherelse", func() {
		It("Check TLS config on otherelse", func() {
			By("Check TLS config on otherelse")
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "otherelse", "127.0.0.1")
			Expect(err).To(BeNil())
		})
	})

	Context("force error on setupTLSConfig", func() {
		It("force error on setupTLSConfig", func() {
			By("force error on setupTLSConfig")
			patch, _ := mpatch.PatchMethod(certmgr.SetupTLSConfig, func(certCfg certmgr.TLSCertConfig, tlsConfig *tls.Config) error { return errForce })
			defer unpatch(patch)
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			By("Check hash public key")
			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "server", "127.0.0.1")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("force error on ReadFile", func() {
		It("force error on ReadFile", func() {
			By("force error on ReadFile")
			patch, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { return nil, errForce })
			defer unpatch(patch)
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			By("Check hash public key")
			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "server", "127.0.0.1")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("force error on LoadX509KeyPair", func() {
		It("force error on LoadX509KeyPair", func() {
			By("force error on LoadX509KeyPair")
			patch, _ := mpatch.PatchMethod(tls.LoadX509KeyPair, func(certFile, keyFile string) (tls.Certificate, error) {
				return tls.Certificate{}, errForce
			})
			defer unpatch(patch)
			stub := gostub.Stub(&certmgr.RUNTIMEPKIDIR, TESTRUNTIMEPKIDIR).Stub(&certmgr.RUNTIMECFGDIR, TESTRUNTIMECFGDIR)
			defer stub.Reset()

			By("Check hash public key")
			_, err := certmgr.GetTLSConfigByName("fakeworkflow", "server", "127.0.0.1")
			Expect(err).ToNot(BeNil())
		})
	})

})
