package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// parseOrderDate tries to parse a date string in "02/01/2006" or "02-01-06" formats.
func parseOrderDate(dateStr string) (time.Time, error) {
	layouts := []string{"02/01/2006", "02-01-06"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	log.Println("Error reading order date " + dateStr)
	return time.Now(), errors.New("error leyendo la fecha de orden " + dateStr)
}

func process_cvs_backorders(filename string) (string, error) {

	var catalogs []BackordersParts

	f, err := os.Open(filename) // #nosec G304 -- filename is a server-side temp file, not user-controlled input.
	if err != nil {
		return "", err
	}

	retstr := "Resultados: \r\n"

	defer f.Close()

	r := csv.NewReader(f)

	// Validate header
	header, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("error leyendo encabezado: %v", err)
	}
	expectedHeader := []string{"item", "alterno", "dealer", "orden", "pedido", "qty", "fecha", "pronostico"}
	if len(header) != len(expectedHeader) {
		return "", fmt.Errorf("el archivo debe tener %d columnas, tiene %d", len(expectedHeader), len(header))
	}
	for i, v := range expectedHeader {
		if !strings.EqualFold(strings.TrimSpace(header[i]), v) {
			return "", fmt.Errorf("encabezado inválido en columna %d: se esperaba '%s', se obtuvo '%s'", i+1, v, header[i])
		}
	}
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		for i, v := range record {
			// Column 2 ("alterno") is optional and can be empty
			if i == 1 {
				continue
			}
			if v == "" {
				log.Printf("Error leyendo columna %d: valor vacío en registro %v\n", i+1, record)
				retstr += fmt.Sprintf("Error leyendo columna %d: valor vacío en registro %v \r\n", i+1, record)
				continue
			}
		}

		qty, err := strconv.ParseFloat(record[5], 32)
		if err != nil {
			log.Println("Error reading qty " + record[3])
			retstr += fmt.Sprintf("Error Clave %s leyendo la existencia %s \r\n", record[1], record[3])
			qty = 0
		}

		nodealer, err := strconv.Atoi(record[2])
		if err != nil {
			log.Println("Error reading dealer code " + record[0])
			retstr += fmt.Sprintf("Error leyendo el codigo de distribuidor %s \r\n", record[0])
			continue
		}

		part := BackordersParts{}

		part.Cve = record[0]
		part.Cve_Alt = record[1]
		part.NoDealer = nodealer
		part.NoOrder = record[3]
		part.NoPedido = record[4]
		part.Qty = int(qty)
		part.OrderDate, err = parseOrderDate(record[6])
		if err != nil {
			retstr += fmt.Sprintf("Error leyendo la fecha de orden clave %s fecha %s \n", record[1], record[4])
		}
		part.Forecast = record[7]
		catalogs = append(catalogs, part)

	}
	ctx := context.Background()
	db := mgdb.Database("asp")
	collection := db.Collection("backorders")
	err = collection.Drop(ctx)
	if err != nil {
		return "", err
	}

	var catalogInterfaces []interface{}
	for _, catalog := range catalogs {
		catalogInterfaces = append(catalogInterfaces, catalog)
	}
	_, err = collection.InsertMany(ctx, catalogInterfaces)
	if err != nil {
		return "", err
	}
	log.Printf("Backorders: insertados %d registros en la colección 'backorders'", len(catalogs))

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"no_dealer": 1,
		},
	}

	_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return "", err
	}

	indexModel = mongo.IndexModel{
		Keys: bson.M{
			"cve": 1,
		},
	}

	_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return "", err
	}

	retstr += fmt.Sprintf("Se procesaron %d registros \r\n", len(catalogs))
	return retstr, nil
}
