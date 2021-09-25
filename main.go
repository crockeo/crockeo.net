package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	address, port := serverAddress()
	serverHandler := makeServerHandler()
	server := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", address, port),
		Handler: serverHandler,
	}
	log.Printf("listening on %v:%v...\n", address, port)

	var err error
	if useTLS() {
		err = server.ListenAndServeTLS(
			"/etc/letsencrypt/live/crockeo.net/fullchain.pem",
			"/etc/letsencrypt/live/crockeo.net/privkey.pem",
		)
	} else {
		err = server.ListenAndServe()
	}
	log.Fatal(err)
}

func serverAddress() (net.IP, uint16) {
	address := net.ParseIP(os.Getenv("SERVER_ADDRESS"))
	if address == nil {
		address = net.IPv4(127, 0, 0, 1)
	}

	unstructured_port := os.Getenv("SERVER_PORT")
	port, err := strconv.ParseUint(unstructured_port, 10, 16)
	if err != nil {
		port = 8080
	}

	return address, uint16(port)
}

func useTLS() bool {
	return len(os.Getenv("SKIP_TLS")) == 0
}

type funcHandler struct {
	handlerFunc http.HandlerFunc
}

func newFuncHandler(handlerFunc http.HandlerFunc) funcHandler {
	return funcHandler{
		handlerFunc,
	}
}

func (f funcHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	f.handlerFunc(res, req)
}

func makeServerHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHomepage)
	mux.HandleFunc("/qr.png", serveQRCode)
	mux.HandleFunc("/qr", serveRedirect)
	return accessMiddleware(mux)
}

func accessMiddleware(next http.Handler) http.Handler {
	return newFuncHandler(func(res http.ResponseWriter, req *http.Request) {
		log.Printf("%v %v\n", req.Method, req.URL.Path)
		next.ServeHTTP(res, req)
	})
}

func serveHomepage(res http.ResponseWriter, req *http.Request) {
	file, err := os.Open("static/index.html")
	if err != nil {
	}

	contents, err := ioutil.ReadAll(file)
	res.WriteHeader(200)
	res.Write(contents)
}

func serveQRCode(res http.ResponseWriter, req *http.Request) {
	png, err := qrcode.Encode("https://crockeo.net/qr", qrcode.Medium, 1024)
	if err != nil {
	}

	res.Header().Add("Content-Type", "image/png")
	res.WriteHeader(200)
	res.Write(png)
}

func serveRedirect(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Location", "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	res.WriteHeader(307)
}
