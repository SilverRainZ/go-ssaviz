package a

import "fmt"

func print(s string) {
	if s == "" {
		fmt.Println("nothing")
	} else {
		fmt.Println("value:", s)
	}
}
