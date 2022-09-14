# How to use Cert-manager

This document is about how we can use cert-manager for managing certificates and enabling TLS connection for services inside the cluster.

## Preparation

Follow [HW Requirements for Edge Conductor Day-0 Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and System Requirements for Edge Conductor Day-0 Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host) to **prepare the Day-0 host** hardware and software.

Follow [Build and Install Edge Conductor Tool](../../README.md#download-and-build-edge-conductor-code-internal-users-only) to **download and build Edge Conductor tool**. Enter `_workspace` folder to run Edge Conductor tool.

Ensure the verification of installation of Cert-manager component by using the [cmctl](https://cert-manager.io/docs/installation/verify/) command as follows

               .../_workspace$ cmctl check api
                The cert-manager API is ready 

Once the cert-manager API is verified, we can proceed to further steps.

## Installation and Uninstallation

Cert-manager is already integrated into the EC through component manifest & common.yml file and thus it will be installed during the EC service build and service deploy commands

In case of uninstallation, we need to remove the 2 segment of code from common.yml which are 

  - name: cert-manager
  - name: cert-manager-cluster-issuer


Ensuring this components are removed and service build and deploy commands are reused. We can see that cert-manager will be uninstalled in the EC. This can be further verified using the command "cmctl check api " and check if the output corresponds to the " the cert-manager crds are not yet installed on the kubernetes api server"


## Certificate creation using the Cert-Manager


Cert-manager can be also used to provision certificates using yaml based configuration. For the Edge Conductor, a cluster-wide issuer is used to provision certificates. Following example shows how we can create a Certificate by requesting from cert-manager API.


    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
    name: example-com
    namespace: sandbox
    spec:
    # Secret names are always required.
    secretName: example-com-tls

    # secretTemplate is optional. If set, these annotations and labels will be
    # copied to the Secret named example-com-tls. These labels and annotations will
    # be re-reconciled if the Certificate's secretTemplate changes. secretTemplate
    # is also enforced, so relevant label and annotation changes on the Secret by a
    # third party will be overwriten by cert-manager to match the secretTemplate.
    secretTemplate:
        annotations:
        my-secret-annotation-1: "foo"
        my-secret-annotation-2: "bar"
        labels:
        my-secret-label: foo

    duration: 2160h # 90d
    renewBefore: 360h # 15d
    subject:
        organizations:
        - jetstack
    # The use of the common name field has been deprecated since 2000 and is
    # discouraged from being used.
    commonName: example.com
    isCA: false
    privateKey:
        algorithm: RSA
        encoding: PKCS1
        size: 2048
    usages:
        - server auth
        - client auth
    # At least one of a DNS Name, URI, or IP address is required.
    dnsNames:
        - example.com
        - www.example.com
    uris:
        - spiffe://cluster.local/ns/sandbox/sa/example
    ipAddresses:
        - 192.168.0.5
    # Issuer references are always required.
    issuerRef:
        name: ca-issuer
        # We can reference ClusterIssuers by changing the kind here.
        # The default value is Issuer (i.e. a locally namespaced Issuer)
        kind: Issuer
        # This is optional since cert-manager will default to this value however
        # if you are using an external issuer, change this to that issuer group.
        group: cert-manager.io



For further information regarding renewal and revocation, kindly follow the documentation. https://cert-manager.io/docs/usage/certificate/#creating-certificate-resources

## Integration with Kubernetes services

- ###  Prometheus
We divide the entire integration of Cert-manager with Prometheus in 2 components. Part 1 describe how to enable TLS support and part 2 describes security verification.


  - ####  Part 1. Prometheus - Enable TLS

For ensuring TLS communication through the Prometheus service, we need to mount the Kubernetes Secrets to the Pods and also perform addition configuration in helm chart using override file 


Please refer to following URLs for better understanding of the configuration
  1. [Understanding usage of WebTLS component](https://prometheus.io/docs/prometheus/latest/configuration/https/)
  2. [WebTLS configuration](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#webtlsconfig)
  3. [Kube-Prometheus-Stack values](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
  4. [Operator Webhooks](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/user-guides/webhook.md)
            
            
The following configuration enables Prometheus Operator to use Cert-manager for creating certificates. Note that certificates are created with Intel TLS guidelines and Cert-manager spec. The code will create a secret called "prometheus-kube-prometheus-admission" in Prometheus namespace.


    prometheusOperator:
      admissionWebhooks:
        certManager:
          enabled: true
          rootCert:
            duration: "26280h"
          admissionCert:
            duration: "26280h"
          issuerRef:
            name: "edge-conductor-ca"
            kind: "ClusterIssuer"


Once secret is created, it can be mounted and used in WebTLS configuration to allow HTTPs connection to Prometheus service.

    prometheus:
      prometheusSpec:
        externalUrl: "https://prometheus-kube-prometheus-prometheus.prometheus:9090"
        secrets: ['prometheus-kube-prometheus-admission']
        web:
          tlsConfig:
            keySecret:
              name: prometheus-kube-prometheus-admission
              key: tls.key
            cert:
              secret:
                name: prometheus-kube-prometheus-admission
                key: tls.crt


- #### Part 2. Prometheus - Testing and validation

Once, the prometheus HTTPS service has been enabled, it can be easily validated by entering inside kubernetes cluster and and issusing the following command on one of its Pod's terminal.  


    curl -kv https://prometheus-kube-prometheus-prometheus.prometheus:9090
    

The following output shows the TLS handshakes and ensures TLS connection is established for sending HTTPS request to HTTPS server exposed by Prometheus Pods.

       Trying 10.96.77.70:9090...
    * Connected to prometheus-kube-prometheus-prometheus.prometheus (10.96.77.70) port 9090 (#0)
    * ALPN, offering h2
    * ALPN, offering http/1.1
    * successfully set certificate verify locations:
    *  CAfile: /etc/ssl/certs/ca-certificates.crt
    *  CApath: /etc/ssl/certs
    * TLSv1.3 (OUT), TLS handshake, Client hello (1):
    * TLSv1.3 (IN), TLS handshake, Server hello (2):
    * TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
    * TLSv1.3 (IN), TLS handshake, Certificate (11):
    * TLSv1.3 (IN), TLS handshake, CERT verify (15):
    * TLSv1.3 (IN), TLS handshake, Finished (20):
    * TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
    * TLSv1.3 (OUT), TLS handshake, Finished (20):
    * SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
    * ALPN, server accepted to use h2
    * Server certificate:
    *  subject: CN=prometheus-ca-default
    *  start date: Aug 10 13:54:49 2022 GMT
    *  expire date: Aug  9 13:54:49 2025 GMT
    *  issuer: CN=edgecon-ca-default
    *  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
    * Using HTTP2, server supports multi-use
    * Connection state changed (HTTP/2 confirmed)
    * Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
    * Using Stream ID: 1 (easy handle 0x55a64f8562c0)
    > GET / HTTP/2
    > Host: prometheus-kube-prometheus-prometheus.prometheus:9090
    > user-agent: curl/7.74.0
    > accept: */*
    > 
    * TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
    * Connection state changed (MAX_CONCURRENT_STREAMS == 250)!
    < HTTP/2 302 
    < content-type: text/html; charset=utf-8
    < location: /graph
    < content-length: 29
    < date: Mon, 22 Aug 2022 10:23:50 GMT
    < 
    <a href="/graph">Found</a>.

    * Connection #0 to host prometheus-kube-prometheus-prometheus.prometheus left intact
  
- ### Node Feature Discovery.

 - #### Node Feature Discovery - Enable TLS

Please use the [official documentation](https://kubernetes-sigs.github.io/node-feature-discovery/stable/get-started/index.html) for understanding the Node Feature Discovery in depth. The cert-manager is enabled using the [doc](https://kubernetes-sigs.github.io/node-feature-discovery/stable/get-started/index.html)

The following code in NFD override file enables TLS in the NFD services.

    tls:
      enable: true
    {{- if ( has "cert-manager" .Kitconfig.Components.Selector) }}
      certManager: true
    {{- else }}
      certManager: false
    {{- end }}

  - #### Node Feature Discovery - Testing and validation


Using Kubectl command, we can verify if the ceritificates are being created in the required namespace.

         ../_workspace$ ./kubectl get certificate -n node-feature-discovery                                          
              NAME              READY   SECRET            AGE
              nfd-ca-cert       True    nfd-ca-cert       50m
              nfd-master-cert   True    nfd-master-cert   50m
              nfd-worker-cert   True    nfd-worker-cert   50m

Additionally, the secrets can be also verified using 

         ../_workspace$ ./kubectl get secret -n node-feature-discovery | grep tls                                    
         nfd-ca-cert                                     kubernetes.io/tls                     3      50m
         nfd-master-cert                                 kubernetes.io/tls                     3      50m
         nfd-worker-cert                                 kubernetes.io/tls                     3      50m
         
Furthermore, the certificates are being created by Cert-manager itself can be verified using 


      ../_workspace$ ./kubectl describe certificate  -n node-feature-discovery
        Name:         nfd-ca-cert
        Namespace:    node-feature-discovery
        Labels:       app.kubernetes.io/managed-by=Helm
        Annotations:  meta.helm.sh/release-name: nfd
                      meta.helm.sh/release-namespace: node-feature-discovery
        API Version:  cert-manager.io/v1
        Kind:         Certificate
        Metadata:
          Creation Timestamp:  2022-08-22T10:40:17Z
          Generation:          1
          Managed Fields:
            API Version:  cert-manager.io/v1
            Fields Type:  FieldsV1
            fieldsV1:
              f:metadata:
                f:annotations:
                  .:
                  f:meta.helm.sh/release-name:
                  f:meta.helm.sh/release-namespace:
                f:labels:
                  .:
                  f:app.kubernetes.io/managed-by:
              f:spec:
                .:
                f:commonName:
                f:isCA:
                f:issuerRef:
                  .:
                  f:group:
                  f:kind:
                  f:name:
                f:secretName:
                f:subject:
                  .:
                  f:organizations:
            Manager:      conductor
            Operation:    Update
            Time:         2022-08-22T10:40:17Z
            API Version:  cert-manager.io/v1
            Fields Type:  FieldsV1
            fieldsV1:
              f:status:
                f:revision:
            Manager:      cert-manager-certificates-issuing
            Operation:    Update
            Subresource:  status
            Time:         2022-08-22T10:40:18Z
            API Version:  cert-manager.io/v1
            Fields Type:  FieldsV1
            fieldsV1:
              f:status:
                .:
                f:conditions:
                  .:
                  k:{"type":"Ready"}:
                    .:
                    f:lastTransitionTime:
                    f:message:
                    f:observedGeneration:
                    f:reason:
                    f:status:
                    f:type:
                f:notAfter:
                f:notBefore:
                f:renewalTime:
            Manager:         cert-manager-certificates-readiness
            Operation:       Update
            Subresource:     status
            Time:            2022-08-22T10:40:18Z
          Resource Version:  2733
          UID:               ca43a9ad-18bb-4c10-b14e-39422e84cca4
        Spec:
          Common Name:  nfd-ca-cert
          Is CA:        true
          Issuer Ref:
            Group:      cert-manager.io
            Kind:       Issuer
            Name:       nfd-ca-bootstrap
          Secret Name:  nfd-ca-cert
          Subject:
            Organizations:
              node-feature-discovery
        Status:
          Conditions:
            Last Transition Time:  2022-08-22T10:40:18Z
            Message:               Certificate is up to date and has not expired
            Observed Generation:   1
            Reason:                Ready
            Status:                True
            Type:                  Ready
          Not After:               2022-11-20T10:40:18Z
          Not Before:              2022-08-22T10:40:18Z
          Renewal Time:            2022-10-21T10:40:18Z
          Revision:                1
        Events:
          Type    Reason     Age    From                                       Message
          ----    ------     ----   ----                                       -------
          Normal  Issuing    3m55s  cert-manager-certificates-trigger          Issuing certificate as Secret does not exist
          Normal  Generated  3m54s  cert-manager-certificates-key-manager      Stored new private key in temporary Secret resource "nfd-ca-cert-vkk6m"
          Normal  Requested  3m54s  cert-manager-certificates-request-manager  Created new CertificateRequest resource "nfd-ca-cert-284th"
          Normal  Issuing    3m54s  cert-manager-certificates-issuing          The certificate has been successfully issued



To futher verify if the Pods are using HTTPs, we can check the arguments of the container/pod of NFD and verify that 3 components i.e. CA-cert, TLS key, TLS cert for ensuring TLS are being mounted to pods.

      ../_workspace$ ./kubectl get  po nfd-node-feature-discovery-master-cc47fc58c-r92z9 -n node-feature-discovery -o json | jq '.spec.containers[0].args'
      [
        "--extra-label-ns=gpu.intel.com",
        "--resource-labels=gpu.intel.com/memory.max,gpu.intel.com/millicores,gpu.intel.com/tiles",
        "-featurerules-controller=true",
        "--ca-file=/etc/kubernetes/node-feature-discovery/certs/ca.crt",
        "--key-file=/etc/kubernetes/node-feature-discovery/certs/tls.key",
        "--cert-file=/etc/kubernetes/node-feature-discovery/certs/tls.crt"
      ]



Also, describing pods verifies that HTTPs connection is ensured.

      ../_workspace$ ./kubectl  describe po  nfd-node-feature-discovery-worker-r2t92  -n node-feature-discovery
      Name:         nfd-node-feature-discovery-worker-r2t92
      Namespace:    node-feature-discovery
      Priority:     0
      Node:         192.114.9.100/192.114.9.100
      Start Time:   Mon, 22 Aug 2022 11:08:17 +0000
      Labels:       app.kubernetes.io/instance=nfd
                    app.kubernetes.io/name=node-feature-discovery
                    controller-revision-hash=7ffd9875dd
                    pod-template-generation=1
                    role=worker
      Annotations:  cni.projectcalico.org/containerID: 76be4e6fe8fbbb8be2b9b9f4f0af425d827572d57925afa4fa87f4e5fdcdddda
                    cni.projectcalico.org/podIP: 10.42.0.19/32
                    cni.projectcalico.org/podIPs: 10.42.0.19/32
                    k8s.v1.cni.cncf.io/network-status:
                      [{
                          "name": "k8s-pod-network",
                          "ips": [
                              "10.42.0.19"
                          ],
                          "default": true,
                          "dns": {}
                      }]
                    k8s.v1.cni.cncf.io/networks-status:
                      [{
                          "name": "k8s-pod-network",
                          "ips": [
                              "10.42.0.19"
                          ],
                          "default": true,
                          "dns": {}
                      }]
      Status:       Running
      IP:           10.42.0.19
      IPs:
        IP:           10.42.0.19
      Controlled By:  DaemonSet/nfd-node-feature-discovery-worker
      Containers:
        worker:
          Container ID:  docker://744fedfeb6013a7eabfbf9eb611e8323e5709e284570f1dd39a53608f5b07bd9
          Image:         k8s.gcr.io/nfd/node-feature-discovery:v0.11.0
          Image ID:      docker-pullable://k8s.gcr.io/nfd/node-feature-discovery@sha256:19e319110e5cf935f191ba5d841a6c40b899a1951dbacf0d098cf83de92d6364
          Port:          <none>
          Host Port:     <none>
          Command:
            nfd-worker
          Args:
            --server=nfd-node-feature-discovery-master:8080
            --ca-file=/etc/kubernetes/node-feature-discovery/certs/ca.crt
            --key-file=/etc/kubernetes/node-feature-discovery/certs/tls.key
            --cert-file=/etc/kubernetes/node-feature-discovery/certs/tls.crt
          State:          Running
            Started:      Mon, 22 Aug 2022 11:08:26 +0000
          Ready:          True
          Restart Count:  0
          Environment:
            NODE_NAME:   (v1:spec.nodeName)
          Mounts:
            /etc/kubernetes/node-feature-discovery from nfd-worker-conf (ro)
            /etc/kubernetes/node-feature-discovery/certs from nfd-worker-cert (ro)
            /etc/kubernetes/node-feature-discovery/features.d/ from features-d (ro)
            /etc/kubernetes/node-feature-discovery/source.d/ from source-d (ro)
            /host-boot from host-boot (ro)
            /host-etc/os-release from host-os-release (ro)
            /host-sys from host-sys (ro)
            /host-usr/lib from host-usr-lib (ro)
            /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-rmsmr (ro)
      Conditions:
        Type              Status
        Initialized       True
        Ready             True
        ContainersReady   True
        PodScheduled      True
      Volumes:
        host-boot:
          Type:          HostPath (bare host directory volume)
          Path:          /boot
          HostPathType:
        host-os-release:
          Type:          HostPath (bare host directory volume)
          Path:          /etc/os-release
          HostPathType:
        host-sys:
          Type:          HostPath (bare host directory volume)
          Path:          /sys
          HostPathType:
        host-usr-lib:
          Type:          HostPath (bare host directory volume)
          Path:          /usr/lib
          HostPathType:
        source-d:
          Type:          HostPath (bare host directory volume)
          Path:          /etc/kubernetes/node-feature-discovery/source.d/
          HostPathType:
        features-d:
          Type:          HostPath (bare host directory volume)
          Path:          /etc/kubernetes/node-feature-discovery/features.d/
          HostPathType:
        nfd-worker-conf:
          Type:      ConfigMap (a volume populated by a ConfigMap)
          Name:      nfd-node-feature-discovery-worker-conf
          Optional:  false
        nfd-worker-cert:
          Type:        Secret (a volume populated by a Secret)
          SecretName:  nfd-worker-cert
          Optional:    false
        kube-api-access-rmsmr:
          Type:                    Projected (a volume that contains injected data from multiple sources)
          TokenExpirationSeconds:  3607
          ConfigMapName:           kube-root-ca.crt
          ConfigMapOptional:       <nil>
          DownwardAPI:             true
      QoS Class:                   BestEffort
      Node-Selectors:              <none>
      Tolerations:                 node.kubernetes.io/disk-pressure:NoSchedule op=Exists
                                   node.kubernetes.io/memory-pressure:NoSchedule op=Exists
                                   node.kubernetes.io/not-ready:NoExecute op=Exists
                                   node.kubernetes.io/pid-pressure:NoSchedule op=Exists
                                   node.kubernetes.io/unreachable:NoExecute op=Exists
                                   node.kubernetes.io/unschedulable:NoSchedule op=Exists
      Events:
        Type     Reason          Age                  From               Message
        ----     ------          ----                 ----               -------
        Normal   Scheduled       3m7s                 default-scheduler  Successfully assigned node-feature-discovery/nfd-node-feature-discovery-worker-r2t92 to 192.114.9.100
        Warning  FailedMount     3m4s (x4 over 3m7s)  kubelet            MountVolume.SetUp failed for volume "nfd-worker-cert" : secret "nfd-worker-cert" not found
        Normal   AddedInterface  2m59s                multus             Add eth0 [10.42.0.19/32] from k8s-pod-network
        Normal   Pulled          2m59s                kubelet            Container image "k8s.gcr.io/nfd/node-feature-discovery:v0.11.0" already present on machine
        Normal   Created         2m59s                kubelet            Created container worker
        Normal   Started         2m59s                kubelet            Started container worker



Copyright (c) 2022 Intel Corporation
SPDX-License-Identifier: Apache-2.0
