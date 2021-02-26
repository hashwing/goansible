package actions

import (
	"context"
	"fmt"
	"io/ioutil"
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
		fmt.Println("dddd", parseAction.RootCrtPath)
		rootCrtB, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, parseAction.RootCrtPath))
		if err != nil {
			return "", err
		}
		rootKeyB, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, parseAction.RootKeyPath))
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
	err = ioutil.WriteFile(filepath.Join(conf.PlaybookFolder, parseAction.CrtPath), crtB, 0664)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filepath.Join(conf.PlaybookFolder, parseAction.KeyPath), keyB, 0664)
	if err != nil {
		return "", err
	}
	return "", nil
}
