package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func NewRequestError(msg string, err ...error) error {
	if len(err) > 0 {
		msg += ":" + err[0].Error()
	}
	return errors.New(msg)
}

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	ModeList string
}

type Client struct {
	connectionOptions
	client    *gossh.Client
	connected bool
	session   *gossh.Session
	in        io.WriteCloser
	out       io.Reader
}

func Connect(user, password, address string, port int, opts ...ConnectionOpt) (*Client, error) {
	client := &Client{}

	if port == 0 {
		port = 22
	}

	client.user = user
	client.password = password

	address += ":" + strconv.Itoa(port)
	client.connectionOptions.ClientConfig = &gossh.ClientConfig{}
	client.connectionOptions.ClientConfig.SetDefaults()
	//aes128-cbc aes256-cbc 3des-cbc des-cbc
	client.connectionOptions.ClientConfig.Ciphers = append(client.connectionOptions.ClientConfig.Ciphers, "3des-cbc")
	client.connectionOptions.ClientConfig.KeyExchanges = append(client.connectionOptions.ClientConfig.KeyExchanges, "diffie-hellman-group1-sha1")
	client.connectionOptions.ClientConfig.HostKeyCallback = gossh.InsecureIgnoreHostKey()
	client.connectionOptions.ClientConfig.Timeout = 5 * time.Second
	client.Address = address
	client.NetworkType = "tcp"

	for _, v := range opts {
		v(&client.connectionOptions)
	}

	client.connectionOptions.ClientConfig.User = client.user

	if client.password != "" {
		if client.authType == "keyboardInteractive" {
			keyboardInteractiveChallenge := func(
				user,
				instruction string,
				questions []string,
				echos []bool,
			) (answers []string, err error) {
				if len(questions) == 0 {
					return []string{}, nil
				}
				return []string{client.password}, nil
			}
			client.connectionOptions.ClientConfig.Auth = []gossh.AuthMethod{gossh.KeyboardInteractive(keyboardInteractiveChallenge), gossh.Password(client.password)}
		} else {
			client.connectionOptions.ClientConfig.Auth = []gossh.AuthMethod{gossh.Password(client.password)}
		}
	}

	err := client.Conn()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Conn() error {
	if c.connected {
		return nil
	}
	var err error
	c.client, err = gossh.Dial(c.NetworkType, c.Address, c.ClientConfig)
	if err != nil {
		return NewRequestError("连接ssh失败", err)
	}

	c.connected = true

	return nil
}

func (c *Client) Close() error {
	if c.session != nil {
		c.session.Close()
	}
	if err := c.client.Close(); err != nil {
		return NewRequestError("关闭ssh失败", err)
	}
	return nil
}

func (c *Client) encode(cmd *string) error {
	switch c.encoding {
	case "", "utf8":
	case "gbk":
		d, encodingErr := Utf8ToGbk([]byte(*cmd))
		if encodingErr != nil {
			return NewRequestError("utf8转gbk编码错误", encodingErr)
		}
		*cmd = string(d)
	default:
		return NewRequestError(fmt.Sprintf("不支持的编码:%s", c.encoding))
	}

	return nil
}

func (c *Client) decode(cmd *string) error {
	switch c.encoding {
	case "", "utf8":
	case "gbk":
		d, encodingErr := GbkToUtf8([]byte(*cmd))
		if encodingErr != nil {
			return NewRequestError("gbk转utf8编码错误", encodingErr)
		}
		*cmd = string(d)
	default:
		return NewRequestError(fmt.Sprintf("不支持的编码:%s", c.encoding))
	}

	return nil
}

func (c *Client) Exec(cmd string, opts ...ExecOpt) (string, error) {
	options := execOptions{}
	for _, v := range opts {
		v(&options)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session, err := c.client.NewSession()
	if nil != err {
		return "", NewRequestError("ssh连接失败", err)
	}

	defer func() {
		if err := session.Close(); err != nil {
			if err.Error() != "EOF" {
				log.WithError(NewRequestError("关闭session失败", err)).Error()
			}
		}
	}()

	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = c.encode(&cmd)
	if err != nil {
		return "", err
	}

	err = session.Run(cmd)
	if err != nil {
		stdout := stdoutBuf.String()
		stderr := stderrBuf.String()
		return "", NewRequestError(fmt.Sprintf("ssh操作失败,%s,%s", stdout, stderr), err)
	}

	out := stdoutBuf.String()

	return c.sanitizeOutPut(&out, cmd, &options)
}

func (c *Client) GetPrompt() string {
	return c.prompt
}

func (c *Client) termSession(opts ...ExecOpt) error {
	var err error
	options := execOptions{}
	for _, v := range opts {
		v(&options)
	}

	for i := 0; i < 3; i++ {
		c.session, err = c.client.NewSession()
		if err != nil && strings.Contains(err.Error(), "prohibited") {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
	//c.session, err = c.client.NewSession()
	if err != nil {
		return NewRequestError("ssh连接失败", err)
	}

	c.session.Stderr = os.Stderr

	modes := gossh.TerminalModes{
		gossh.ECHO:          0,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}

	if err := c.session.RequestPty("xterm", 768, 1024, modes); err != nil {
		return NewRequestError("ssh连接失败", err)
	}

	c.in, err = c.session.StdinPipe()
	c.out, _ = c.session.StdoutPipe()
	if err != nil {
		return NewRequestError("ssh连接失败", err)
	}

	err = c.session.Shell()
	if err != nil {
		return NewRequestError("ssh连接失败", err)
	}

	buff := make([]byte, 1024)
	var out string

	for {
		n, err := c.out.Read(buff)
		if err != nil {
			return NewRequestError("ssh连接失败", err)
		} else {
			out += string(buff[:n])
			if options.execFun != nil {
				cmdFinish := CmdStart
				if err := options.execFun(&cmdFinish, &out, c.in); err != nil {
					return NewRequestError("ssh连接失败", err)
				}
			}
			if c.prompt != "" {
				if strings.Contains(string(buff[:n]), c.prompt) {
					break
				}
			} else {
				if strings.Contains(string(buff[:n]), c.promptEnd) {
					break
				}
			}
		}
	}

	_, err = c.in.Write([]byte("\nfind_prompt\n"))
	if err != nil {
		return NewRequestError("ssh连接失败", err)
	}

	if c.prompt != "" {
		return nil
	}

	var clearTmp string

	for {
		n, err := c.out.Read(buff)
		if err != nil {
			return NewRequestError("ssh连接失败", err)
		} else {
			c.prompt += string(buff[:n])
			if strings.Contains(c.prompt, "find_prompt") {
				c.prompt, clearTmp = parseErrorStr(c.prompt, "find_prompt")
				break
			} else if strings.Contains(c.prompt, "# error -") {
				c.prompt, clearTmp = parseErrorStr(c.prompt, " error -")
				break
			}
		}
	}

	//clear the rest unread data
	for {
		if strings.Contains(clearTmp, c.prompt) {
			break
		}
		n, err := c.out.Read(buff)
		if err != nil {
			return NewRequestError("ssh连接失败", err)
		} else {
			clearTmp += string(buff[:n])
		}
	}

	return nil
}

func parseErrorStr(prompt, errStr string) (resPrompt, resClearTmp string) {
	tmpStr := prompt
	index := strings.Index(prompt, errStr)
	if index != -1 {
		resPrompt = strings.TrimSpace(tmpStr[:index])
		resClearTmp = tmpStr[index:]
	}

	return
}

func (c *Client) ConnectTermSession(opts ...ExecOpt) error {
	if c.session == nil {
		err := c.termSession(opts...)
		if err != nil {
			return NewRequestError("ssh连接失败", err)
		}
	}

	return nil
}

func (c *Client) CloseTermSession() error {
	if c.session == nil {
		return nil
	}

	if err := c.session.Close(); err != nil {
		return err
	}

	c.session = nil
	return nil
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func (c *Client) ExecWithSession(cmd string, opts ...ExecOpt) (string, error) {
	options := execOptions{}
	for _, v := range opts {
		v(&options)
	}

	if c.session == nil {
		err := c.termSession()
		if err != nil {
			return "", NewRequestError("ssh连接失败", err)
		}
	}

	var out string

	err := c.encode(&cmd)
	if err != nil {
		return "", err
	}

	_, err = c.in.Write([]byte(cmd))
	if err != nil {
		return "", NewRequestError(fmt.Sprintf("ssh操作失败,%s", cmd), err)
	}

	buff := make([]byte, 2048)
	prompt := options.prompt
	if prompt == "" {
		prompt = c.prompt
	}

	cmdFinish := CmdFinish
	if options.execFun != nil {
		cmdFinish = CmdStart
	}

	for {
		n, err := c.out.Read(buff)
		if err != nil {
			return "", NewRequestError(fmt.Sprintf("ssh操作失败,%s", cmd), err)
		}

		resStr := strings.Split(string(buff[:n]), "\r\n")
		for _, val := range resStr {
			if strings.HasPrefix(val, "error -") {
				return "", NewRequestError(fmt.Sprintf("ssh操作失败,%s", val), err)
			}
		}

		out += string(buff[:n])
		if strings.Contains(out, prompt) && cmdFinish == CmdFinish {
			l := strings.Split(out, prompt)
			out = l[0]
			break
		} else {
			if options.execFun != nil {
				if err := options.execFun(&cmdFinish, &out, c.in); err != nil {
					return "", NewRequestError(fmt.Sprintf("ssh操作失败,%s", cmd), err)
				}
				if cmdFinish == PromptFounded {
					break
				}
				if cmdFinish == CmdFinish {
					if strings.Contains(out, prompt) {
						l := strings.Split(out, prompt)
						out = l[0]
						break
					}
				}
			}
		}
	}

	return c.sanitizeOutPut(&out, cmd, &options)
}

func (c *Client) sanitizeOutPut(data *string, cmd string, options *execOptions) (string, error) {
	if options.ignoreSanitizeOutPut {
		return *data, nil
	}

	cmd = strings.TrimSpace(cmd)

	l := strings.Split(*data, cmd)
	if len(l) > 1 {
		err := c.decode(&l[1])
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(l[1]), nil
	}
	if len(l) == 0 {
		return "", nil
	}

	err := c.decode(&l[0])
	if err != nil {
		return "", err
	}

	return l[0], nil
}
