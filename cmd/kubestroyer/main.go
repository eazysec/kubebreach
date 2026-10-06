package main

import (
	"fmt"

	"github.com/Auditeur/KubeBreach/pkg"
	"github.com/Auditeur/KubeBreach/utils"
)

func main() {

	fmt.Println("\x1b[1;36m" + utils.Toolname + "\x1b[0m")
	fmt.Println(utils.Author)
	fmt.Println(utils.Split)

	utils.Config()
	pkg.Execute()

}
