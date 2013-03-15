// The maildir package provides an interface to mailboxes in the Maildir format.
package main //ldir

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	//"io/ioutil"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

var KeyError = errors.New("Key maps to multiple or no Messages")

// A Dir represents a single directory in a Maildir mailbox.
type Dir string

// Unseen moves messages from new to cur (they are now "seen") and returns their keys.
// This is the only function in this package that operates on any subdirectory
// other than "cur".
func (d Dir) Unseen() ([]string, error) {
	f, err := os.Open(string(d) + "/new/")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, n := range names {
		if n[0] != '.' {
			split := strings.FieldsFunc(n, func(r rune) bool {
				return r == ':'
			})
			keys = append(keys, split[0])
			os.Rename(string(d) + "/new/" + n, string(d) + "/cur/" + n + ":2,S")
		}
	}
	return keys, nil
}

// Keys returns a slice of valid keys to access messages by.
func (d Dir) Keys() ([]string, error) {
	f, err := os.Open(string(d) + "/cur/")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, n := range names {
		if n[0] != '.' {
			split := strings.FieldsFunc(n, func(r rune) bool {
				return r == ':'
			})
			keys = append(keys, split[0])
		}
	}
	return keys, nil
}

func (d Dir) filename(key string) (string, error) {
	matches, err := filepath.Glob(string(d) + "/cur/" + key + "*")
	if err != nil {
		return "", err
	}
	if matches == nil || len(matches) > 1 {
		return "", KeyError
	}
	return matches[0], nil
}

// Header returns the corresponding mail header to a key.
func (d Dir) Header(key string) (header mail.Header, err error) {
	filename, err := d.filename(key)
	if err != nil {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	tp := textproto.NewReader(bufio.NewReader(file))
	hdr, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}
	header = mail.Header(hdr)
	return
}

// Headers returns headers for all mail in this Dir.
func (d Dir) Headers() ([]mail.Header, error) {
	ks, err := d.Keys()
	if err != nil {
		return nil, err
	}
	headers := make([]mail.Header, 0)
	for _, k := range ks {
		h, err := d.Header(k)
		if err != nil {
			return headers, err
		}
		headers = append(headers, h)
	}
	return headers, nil
}

func (d Dir) Message(key string) (*mail.Message, error) {
	filename, err := d.filename(key)
	if err != nil {
		return &mail.Message{}, err
	}
	r, err := os.Open(filename)
	if err != nil {
		return &mail.Message{}, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	if err != nil {
		return &mail.Message{}, err
	}
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func main() {
	d := Dir("~/mdtest")
	ks, _ := d.Unseen()
	fmt.Println(ks)
}
