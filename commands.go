package main

import (
	"fmt"
	"os"
)

func commands() {
	if len(os.Args) == 3 && os.Args[1] == "html" {
		p, err := loadPage(os.Args[2]);
		if err != nil {
			fmt.Println(err);
		} else {
			p.renderHtml();
			fmt.Println(p.Html);
		}
	} else {
		fmt.Printf("Unknown command: %v\n", os.Args[1:])
		fmt.Print("Without any arguments, serves a wiki.\n")
		fmt.Print(" Environment variable ODDMUSE_PORT controls the port.\n")
		fmt.Print(" Environment variable ODDMUSE_LANGAUGES controls the languages detected.\n")
		fmt.Print("html PAGENAME\n")
		fmt.Print(" Print the HTML of the page.\n")
	}
}
