package actions

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/utils/cert"
)

type CertAction struct {
	RootCrtPath        string `yaml:"root_crt_path"`
	RootKeyPath        string `yaml:"root_key_path"`
	CrtPath            string `yaml:"crt_path"`
	KeyPath            string `yaml:"key_path"`
	Country            string `yaml:"country,omitempty"`
	Organization       string `yaml:"organization,omitempty"`
	OrganizationalUnit string `yaml:"organizational_unit,omitempty"`
	EmailAddress       string `yaml:"email_address,omitempty"`
	Province           string `yaml:"province,omitempty"`
	Locality           string `yaml:"locality,omitempty"`
	CommonName         string `yaml:"common_name,omitempty"`
	IsCA               string `yaml:"is_ca"`
	IPAddresses        string `yaml:"ip_addresses,omitempty"`
	Domains            string `yaml:"domains,omitempty"`
	Expires            string `yaml:"expires,omitempty"`
	P12                string `yaml:"p12,omitempty"`
	P12Password        string `yaml:"p12_password,omitempty"`
}

func (a *CertAction) parse(vars *model.Vars) (*CertAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &CertAction{
		CrtPath:            common.ParseTplWithPanic(a.CrtPath, vars),
		KeyPath:            common.ParseTplWithPanic(a.KeyPath, vars),
		RootCrtPath:        common.ParseTplWithPanic(a.RootCrtPath, vars),
		RootKeyPath:        common.ParseTplWithPanic(a.RootKeyPath, vars),
		CommonName:         common.ParseTplWithPanic(a.CommonName, vars),
		Domains:            common.ParseTplWithPanic(a.Domains, vars),
		IPAddresses:        common.ParseTplWithPanic(a.IPAddresses, vars),
		Organization:       common.ParseTplWithPanic(a.Organization, vars),
		OrganizationalUnit: common.ParseTplWithPanic(a.OrganizationalUnit, vars),
		EmailAddress:       common.ParseTplWithPanic(a.EmailAddress, vars),
		Country:            common.ParseTplWithPanic(a.Country, vars),
		Province:           common.ParseTplWithPanic(a.Province, vars),
		Locality:           common.ParseTplWithPanic(a.Locality, vars),
		IsCA:               common.ParseTplWithPanic(a.IsCA, vars),
		Expires:            common.ParseTplWithPanic(a.Expires, vars),
		P12:                common.ParseTplWithPanic(a.P12, vars),
		P12Password:        common.ParseTplWithPanic(a.P12Password, vars),
	}, gerr
}

func (a *CertAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	info := &cert.CertInformation{
		CommonName:         parseAction.CommonName,
		Domains:            common.ParseArray(parseAction.Domains),
		IPAddresses:        common.ParseArray(parseAction.IPAddresses),
		Organization:       common.ParseArray(parseAction.Organization),
		OrganizationalUnit: common.ParseArray(parseAction.OrganizationalUnit),
		EmailAddress:       common.ParseArray(parseAction.EmailAddress),
		Country:            common.ParseArray(parseAction.Country),
		Province:           common.ParseArray(parseAction.Province),
		Locality:           common.ParseArray(parseAction.Locality),
		IsCA:               parseAction.IsCA == "true",
		Expires:            parseAction.Expires,
	}

	var crtB []byte
	var keyB []byte
	if info.IsCA {
		crtB, keyB, err = cert.CreateCRT(nil, nil, info)
	} else {
		rcpath := parseAction.RootCrtPath
		if !filepath.IsAbs(rcpath) {
			rcpath = filepath.Join(conf.PlaybookFolder, parseAction.RootCrtPath)
		}
		rootCrtB, err := ioutil.ReadFile(rcpath)
		if err != nil {
			return "", err
		}
		rkpath := parseAction.RootKeyPath
		if !filepath.IsAbs(rkpath) {
			rkpath = filepath.Join(conf.PlaybookFolder, parseAction.RootKeyPath)
		}
		rootKeyB, err := ioutil.ReadFile(rkpath)
		if err != nil {
			return "", err
		}

		rootCrt, err := cert.ParseCrt(rootCrtB)
		if err != nil {
			return "", err
		}
		rootKey, err := cert.ParseKey(rootKeyB)
		if err != nil {
			return "", err
		}
		fmt.Println(cert.ParseCertToInfo(rootCrt))
		crtB, keyB, err = cert.CreateCRT(rootCrt, rootKey, info)
	}
	if err != nil {
		return "", err
	}
	cpath := parseAction.CrtPath
	if !filepath.IsAbs(cpath) {
		cpath = filepath.Join(conf.PlaybookFolder, parseAction.CrtPath)
	}
	os.MkdirAll(filepath.Dir(cpath), 0755)
	err = ioutil.WriteFile(cpath, crtB, 0664)
	if err != nil {
		return "", err
	}
	kpath := parseAction.KeyPath
	if !filepath.IsAbs(kpath) {
		kpath = filepath.Join(conf.PlaybookFolder, parseAction.KeyPath)
	}
	os.MkdirAll(filepath.Dir(kpath), 0755)
	err = ioutil.WriteFile(kpath, keyB, 0664)
	if err != nil {
		return "", err
	}
	if parseAction.P12 != "" {
		p12cert, err := cert.CertToP12(crtB, keyB, parseAction.P12Password)
		if err != nil {
			return "", err
		}
		p12path := parseAction.P12
		if !filepath.IsAbs(p12path) {
			p12path = filepath.Join(conf.PlaybookFolder, parseAction.P12)
		}
		os.MkdirAll(filepath.Dir(p12path), 0755)
		err = ioutil.WriteFile(p12path, []byte(p12cert), 0664)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}
