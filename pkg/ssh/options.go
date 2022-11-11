package ssh

import (
	"io"
	"strconv"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

const (
	CmdStart = iota
	CmdFinish
	PromptFounded
	Auto     = "auto"
	Password = "password"
	Disabled = "disabled"
)

type connectionOptions struct {
	NetworkType  string
	ClientConfig *gossh.ClientConfig
	Address      string
	prompt       string
	promptEnd    string
	encoding     string
	authType     string
	user         string
	password     string
}

type ExecHandler func(cmdFinish *int, out *string, in io.WriteCloser) error

type execOptions struct {
	ignoreSanitizeOutPut bool
	execFun              ExecHandler
	prompt               string
}

type wsTermOptions struct {
	ptyCols uint32
	ptyRows uint32
	pingDur time.Duration
}

type scpOptions struct {
	fileName string
}

type ConnectionOpt func(*connectionOptions)

type ExecOpt func(*execOptions)

type WebSocketTermOpt func(*wsTermOptions)

type SCPOpt func(*scpOptions)

func WithTermSize(cols, rows string) WebSocketTermOpt {
	return func(options *wsTermOptions) {
		ptyCols, _ := strconv.ParseUint(cols, 10, 32)
		ptyRows, _ := strconv.ParseUint(rows, 10, 32)

		options.ptyRows = uint32(ptyRows)
		options.ptyCols = uint32(ptyCols)
	}
}

func WithFileName(filename string) SCPOpt {
	return func(options *scpOptions) {
		options.fileName = filename
	}
}

func WithPrompt(prompt string) ConnectionOpt {
	return func(options *connectionOptions) {
		options.prompt = prompt
	}
}

func WithEncoding(encoding string) ConnectionOpt {
	return func(options *connectionOptions) {
		options.encoding = encoding
	}
}

func WithPromptEnd(prompt string) ConnectionOpt {
	return func(options *connectionOptions) {
		options.promptEnd = prompt
	}
}

func WithExecFun(fun ExecHandler) ExecOpt {
	return func(options *execOptions) {
		options.execFun = fun
	}
}

func WithExecPrompt(prompt string) ExecOpt {
	return func(options *execOptions) {
		options.prompt = prompt
	}
}

func WithKeyboardInteractive() ConnectionOpt {
	return func(options *connectionOptions) {
		options.authType = "keyboardInteractive"
	}
}

func WithIgnoreSanitizeOutPut() ExecOpt {
	return func(options *execOptions) {
		options.ignoreSanitizeOutPut = true
	}
}
