package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func insert_record(part Part) error {
	ctx := context.Background()
	db := mgdb.Database("asp")
	collection := db.Collection("main_catalog")

	filter := bson.M{"cve": part.Cve}
	update := bson.M{
		"$setOnInsert": bson.M{
			"cve":  part.Cve,
			"name": part.Name,
		},
	}
	options := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return err
	}
	return nil
}

func update_date_catalogs(date1 *time.Time, date2 *time.Time, typedoc, brand string) {

	collection := mgdb.Database("asp").Collection("dbstats")

	filter := bson.M{}

	var stats Stats

	err := collection.FindOne(context.TODO(), filter).Decode(&stats)
	if err != nil {
		log.Println(err)
		return
	}

	var campo string
	switch typedoc {
	case "asp":
		campo = "PartsUpdate." + brand
		if brand == "Honda" {
			if stats.PartsUpdate.Honda.StartDate.Before(*date1) {
				stats.PartsUpdate.Honda.StartDate = *date1
				stats.PartsUpdate.Honda.EndDate = *date2
			}
		} else {
			if stats.PartsUpdate.Acura.StartDate.Before(*date1) {
				stats.PartsUpdate.Acura.StartDate = *date1
				stats.PartsUpdate.Acura.EndDate = *date2
			}
		}
	case "obsoletos":
		stats.ObsoletesUpdate = *date1
		campo = "ObsoletesUpdate"
	case "backorders":
		stats.BackordersUpdate = *date1
		campo = "BackordersUpdate"
	case "inventory":
		stats.InventoryUpdate = *date1
		campo = "InventoryUpdate"
	}

	collection.FindOneAndReplace(context.TODO(), filter, stats)
	log.Printf("Fecha de última actualización guardada en dbstats.%s (tipo=%q)", campo, typedoc)
}
