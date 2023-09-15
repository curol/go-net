package reader

import (
	"fmt"
	"io"
	"os"
)

func ReadStream(fn string) {
	file, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		fmt.Print(string(buffer[:n]))
	}
}
