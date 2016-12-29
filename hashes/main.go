package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var (
	algo string
	text string
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: hashes -a bcrypt -t \"some text\"")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.StringVar(&algo, "a", "bcrypt", "type of hash algorithm")
	flag.StringVar(&text, "t", "", "text to hash")
	flag.Parse()

	if text == "" {
		flag.Usage()
		os.Exit(1)
	}

	switch algo {
	case "bcrypt":
		out(BCrypt())
	case "md5":
		out(MD5())
	default:
		fmt.Fprintf(os.Stdout, "Unknown type of hash algorithm - %s\n", algo)
	}
}

func BCrypt() string {
	btext := []byte(text)
	htext, err := bcrypt.GenerateFromPassword(btext, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(htext)
}

func MD5() string {
	h := md5.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}

func out(hash string) {
	fmt.Fprintln(os.Stdout, "Input:")
	fmt.Fprintf(os.Stdout, "\talgorithm - %s\n", algo)
	fmt.Fprintf(os.Stdout, "\ttext - %s\n", text)
	fmt.Fprintln(os.Stdout, "Output:")
	fmt.Fprintf(os.Stdout, "\thash - %s\n", hash)
}
