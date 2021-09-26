package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	address := serverAddress()
	usingTLS := useTLS()
	var port uint16
	if usingTLS {
		port = 443
	} else {
		port = 80
	}
	server := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", address, port),
		Handler: makeServerHandler(),
	}
	log.Printf("listening on %v:%v...\n", address, port)

	var err error
	if useTLS() {
		group := sync.WaitGroup{}
		group.Add(2)
		go func() {
			var handler http.Handler
			handler = newFuncHandler(serveHTTPSRedirect)
			handler = accessMiddleware(handler)
			http.ListenAndServe(fmt.Sprintf("%v:80", address), handler)
			group.Done()
		}()

		go func() {
			err = server.ListenAndServeTLS(
				"/etc/letsencrypt/live/crockeo.net/fullchain.pem",
				"/etc/letsencrypt/live/crockeo.net/privkey.pem",
			)
			group.Done()
		}()
		group.Wait()
	} else {
		err = server.ListenAndServe()
	}
	log.Fatal(err)
}

func serverAddress() net.IP {
	address := net.ParseIP(os.Getenv("SERVER_ADDRESS"))
	if address == nil {
		address = net.IPv4(127, 0, 0, 1)
	}

	return address
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
	mux.Handle("/static/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/qr.png", serveQRCode)
	mux.HandleFunc("/qr", serveRedirect)
	mux.HandleFunc("/", serveHomepage)
	return accessMiddleware(mux)
}

func accessMiddleware(next http.Handler) http.Handler {
	return newFuncHandler(func(res http.ResponseWriter, req *http.Request) {
		log.Printf("%v %v\n", req.Method, req.URL.Path)
		next.ServeHTTP(res, req)
	})
}

func serveHTTPSRedirect(res http.ResponseWriter, req *http.Request) {
	newLocation := fmt.Sprintf("https://crockeo.net%v", req.URL.Path)
	res.Header().Add("Location", newLocation)
	res.WriteHeader(301)
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

func serveHomepage(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	index, err := readFile("static/index.html")
	if err != nil {
	}
	contents, err := interpolateBody(index)
	if err != nil {
	}

	res.WriteHeader(200)
	res.Write(contents)
}

func interpolateBody(body []byte) ([]byte, error) {
	header, err := readFile("static/header.html")
	if err != nil {
		return []byte{}, err
	}

	footer, err := readFile("static/footer.html")
	if err != nil {
		return []byte{}, err
	}

	return append(header, append(body, footer...)...), nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}
