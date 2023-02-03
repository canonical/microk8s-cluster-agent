package util

import (
	"bytes"
	"fmt"
	"net"
	"text/template"
)

var csrConfTemplate = template.Must(template.New("csrConf").Parse(`
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = GB
ST = Canonical
L = Canonical
O = Canonical
OU = Canonical
CN = 127.0.0.1

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
{{ range $i, $a := .DNSNames }}DNS.{{ $i }} = {{ $a }}
{{ end }}

{{ range $i, $a := .IPAddresses }}IP.{{ $i }} = {{ $a }}
{{ end }}
#MOREIPS

[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment,digitalSignature
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
`))

type templateData struct {
	IPAddresses []string
	DNSNames    []string
}

// GenerateCSRConf generates a csr.conf.template file for the MicroK8s node.
// extraSANs are a list of extra Subject Alternate Names to be added to the certificates.
func GenerateCSRConf(extraSANs []string) ([]byte, error) {
	ips := []string{"127.0.0.1", "10.152.183.1"}
	dns := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local"}
	for _, v := range extraSANs {
		if net.ParseIP(v) == nil {
			dns = append(dns, v)
		} else {
			ips = append(ips, v)
		}
	}
	var b bytes.Buffer
	if err := csrConfTemplate.Execute(&b, templateData{IPAddresses: ips, DNSNames: dns}); err != nil {
		return nil, fmt.Errorf("failed to render csr.conf.template: %w", err)
	}

	return b.Bytes(), nil
}
