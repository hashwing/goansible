package ftp

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// Client ftp client
type Client struct {
	Config
}

type Config struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
	User string `yaml:"user" json:"user"`
	PWD  string `yaml:"password" json:"password"`
	Dir  string `yaml:"dir" json:"dir"`
	Tmp  string `yaml:"tmp" json:"tmp_dir"`
}

// NewClient new client
func NewClient(cfg *Config) (client *Client) {
	client = &Client{
		*cfg,
	}
	return
}

// ftpLogin login ftp
func (c *Client) ftpLogin() (*ftp.ServerConn, error) {
	opts := []ftp.DialOption{
		ftp.DialWithDisabledEPSV(true),
		ftp.DialWithTimeout(5 * time.Second),
	}
	ftpClient, err := ftp.Dial(c.Host+":"+c.Port, opts...)
	if err != nil {
		return nil, err
	}
	err = ftpClient.Login(c.User, c.PWD)
	if err != nil {
		ftpClient.Quit()
		return nil, err
	}
	return ftpClient, nil
}

// Upload upload file
func (c *Client) Upload(r io.Reader, path string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()
	ftpClient.MakeDir(strings.Replace(filepath.Dir(path), "\\", "/", -1))
	err = ftpClient.Stor(path, r)
	if err != nil {
		return err
	}
	return nil
}

// FUpload upload file
func (c *Client) FUpload(local, path string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()

	fd, err := os.Open(local)
	if err != nil {
		return err
	}
	defer fd.Close()
	ftpClient.MakeDir(filepath.Dir(path))
	//c.MkDir(filepath.Dir(path))
	err = ftpClient.Stor(path, fd)
	if err != nil {
		return err
	}
	return nil
}

// Download dowload file
func (c *Client) Download(path string) (*ftp.Response, error) {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return nil, err
	}
	r, err := ftpClient.Retr(path)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// FDownload dowload file
func (c *Client) FDownload(remote, local string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()
	r, err := ftpClient.Retr(remote)
	if err != nil {
		return err
	}
	defer r.Close()
	fd, err := os.OpenFile(local, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer fd.Close()
	for {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		if n == 0 {
			break
		}
		fd.Write(buf[:n])
	}

	fileInfo, err := os.Stat(local)
	if err != nil {
		return err
	}
	size, err := c.Size(remote)
	if err != nil {
		return err
	}

	if size != fileInfo.Size() {
		return errors.New("[FTPDownload] download size error")
	}
	return nil
}

// Size get file size
func (c *Client) Size(path string) (int64, error) {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return 0, err
	}
	defer ftpClient.Quit()
	return ftpClient.FileSize(path)
}

//Rename rename file
func (c *Client) Rename(from, to string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()
	ftpClient.MakeDir(filepath.Dir(to))
	err = ftpClient.Rename(from, to)
	return err
}

//Rename rename file
func (c *Client) TranFinish(path string) (bool, error) {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return false, err
	}
	defer ftpClient.Quit()
	err = ftpClient.Rename(path, path+".ferry")
	if err != nil {
		return false, nil
	}
	ftpClient.Rename(path+".ferry", path)
	return true, err
}

//Rename rename file
func (c *Client) TranFinishs(paths []string) ([]string, error) {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return nil, err
	}
	defer ftpClient.Quit()
	res := make([]string, 0)
	for _, path := range paths {
		err = ftpClient.Rename(path, path+".ferry")
		if err != nil {
			continue
		}
		ftpClient.Rename(path+".ferry", path)
		res = append(res, path)
	}
	return res, err
}

//Delete delete file
func (c *Client) Delete(path string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()
	err = ftpClient.Delete(path)
	return err
}

//MkDir make dir
func (c *Client) MkDir(path string) error {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return err
	}

	defer ftpClient.Quit()
	err = ftpClient.MakeDir(path)
	return err
}

type Entry struct {
	Name   string
	Target string // target of symbolic link
	Type   EntryType
	Size   uint64
	Time   int64
	Path   string
}

type EntryType int

// The differents types of an Entry
const (
	EntryTypeFile EntryType = iota
	EntryTypeFolder
	EntryTypeLink
)

// FindList find list
func (c *Client) FindList(path string) (fileList []Entry, err error) {
	fileList = make([]Entry, 0)
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return
	}
	defer ftpClient.Quit()
	c.fileWalk(path, ftpClient, &fileList)
	return
}

// fileWalk walk all file
func (c *Client) fileWalk(path string, ftpClient *ftp.ServerConn, fileQueue *[]Entry) error {
	list, err := ftpClient.List(path)
	if err != nil {
		return err
	}
	if len(list) != 0 {
		for _, fileList := range list {
			if fileList.Name == "." || fileList.Name == ".." {
				continue
			}
			if fileList.Type == ftp.EntryTypeFile {
				if strings.HasSuffix(fileList.Name, ".unimastmp") || strings.HasSuffix(fileList.Name, ".unimastest") {
					continue
				}
				entry := Entry{
					Name:   fileList.Name,
					Path:   strings.TrimSuffix(path, "/") + "/" + fileList.Name,
					Target: fileList.Target,
					Size:   fileList.Size,
					Time:   fileList.Time.Unix(),
					Type:   EntryType(fileList.Type),
				}
				(*fileQueue) = append((*fileQueue), entry)
			}
			if fileList.Type == ftp.EntryTypeFolder {
				err = c.fileWalk(strings.TrimSuffix(path, "/")+"/"+fileList.Name, ftpClient, fileQueue)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CleanEmptyDir clean empty dir
func (c *Client) CleanEmptyDir(path string) {
	ftpClient, err := c.ftpLogin()
	if err != nil {
		return
	}
	defer ftpClient.Quit()
	for {
		err := delEmptyDir(path, path, ftpClient)
		if err != nil {
			return
		}
		time.Sleep(time.Second * 5)
	}

}

// delEmptyDir walk and dir empty dir
func delEmptyDir(path, base string, ftpClient *ftp.ServerConn) error {
	list, err := ftpClient.List(path)
	if err != nil {
		return errors.New("reconn")
	}
	if len(list) != 0 {
		for _, fileList := range list {
			if fileList.Type == ftp.EntryTypeFolder {
				err = delEmptyDir(path+"/"+fileList.Name+"/", base, ftpClient)
				if err != nil {
					return err
				}
			}
		}
	} else {
		if path != base {
			err = ftpClient.RemoveDir(path)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}
