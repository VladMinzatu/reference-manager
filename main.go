package main

import (
	"github.com/VladMinzatu/reference-manager/web"
)

func main() {
	handler := web.NewHandler()
	web.StartServer(handler)
}
