package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("http://www.google.com")

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// b := make([]byte, 99999)
	// resp.Body.Read(b)
	// println(b)

	// Copy basically reads data from resp.Body which implements Reader
	// and moves it to os.Stdout, which implements Writer. Writer takes
	// some byteslice data and writes it somewhere, in this case the
	// standard output of the OS. But it can be a file as well for example
	io.Copy(os.Stdout, resp.Body)
}
