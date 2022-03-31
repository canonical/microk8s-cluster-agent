package snap_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

var (
	caCrt = `-----BEGIN CERTIFICATE-----
MIIDDzCCAfegAwIBAgIURKOFladyi/c/Srb+1LGiHpVkrb8wDQYJKoZIhvcNAQEL
BQAwFzEVMBMGA1UEAwwMMTAuMTUyLjE4My4xMB4XDTIyMDMzMTIwMTA0NVoXDTMy
MDMyODIwMTA0NVowFzEVMBMGA1UEAwwMMTAuMTUyLjE4My4xMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEArEMz1mw0QmeKo1IH3+uJWl8UxsnsXJdtQNzs
0apy8wsCpJMCYLsElB6PDjYdQOXyBO7zPdIVAGmHsD+X8nSBHPl8YdAaIk1cSCxy
N5Kee39v2P/luCIY4Gfc8465tQmM7tnBVMvBf1+jCC0s5I6SQz8Z4VlbKf/0kvFy
+n4UyQxDzQ/PwpUEPWWSwey95ULutsRx4X6Xa0HBxCExXXIPVOGAnbQQi3bPQ7ch
lEpGPDNo8YUaTHTLjphY3vDG/NVHFbnxhwxkt/O3/rFnkG2XGTqILZp135ADLdGo
ARRpvMQEahQGPfrq+OdL+KiVQtXX7hZQQMOyn13PvbloBrxK5wIDAQABo1MwUTAd
BgNVHQ4EFgQUHbhQiEEpnlosos2880pu0PKgttUwHwYDVR0jBBgwFoAUHbhQiEEp
nlosos2880pu0PKgttUwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOC
AQEAYAu1Vn9prSEAp3BOpT0R9QirZc+tYQAzc3et4ACS5uYM2vlcnh0qZQ3xg997
+rH0ImIbEzFYEzPywRbsVsDGuPZEY9PSgJcgyGNiCxEKr3qY62x1l5ngFxuleCZj
nQK+Yu1rjVlt7E/7TljdVR1a37hRoA4U3LuGKvF7iJt3KOPXdoEGkA41BCgddkUX
TLNGE4/gOkSSB7IEje3BrxsDu2cbiPIwA2+KgqS5jaOnyE/4hy3sdW6Qar/NRVSQ
DpIQiRyiIZich8nWTJrF3SMXdCRTbzczfNYRb17YYWC21cEJOvtqbZtCbu9vctGY
LnQGADk0liOWnYVWtUTG8DP6Kw==
-----END CERTIFICATE-----
`
	caKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEArEMz1mw0QmeKo1IH3+uJWl8UxsnsXJdtQNzs0apy8wsCpJMC
YLsElB6PDjYdQOXyBO7zPdIVAGmHsD+X8nSBHPl8YdAaIk1cSCxyN5Kee39v2P/l
uCIY4Gfc8465tQmM7tnBVMvBf1+jCC0s5I6SQz8Z4VlbKf/0kvFy+n4UyQxDzQ/P
wpUEPWWSwey95ULutsRx4X6Xa0HBxCExXXIPVOGAnbQQi3bPQ7chlEpGPDNo8YUa
THTLjphY3vDG/NVHFbnxhwxkt/O3/rFnkG2XGTqILZp135ADLdGoARRpvMQEahQG
Pfrq+OdL+KiVQtXX7hZQQMOyn13PvbloBrxK5wIDAQABAoIBAQCI/wkCxhrplJ7V
NC+/DPq3PFaxoXSwTl088HYGaJ/yWhdH+wIxG4qQoZzzmGW7byQCeGZKDAFXypV2
vZyB3dWwzVj2ESI3lX6Gh5JuT9RqMvcohJqUzckq7m7x49gc+fNzIQp3/XNtcQmf
Pw4s1pXLlStn5EB0LnK7Sfo+7HZrTF86p4GxJqhB3bzo/CTxGjkb+3s9Q+bHOfvu
3iwR9f7hJtGH5M4gCACMcAE7MVuA0+ffLwEBdxyBuSmREU3hjIhiaUyZXh6vlwdJ
G3seLUFDuIUyaB3gXnNls1IkzM+X01sJ9G7K+vK4VVniJaIOVfvUrLcLli2vJlhZ
zGcahj+BAoGBAN3Rdu27gHsBQQZT/u2nSlANVc3PXhrrdMhzyifyKU5o0zE2mdoh
oDvfL2fxAO32fy8H3JVoAIwN3qYtkS/0I4rLCRxQDl9KmfvKFKt8Tr0YGRxS1NDu
gMcwo2nj349mxuloypDR0Ww2fVJWHQeALwkAP3jvMCeKOj6JntwqaSe/AoGBAMbO
zu4WcUNKBfvfj9Ugx6YO08mNtea8EubNXOqcqqFzq6tRTWwav6ofQJNFcJsZvKvF
eV2OjpW/uZpwA3NrbZcf+FA7JOH0urfX3W5Y2AHiCk/c6Di5R7hzqHyhUVvTcJOI
YdE3TZ4ZkUwnmMZxJ+DVg3r59ulSkyoYdOr/HObZAoGAfKVG+kIh6X0D6CVtHGik
NqW3sKY1UFU9U6LVV2sZ0QjQnFf9TnkUzHAW+IaSKiYYw/nb90zw+cKVebYjXtoG
2uhK31ERMnT+YGHnCZIZwOJ4wdS96AYN8WCgg1Fcf/2WCvUq1wRAdVmNRKZFO8DJ
LXqpMDDgU2e2YQv+a+OdIYsCgYB4+tavydZo1T3o9TWow4baxYEZ0POVknOKIgRd
/LJVB3e7DAGqPGjQFK6OMB6DM9k9SjE9voeEFyTSF0HyVbhd06We5S7flbaeM21b
PhNMqgoOaWajRhSf6TnphZ1l3LhP/xlPYHEKOZLSnfH5KFjVF/knt78KYyaM4k8b
xd0HmQKBgC+oTh6STV7vzGjtRH53SAbUbJyEzBesjIVrTxyRics8Hu1JGBjnKK6E
Qj+w5YAO9H8vYrySeeNMItOV8fkOE0mCBzpmonjeWCEKKj7rcPtUgIi8CLbV5wLw
+HOoxP4vxs6qLpRnuoN7q78+byFXKvSAfvQ4ox1GE3cN9oAGI2xs
-----END RSA PRIVATE KEY-----
`
	csr = `-----BEGIN CERTIFICATE REQUEST-----
MIICdTCCAV0CAQAwMDEXMBUGA1UEAwwOc3lzdGVtOm5vZGU6YTIxFTATBgNVBAoM
DHN5c3RlbTpub2RlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMCP
SEMIC9W+S/6Td1erRoAfEMMZfehLdNmE8m6wUgNU7AACo1f8AVGFyULFaHmc19PY
wDh9o4+DyIQALfvRVTqbTt1eJwLqGNvvw+YkVuKiBw/VgeNz9W34mel03NtqJniH
drXEHu/P4136bQxQWZY/bqKpPpBTiljBLENihd7Y+/S3fmuo1PM/PCr43QlB9gWD
C0X1Qqzt9BUN0alYLwKChdnbPsfZcauosE82x9bjkRQopsTv7pnjITa3ure+I+jn
34msNOuUqgkwGIxH+qvTulcqV0a1RSX0QZxHILq1NEwTFzozXCLC0WVdKqEoPFrO
tS8lD0O5kAwYyEiTNb0CAwEAAaAAMA0GCSqGSIb3DQEBCwUAA4IBAQB0SyO5s5TC
Jy4cbDm/LqoeU/jj1dnMuGgRqB0aWd8oRpfREQgfVVgcN2t3uJtOh7jQkLhpqJrt
xoeuDX2FUpas2VggEifO3pF67yJdJx9bTS/5Uk5eQDM+CeFA8oLauIj0By8MsD2O
mS1O6TxS5zqrqh5EbGvE3C3HaSjxSqIGoWys+WNl+NJRaYRI12wpvoABd6VYeCc6
NVqEX5fgKvF0tTYv21F/l6TD83V/5YG5N9ffjRs3VUBYu4JXoI5VI7DGh9dEFfRK
Jj9HwRBP2cYf6w9AcuLgR3UBac1ICZsrN7ldjAR7vWE5yVbQXkU3YLBWV/PGyGrT
ZVWi2dITDvqF
-----END CERTIFICATE REQUEST-----
`
)

func TestSignCert(t *testing.T) {
	if err := os.MkdirAll("testdata/certs", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %s", err)
	}
	defer os.RemoveAll("testdata/certs")

	if err := os.WriteFile("testdata/certs/ca.crt", []byte(caCrt), 0600); err != nil {
		t.Fatalf("Failed to write test certificate: %s", err)
	}
	if err := os.WriteFile("testdata/certs/ca.key", []byte(caKey), 0600); err != nil {
		t.Fatalf("Failed to write test certificate key: %s", err)
	}

	s := snap.NewSnap("testdata", "testdata", nil)

	certificate, err := s.SignCertificate(context.Background(), []byte(csr))
	if err != nil {
		t.Fatalf("Expected no errors when signing certificate, but received %s", err)
	}
	if !strings.HasPrefix(string(certificate), "-----BEGIN CERTIFICATE-----") {
		t.Fatalf("Expected certificate to not be empty")
	}
}
