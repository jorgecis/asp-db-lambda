package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func process_xml_asp(filename string, date1 time.Time, date2 time.Time, brand string) (string, error) {

	var catalog Catalog

	catalog.Start_date = date1
	catalog.End_date = date2
	catalog.Brand = brand

	f, err := excelize.OpenFile(filename)
	if err != nil {
		return "", err
	}

	retstr := "Resultados: \r\n"

	list := f.GetSheetList()
	for _, sheet := range list {
		val, err := f.GetCellValue(sheet, "A1")
		if err != nil || val == "" {
			log.Println("Excel sheet is not valid")
			continue

		}

		row := 1
		for {
			cve, err := f.GetCellValue(sheet, "A"+strconv.Itoa(row))
			name, _ := f.GetCellValue(sheet, "B"+strconv.Itoa(row))
			pricestr, _ := f.GetCellValue(sheet, "C"+strconv.Itoa(row))
			if err != nil || cve == "" {
				log.Println("Error reading cell or is empty")
				log.Println(row)
				retstr += fmt.Sprintf("Error leyendo la celda A%d \r\n", row)
				break
			}
			price, err := strconv.ParseFloat(pricestr, 64)
			if err != nil {
				log.Println("Error reading price " + pricestr)
				price = 0
				retstr += fmt.Sprintf("Error leyendo el precio %s renglon %d \r\n", pricestr, row)
			}

			catalog.Parts = append(catalog.Parts, Part{Cve: cve, Name: name, Price: price})
			if err := insert_record(Part{Cve: cve, Name: name}); err != nil {
				log.Println("Error insertando en main_catalog:", err)
			}
			row++
		}
	}
	catalog.Parts_Count = len(catalog.Parts)

	ctx := context.Background()
	db := mgdb.Database("asp")
	_, err = db.Collection("catalog").InsertOne(ctx, catalog)
	if err != nil {
		return "", err
	}
	retstr += fmt.Sprintf("Se procesaron %d registros \r\n", len(catalog.Parts))
	retstr += fmt.Sprintf("Fecha inicial %s \r\n", catalog.Start_date.Format("02/01/2006"))
	retstr += fmt.Sprintf("Fecha final %s \r\n", catalog.End_date.Format("02/01/2006"))
	retstr += fmt.Sprintf("Marca %s \r\n", catalog.Brand)

	return retstr, nil

}

func process_cvs_asp(filename string, date1 time.Time, date2 time.Time, brand string) (string, error) {

	var catalog Catalog
	catalog.Start_date = date1
	catalog.End_date = date2
	catalog.Brand = brand

	f, err := os.Open(filename) // #nosec G304 -- filename is a server-side temp file, not user-controlled input.
	if err != nil {
		return "", err
	}

	retstr := "Resultados: \r\n"

	defer f.Close()

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		if record[0] == "" || record[1] == "" || record[2] == "" {
			continue
		}

		price, err := strconv.ParseFloat(record[2], 32)
		if err != nil {
			log.Println("Error reading price " + record[2])
			retstr += fmt.Sprintf("Error Clave %s leyendo el precio %s \r\n", record[0], record[2])
			price = 0
		}

		catalog.Parts = append(catalog.Parts, Part{Cve: record[0], Name: record[1], Price: price})
	}
	catalog.Parts_Count = len(catalog.Parts)

	ctx := context.Background()
	db := mgdb.Database("asp")
	_, err = db.Collection("catalog").InsertOne(ctx, catalog)
	if err != nil {
		return "", err
	}
	retstr += fmt.Sprintf("Se procesaron %d registros \r\n", len(catalog.Parts))
	retstr += fmt.Sprintf("Fecha inicial %s \r\n", catalog.Start_date.Format("02/01/2006"))
	retstr += fmt.Sprintf("Fecha final %s \r\n", catalog.End_date.Format("02/01/2006"))
	retstr += fmt.Sprintf("Marca %s \r\n", catalog.Brand)

	return retstr, nil

}
