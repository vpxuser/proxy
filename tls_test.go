package proxy

import (
	"testing"
)

func TestTLS(t *testing.T) {
	tlsCfg, err := FromCA(Certificate, PrivateKey)("www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("from ca: \n%v", tlsCfg)

	tlsCfg, err = FromSelfSigned()("www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("from self sign: \n%v", tlsCfg)

	tlsCfg, err = From(Certificate, PrivateKey)("www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("from: \n%v", tlsCfg)
}
