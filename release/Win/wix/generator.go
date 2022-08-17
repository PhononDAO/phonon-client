package main

import (
	"text/template"
	"log"
	"os"
	"strings"
)

type UuidGenerator struct{}

func (UuidGenerator)UUID()string{
	builder := strings.NewBuilder()	
}

func main(){
	 templ, err := template.ParseFiles("phonon.wxs.templ")
	 if err != nil{
		log.Fatal(err)
	 }
	 templ.Execute(os.Stdout, UuidGenerator{})
}
