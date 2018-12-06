package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

var routes = flag.Bool("routes", false, "Generate router documentation")
var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("engagemanager"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(jwtauth.Claims{"client_id": "123"})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func main() {
	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(2 * time.Second))

	r.Get("/", responseRoot)

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Post("/", responseQr)
		r.Get("/list", qrList)
	})

	if *routes {
		fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
			ProjectPath: "github.com/tribalmedia/QrCodeApi",
			Intro:       "Welcome to the QrCodeApi/ generated docs.",
		}))
		return
	}

	http.ListenAndServe(":3333", r)

}

func qrList(w http.ResponseWriter, r *http.Request) {

	f, err := os.Open("qrcreate.log")
	if err != nil {
		log.Println("unable open qrcreate.log file.")
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for i := 0; i < 10; i++ {
		s.Scan()
		log.Println(s.Text())
	}
	if err := s.Err(); err != nil {
		log.Println("error read qrcreate.log file.")
	}

}

func responseRoot(w http.ResponseWriter, r *http.Request) {

	data := &Qrcodeurl{URL: "http://www.tribalmedia.co.jp/"}
	qrCode := createQr(data)
	createResponse(w, r, qrCode)

}

func responseQr(w http.ResponseWriter, r *http.Request) {

	data := &Qrcodeurl{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	qrCode := createQr(data)
	createResponse(w, r, qrCode)

}

func createResponse(w http.ResponseWriter, r *http.Request, qrCode barcode.Barcode) {

	w.Header().Set("Content-Type", "image/png")

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, qrCode); err != nil {
		log.Println("unable to encode png.")
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println("unable to write image.")
	}

}

func createQr(data *Qrcodeurl) (qrCode barcode.Barcode) {

	qrCode, _ = qr.Encode(data.URL, qr.H, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	urlEnc := url.QueryEscape(data.URL)
	file, _ := os.Create("images/" + urlEnc)
	defer file.Close()
	png.Encode(file, qrCode)

	createlog, err := os.OpenFile("qrcreate.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println("error create qrcreate.log file.")
	}
	defer createlog.Close()

	fmt.Fprintf(createlog, "%s,%s,%s\n", data.CREATEDTIME, data.CLIENTID, data.URL)

	return qrCode

}

// ErrResponse ...
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render ...
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest ...
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// Qrcodeurl ...
type Qrcodeurl struct {
	URL      string `json:"url"`
	CLIENTID string `json:"client_id"`
	CREATEDTIME string `json:created_time`
}

// Bind ...
func (q *Qrcodeurl) Bind(r *http.Request) error {
	if q.URL == "" {
		return errors.New("missing required Qrcodeurl fields. ")
	}
	log.Println(fmt.Sprintf("posted %s", q.URL))

	_, claims, _ := jwtauth.FromContext(r.Context())
	q.CLIENTID = fmt.Sprint(claims["client_id"])

	q.CREATEDTIME = time.Now().String()

	return nil
}
