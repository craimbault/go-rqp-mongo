package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	rqp "github.com/craimbault/go-rqp-mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("testgo").Collection("testcoll")
}

func main() {

	testId := primitive.NewObjectID()
	// Filter is parameter provided in the Query part of the URL
	//   The lib handles system filters:
	//     * fields - list of fields separated by comma (",") for SELECT statement. Should be validated.
	//     * sort   - list of fields separated by comma (",") for 	DER BY statement. Should be validated. Could includes prefix +/- which means ASC/DESC sorting. Eg. &sort=-id will be ORDER BY id DESC.
	//     * limit  - number for LIMIT statement. Should be greater then 0 by default.
	//     * offset - number for OFFSET statement. Should be greater then or equal to 0 by default.
	//   and user defined filters.
	//
	// Validation is a function for validate some Filter
	//
	// Field is enumerated in the Filter "fields" field which lib must put into SELECT statement.

	// url, _ := url.Parse("http://localhost/?sort=+name_test,-id&limit=10&id[in]=64e9c6d61209c16ffaa3062e,64e9c6d61209c16ffaa3062f&i[gte]=5&s[in]=one,two&email[like]=tim|name_test[like]=*tim*")
	// url, _ := url.Parse("http://localhost/?sort=-name,-id&limit=10&i[gte]=5&s=one|s=two&email[like]=*tim*|name[like]=*tim*")
	// url, _ := url.Parse("http://localhost/?sort=+name,-id&limit=10&i[gte]=5&s[eq]=one&name[like]=*tim*")
	// url, _ := url.Parse("http://localhost/?sort=+name,-id&limit=10&i[gte]=2&s[eq]=one")
	// url, _ := url.Parse("http://localhost/?id=1,2")
	url, _ := url.Parse("http://localhost/?id=" + testId.Hex() + "&limit=10")
	q, err := rqp.NewParseReplaced(
		url.Query(),
		rqp.Validations{
			// FORMAT: [field name] : [ ValidationFunc | nil ]

			// validation will work if field will be provided in the Query part of the URL
			// but if you add ":required" tag the Parser raise an Error if the field won't be in the Query part

			// special system fields: fields, limit, offset, sort
			// filters "fields" and "sort" must be always validated
			// If you won't define ValidationFunc but include "fields" or "sort" parameter to the URL the Parser raises an Error
			"limit:required": rqp.MinMax(10, 100),       // limit must present in the Query part and must be between 10 and 100 (default: Min(1))
			"sort":           rqp.In("id", "name_test"), // sort could be or not in the query but if it is present it must be equal to "in" or "name"

			"s":          rqp.In("one", "two"), // filter: s - string and equal
			"id:mongoid": nil,                  // filter: id is mongoid without additional validation (can be int for SQL)
			// "id:int": nil, // filter: id is mongoid without additional validation (can be int for SQL)
			"i:int": func(value interface{}) error { // filter: custom func for validating
				if value.(int) > 1 && value.(int) < 10 {
					return nil
				}
				return errors.New("i: must be greater then 1 and lower then 10")
			},
			"email":     nil,
			"name_test": nil,
		},
		rqp.Replacer{"name_test": "name"},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("FULL SQL :", q.SQL("table")) // SELECT id, name FROM table ORDER BY id LIMIT 10
	fmt.Println("ARGS :", q.Args())           // [1 5 one %tim% %tim%]

	mongoQueryFilters, mongoQueryFiltersErr := q.MongoQueryFilters()
	if mongoQueryFiltersErr != nil {
		fmt.Println("ERR : Mongo Filters error :", mongoQueryFiltersErr.Error())
	} else {
		fmt.Println("Mongo Filters : ", mongoQueryFilters)
		// fmt.Println("Mongo Projection : ", q.MongoProjection())
		fmt.Println("Mongo Order : ", q.MongoOrder())

		cur, err := MongoCollectionFind(collection, ctx, q)
		if err != nil {
			fmt.Println("Mongo ERR :", err.Error())
		} else {
			count := 0
			for cur.Next(ctx) {
				count += 1
				fmt.Println("Mongo Result :", cur.Current)
			}
			fmt.Println("Mongo Result found :", count)
		}
	}

	q.AddValidation("fields", rqp.In("id", "name"))
	q.SetUrlString("http://localhost/?fields=id,name&limit=10")
	q.Parse()

	fmt.Println("FULL SQL :", q.SQL("table")) // SELECT id, name FROM table ORDER BY id LIMIT 10
	fmt.Println("ARGS :", q.Args())           // []

	mongoQueryFilters, mongoQueryFiltersErr = q.MongoQueryFilters()
	if mongoQueryFiltersErr != nil {
		fmt.Println("ERR : Mongo Filters error :", mongoQueryFiltersErr.Error())
	} else {
		fmt.Println("Mongo Filters : ", mongoQueryFilters)
		// fmt.Println("Mongo Projection : ", q.MongoProjection())
		fmt.Println("Mongo Order : ", q.MongoOrder())

		// cur, err := q.MongoCollectionFind(collection, ctx)
		// if err != nil {
		// 	fmt.Println("Mongo ERR :", err.Error())
		// } else {
		// 	for cur.Next(ctx) {
		// 		fmt.Println("Mongo Result :", cur.Current)
		// 	}
		// }
	}

	q.AddValidation("fields", rqp.In("id", "name"))
	q.SetUrlString("http://localhost/?fields=id,name&limit=10")
	q.Parse()

	fmt.Println("FULL SQL :", q.SQL("table")) // SELECT id, name FROM table ORDER BY id LIMIT 10
	fmt.Println("ARGS :", q.Args())           // []

	mongoQueryFilters, mongoQueryFiltersErr = q.MongoQueryFilters()
	if mongoQueryFiltersErr != nil {
		fmt.Println("ERR : Mongo Filters error :", mongoQueryFiltersErr.Error())
	} else {
		fmt.Println("Mongo Filters : ", mongoQueryFilters)
		// fmt.Println("Mongo Projection : ", q.MongoProjection())
		fmt.Println("Mongo Order : ", q.MongoOrder())

		// cur, err := q.MongoCollectionFind(collection, ctx)
		// if err != nil {
		// 	fmt.Println("Mongo ERR :", err.Error())
		// } else {
		// 	for cur.Next(ctx) {
		// 		fmt.Println("Mongo Result :", cur.Current)
		// 	}
		// }
	}
}

func MongoCollectionFind(collection *mongo.Collection, ctx context.Context, q *rqp.Query) (*mongo.Cursor, error) {
	filters, err := q.MongoQueryFilters()
	if err != nil {
		return nil, err
	}

	options := options.FindOptions{}
	q.MongoAddFindOptions(&options)

	return collection.Find(
		ctx,
		filters,
		&options,
	)
}
