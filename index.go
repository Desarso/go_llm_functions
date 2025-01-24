package main

import (
	llm "desarso/go_llm_functions/helpers"
	"fmt"
)

func main() {
	fmt.Println(llm.Ell(sayHi)("Steven"))

}

func sayHi(user string) string {
	return fmt.Sprintf("Say hello to the user of name %s", user)
}
