package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	rd "math/rand"
	"net"
	"regexp"
	"strconv"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

func init() {
	rd.Seed(time.Now().UnixNano())
}

//CertInformation cert info
type CertInformation struct {
	Country            []string  `json:"country,omitempty"`
	Organization       []string  `json:"organization,omitempty"`
	OrganizationalUnit []string  `json:"organizationalUnit,omitempty"`
	EmailAddress       []string  `json:"emailAddress,omitempty"`
	Province           []string  `json:"province,omitempty"`
	Locality           []string  `json:"locality,omitempty"`
	CommonName         string    `json:"commonName,omitempty"`
	IsCA               bool      `json:"isCA"`
	IPAddresses        []string  `json:"ipAddresses,omitempty"`
	Domains            []string  `json:"domains,omitempty"`
	Expires            string    `json:"expires,omitempty"`
	NotBefore          time.Time `json:"notBefore,omitempty"`
	NotAfter           time.Time `json:"notAfter,omitempty"`
	Password           string    `json:"password,omitempty"` //暂未实现加密证书
}

//CreateCRT create cert
func CreateCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info *CertInformation) (crtB []byte, keyB []byte, err error) {
	Crt := newCertificate(info)
	Key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	if RootCa == nil || RootKey == nil {
		//创建自签名证书
		crtB, err = x509.CreateCertificate(rand.Reader, Crt, Crt, &Key.PublicKey, Key)
	} else {
		//使用根证书签名
		crtB, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)
	}
	if err != nil {
		return
	}
	crtB = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Headers: map[string]string{}, Bytes: crtB})

	keyB = x509.MarshalPKCS1PrivateKey(Key)
	keyB = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Headers: map[string]string{}, Bytes: keyB})
	return
}

//Parse ...
func Parse(crtB, KeyB []byte) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootcertificate, err = ParseCrt(crtB)
	if err != nil {
		return
	}
	rootPrivateKey, err = ParseKey(KeyB)
	return
}

//ParseCrt ...
func ParseCrt(buf []byte) (*x509.Certificate, error) {
	p := &pem.Block{}
	p, buf = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

//ParseKey ...
func ParseKey(buf []byte) (*rsa.PrivateKey, error) {
	p, buf := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func newCertificate(info *CertInformation) *x509.Certificate {

	if len(info.Country) == 0 {
		info.Country = []string{"CN"}
	}
	if len(info.Organization) == 0 {
		info.Organization = []string{"GoAnsible"}
	}
	if len(info.OrganizationalUnit) == 0 {
		info.OrganizationalUnit = []string{"GoAnsible"}
	}
	if len(info.Province) == 0 {
		info.Province = []string{"Guangdong"}
	}
	if len(info.Locality) == 0 {
		info.Locality = []string{"Guangzhou"}
	}
	if len(info.EmailAddress) == 0 {
		info.EmailAddress = []string{"goansible@hashwing.cn"}
	}
	notAfter, err := parseExpiry(info.Expires)
	if err != nil {
		notAfter = time.Now().AddDate(1, 0, 0)
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(rd.Int63()),
		Subject: pkix.Name{
			Country:            info.Country,
			Organization:       info.Organization,
			OrganizationalUnit: info.OrganizationalUnit,
			Province:           info.Province,
			CommonName:         info.CommonName,
			Locality:           info.Locality,
		},
		NotBefore:             time.Now(),                                                                 //证书的开始时间
		NotAfter:              notAfter,                                                                   //证书的结束时间
		BasicConstraintsValid: true,                                                                       //基本的有效性约束
		IsCA:                  info.IsCA,                                                                  //是否是根证书
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, //证书用途
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses:        info.EmailAddress,
	}
	for _, addr := range info.IPAddresses {
		cert.IPAddresses = append(cert.IPAddresses, net.ParseIP(addr))
	}
	for _, domain := range info.Domains {
		cert.DNSNames = append(cert.DNSNames, domain)
	}
	info.NotBefore = cert.NotBefore
	info.NotAfter = cert.NotAfter
	return cert
}

//ParseCertToInfo ...
func ParseCertToInfo(cert *x509.Certificate) *CertInformation {
	ips := make([]string, 0)
	for _, ipAddr := range cert.IPAddresses {
		ips = append(ips, ipAddr.String())
	}
	return &CertInformation{
		Country:            cert.Subject.Country,
		Organization:       cert.Subject.Organization,
		OrganizationalUnit: cert.Subject.OrganizationalUnit,
		Province:           cert.Subject.Province,
		CommonName:         cert.Subject.CommonName,
		Locality:           cert.Subject.Locality,
		NotAfter:           cert.NotAfter,
		NotBefore:          cert.NotBefore,
		IPAddresses:        ips,
		Domains:            cert.DNSNames,
		IsCA:               cert.IsCA,
	}
}

//CertToP12 ...
func CertToP12(certBuf, keyBuf []byte, certPwd string) (p12Cert string, err error) {
	caBlock, _ := pem.Decode(certBuf)
	crt, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return
	}

	keyBlock, _ := pem.Decode(keyBuf)
	priKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return
	}

	pfx, err := pkcs12.Encode(rand.Reader, priKey, crt, nil, certPwd)
	if err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(pfx), err
}

var nowFunc = time.Now

func parseExpiry(fromNow string) (time.Time, error) {
	now := nowFunc().UTC()
	re := regexp.MustCompile(`\s*(\d+)\s*(d|m|y|h|m|s)?`)
	matches := re.FindAllStringSubmatch(fromNow, -1)
	addDate := map[string]int{
		"d": 0,
		"M": 0,
		"y": 0,
		"h": 0,
		"m": 0,
		"s": 0,
	}
	for _, r := range matches {
		number, err := strconv.ParseInt(r[1], 10, 32)
		if err != nil {
			return now, err
		}
		addDate[r[2]] = int(number)
	}

	// Ensure that we do not overflow time.Duration.
	// Doing so is silent and causes signed integer overflow like issues.
	if _, err := time.ParseDuration(fmt.Sprintf("%dh", addDate["h"])); err != nil {
		return now, fmt.Errorf("hour unit too large to process")
	} else if _, err = time.ParseDuration(fmt.Sprintf("%dm", addDate["m"])); err != nil {
		return now, fmt.Errorf("minute unit too large to process")
	} else if _, err = time.ParseDuration(fmt.Sprintf("%ds", addDate["s"])); err != nil {
		return now, fmt.Errorf("second unit too large to process")
	}

	result := now.
		AddDate(addDate["y"], addDate["M"], addDate["d"]).
		Add(time.Duration(addDate["h"]) * time.Hour).
		Add(time.Duration(addDate["m"]) * time.Minute).
		Add(time.Duration(addDate["s"]) * time.Second)

	if now == result {
		return now, fmt.Errorf("invalid or empty format")
	}

	// ASN.1 (encoding format used by SSL) only supports up to year 9999
	// https://www.openssl.org/docs/man1.1.0/crypto/ASN1_TIME_check.html
	if result.Year() > 9999 {
		return now, fmt.Errorf("proposed date too far in to the future: %s. Expiry year must be less than or equal to 9999", result)
	}

	return result, nil
}
