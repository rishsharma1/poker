package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/lonng/nano"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/serialize/json"
)

func main() {
	components := &component.Components{}
	components.Register(
		NewRoomManager(),
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	log.SetFlags(log.LstdFlags | log.Llongfile)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	nano.Listen(":3250",
		nano.WithIsWebsocket(true),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		nano.WithWSPath("/nano"),
		nano.WithDebugMode(),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(components),
	)
}
