package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	runtime "github.com/aws/aws-lambda-go/lambda"
)

// Request is the payload sent by asp_upload (invoked as a Lambda).
type Request struct {
	Doctype  string `json:"doctype"`
	From     string `json:"from"`
	S3Bucket string `json:"s3_bucket"`
	S3Key    string `json:"s3_key"`
	Ctype    string `json:"ctype"` // "excel" or "cvs"
	Date1    string `json:"date1"` // RFC3339 or empty
	Date2    string `json:"date2"` // RFC3339 or empty
	Brand    string `json:"brand"`
}

// Response is returned to asp_upload. asp_upload is responsible for emailing
// the sender based on these fields.
type Response struct {
	OK         bool   `json:"ok"`
	Authorized bool   `json:"authorized"`
	Summary    string `json:"summary"`
	Error      string `json:"error"`
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func handleRequest(ctx context.Context, req Request) (Response, error) {

	doctype := strings.ToLower(strings.TrimSpace(req.Doctype))

	// Minimal validation: known document type.
	switch doctype {
	case "asp", "inventory", "obsoletos", "backorders":
	default:
		log.Println("Tipo de documento desconocido:", req.Doctype)
		return Response{OK: false, Error: "tipo de documento desconocido: " + req.Doctype}, nil
	}

	// Minimal validation: a file must have been provided.
	if req.S3Key == "" {
		log.Println("Solicitud sin archivo (s3_key vacío)")
		return Response{OK: false, Error: "no se recibió ningún archivo para procesar"}, nil
	}

	formato := req.Ctype
	if formato == "" {
		formato = "cvs"
	}
	log.Printf("Procesando documento tipo=%q formato=%q archivo=%q remitente=%q", doctype, formato, req.S3Key, req.From)

	// Authorization: the sender must be a registered user.
	if !Validate_email(req.From) {
		log.Println("Correo no autorizado:", req.From)
		return Response{OK: true, Authorized: false, Error: "El correo no está autorizado para procesar archivos."}, nil
	}

	bucket := req.S3Bucket
	if bucket == "" {
		bucket = getenv("S3_BUCKET", "asp-emails")
	}

	tmpfile, err := downloadFromS3(bucket, req.S3Key)
	if err != nil {
		return Response{OK: false, Authorized: true, Error: "error descargando el archivo: " + err.Error()}, nil
	}
	defer os.Remove(tmpfile)

	date1 := parseTime(req.Date1)
	date2 := parseTime(req.Date2)
	if date1 == nil {
		now := time.Now()
		date1 = &now
	}

	var summary string

	switch doctype {
	case "asp":
		if date2 == nil {
			date2 = date1
		}
		if req.Ctype == "excel" {
			summary, err = process_xml_asp(tmpfile, *date1, *date2, req.Brand)
		} else {
			summary, err = process_cvs_asp(tmpfile, *date1, *date2, req.Brand)
		}
		if err == nil {
			update_date_catalogs(date1, date2, "asp", req.Brand)
		}
	case "inventory":
		summary, err = process_cvs_inventory(tmpfile)
		if err == nil {
			update_date_catalogs(date1, nil, "inventory", req.Brand)
		}
	case "obsoletos":
		summary, err = process_cvs_obsoletos(tmpfile)
		if err == nil {
			update_date_catalogs(date1, nil, "obsoletos", req.Brand)
		}
	case "backorders":
		summary, err = process_cvs_backorders(tmpfile)
		if err == nil {
			update_date_catalogs(date1, nil, "backorders", req.Brand)
		}
	}

	if err != nil {
		log.Printf("Error procesando documento tipo=%q: %v", doctype, err)
		return Response{OK: false, Authorized: true, Error: err.Error()}, nil
	}

	log.Printf("Documento tipo=%q procesado correctamente", doctype)
	return Response{OK: true, Authorized: true, Summary: summary}, nil
}

func main() {
	log.Print("Starting asp_db Lambda")
	runtime.Start(handleRequest)
}
