package main

import (
	"fmt"

	llm "github.com/desarso/go_llm_functions/helpers"
)

func main() {
	fmt.Println(llm.Ell(sayHi)("Steven"))

}

func sayHi(user string) string {
	return fmt.Sprintf("Say hello to the user of name %s", user)
}
