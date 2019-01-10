package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"github.com/boombuler/barcode/qr"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/boombuler/barcode"
	"github.com/dgrijalva/jwt-go"
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
	_, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"client_id": "123"})
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
			ProjectPath: "github.com/stakada7/QrCodeApi",
			Intro:       "Welcome to the QrCodeApi/ generated docs.",
		}))
		return
	}

	http.ListenAndServe(":3333", r)

}

func responseRoot(w http.ResponseWriter, r *http.Request) {

	data := &qrcodeinfo{URL: "https://stakada7.com/"}
	qrCode := createQr(data)
	createResponse(w, r, qrCode)

}

func responseQr(w http.ResponseWriter, r *http.Request) {

	data := &qrcodeinfo{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errInvalidRequest(err))
		return
	}
	qrCode := createQr(data)
	createResponse(w, r, qrCode)

}

func qrList(w http.ResponseWriter, r *http.Request) {

	f, err := os.Open("qrcreate.log")
	if err != nil {
		render.Render(w, r, errInvalidRequest(err))
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	record, err := reader.ReadAll()
	if err != nil {
		render.Render(w, r, errInvalidRequest(err))
		return
	}

	var qrlist qrlist
	for _, line := range record {
		info := qrcodeinfo{CREATEDTIME: line[0], CLIENTID: line[1], URL: line[2]}
		qrlist.List = append(qrlist.List, info)
	}

	render.JSON(w, r, qrlist)

}

func createQr(data *qrcodeinfo) (qrCode barcode.Barcode) {

	qrCode, _ = qr.Encode(data.URL, qr.H, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 200, 200)

	urlEnc := url.QueryEscape(data.URL)
	p, _ := os.Create("images/" + urlEnc)
	defer p.Close()
	png.Encode(p, qrCode)

	f, err := os.OpenFile("qrcreate.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println("error create qrcreate.log file.")
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	l := []string{data.CREATEDTIME, data.CLIENTID, data.URL}
	err = writer.Write(l)
	if err != nil {
		log.Println("error write qrcreate.log file.")
	}
	writer.Flush()

	return qrCode

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

// errResponse ...
type errResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render ...
func (e *errResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// errInvalidRequest ...
func errInvalidRequest(err error) render.Renderer {
	return &errResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// qrlist ...
type qrlist struct {
	List []qrcodeinfo `json:"list"`
}

// qrcodeinfo ...
type qrcodeinfo struct {
	URL         string `json:"url"`
	CLIENTID    string `json:"client_id"`
	CREATEDTIME string `json:"created_time"`
}

// Bind ...
func (q *qrcodeinfo) Bind(r *http.Request) error {
	if q.URL == "" {
		return errors.New("missing required qrcodeinfo fields. ")
	}
	log.Println(fmt.Sprintf("posted %s", q.URL))

	_, claims, _ := jwtauth.FromContext(r.Context())
	q.CLIENTID = fmt.Sprint(claims["client_id"])

	q.CREATEDTIME = time.Now().String()

	return nil
}
