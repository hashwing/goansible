package actions

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math"
	"os"
)

func FileMd5(path string) (string, error) {
	fs, err := os.Open(path)
	if err != nil {
		return "", err
	}
	info, err := fs.Stat()
	if err != nil {
		return "", err
	}
	filesize := info.Size()
	const filechunk = 4 * 1 << 20
	blocks := uint64(math.Ceil(float64(filesize) / float64(filechunk)))
	hash := md5.New()
	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
		buf := make([]byte, blocksize)
		fs.Read(buf)
		io.WriteString(hash, string(buf))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func MD5(s string) string {
	hash := md5.New()
	io.WriteString(hash, s)
	return hex.EncodeToString(hash.Sum(nil))
}
