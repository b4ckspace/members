package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/b4ckspace/members/internal/ssha"
)

type (
	Args struct {
		Algo string
	}
)

func main() {
	a := Args{}
	flag.StringVar(&a.Algo, "algo", "SSHA512", "hash algorythm")
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Password: ")
	password, _ := reader.ReadString('\n')
	password = password[:len(password)-1]
	fmt.Printf("'%s'", password)
	hash, err := ssha.Hash(password, ssha.HashAlgo(a.Algo))
	if err != nil {
		fmt.Printf("unable to hash: %s", err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
