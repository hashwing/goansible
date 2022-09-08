package model

type Config struct {
	PlaybookFolder string
	Tag            string
	Tags           []string
	Untag          string
	IsUntag        bool
	InvFile        string
	PlaybookFile   string
}
