package main

import (
	"context"
	"encoding/csv"
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

func process_cvs_inventory(filename string) (string, error) {

	var catalogs []Inventory

	f, err := os.Open(filename) // #nosec G304 -- filename is a server-side temp file, not user-controlled input.

	if err != nil {
		return "", err
	}

	retstr := "Resultados: \r\n"

	defer f.Close()

	r := csv.NewReader(f)
mainloop:
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		if record[1] == "CODIGO" {
			record, _ = r.Read()
		}

		if record[0] == "" || record[1] == "" || record[2] == "" || record[3] == "" {
			continue
		}

		qty, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			log.Println("Error reading qty " + record[3])
			retstr += fmt.Sprintf("Error Clave %s leyendo la existencia %s \r\n", record[1], record[3])
			qty = 0
		}

		dealerdata := strings.Split(record[0], "   ")
		if len(dealerdata) < 5 {
			log.Println("Error reading dealer code " + record[0])
			retstr += fmt.Sprintf("Error leyendo el codigo de distribuidor %s \r\n", record[0])
			continue
		}

		nodealer, err := strconv.Atoi(dealerdata[0])
		if err != nil {
			log.Println("Error reading dealer code " + record[0])
			retstr += fmt.Sprintf("Error leyendo el codigo de distribuidor %s \r\n", record[0])
			continue
		}

		part := InventoryParts{}
		part.Cve = record[1]
		part.Name = record[2]
		part.Qty = int(qty)

		for i := range catalogs {
			if catalogs[i].NoDealer == nodealer {
				catalogs[i].Parts = append(catalogs[i].Parts, part)
				continue mainloop
			}
		}

		catalog := Inventory{}
		catalog.NoDealer, _ = strconv.Atoi(dealerdata[0])
		catalog.Name = dealerdata[1]
		catalog.ContactName = dealerdata[2]
		catalog.ContactEmail = dealerdata[3]
		catalog.ContactPhone = dealerdata[4]
		catalog.Parts = append(catalog.Parts, part)
		catalogs = append(catalogs, catalog)

	}
	ctx := context.Background()
	db := mgdb.Database("asp")
	collection := db.Collection("inventory")
	err = collection.Drop(ctx)
	if err != nil {
		return "", err
	}
	time.Sleep(100 * time.Millisecond)

	var catalogInterfaces []interface{}
	for _, catalog := range catalogs {
		catalogInterfaces = append(catalogInterfaces, catalog)
	}
	_, err = collection.InsertMany(ctx, catalogInterfaces)
	if err != nil {
		return "", err
	}

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"parts.cve": 1,
		},
	}

	_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return "", err
	}
	retstr += fmt.Sprintf("Se procesaron %d registros \r\n", len(catalogs))
	return retstr, nil
}
