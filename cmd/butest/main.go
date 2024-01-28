package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/vedranvuk/bleveutils"
)

func check(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

func main() {
	if len(os.Args[1:]) < 1 {
		fmt.Println("usage: butest <query>")
		return
	}
	var idx, err = bleve.Open("index.bleve")
	switch err {
	case bleve.ErrorIndexPathDoesNotExist:
		fmt.Println("Building index...")
		idx, err = bleveutils.Build("index.bleve", dmcb, fmcb, doc())
		check(idx.Index("data", doc()))
		check(err)
	case nil:
	default:
		check(err)
	}
	var fields []string
	fields, err = idx.Fields()
	check(err)
	fmt.Printf("Indexed fields: %v\n", fields)
	var res *bleve.SearchResult
	res, err = idx.Search(
		bleve.NewSearchRequest(
			bleve.NewQueryStringQuery(
				strings.Join(os.Args[1:], " "),
			),
		),
	)
	check(err)
	fmt.Println(res)
}

func dmcb(typ reflect.Type, m *mapping.DocumentMapping) *mapping.DocumentMapping {
	return m
}

func fmcb(typ reflect.Type, m *mapping.FieldMapping) *mapping.FieldMapping {
	return m
}

func doc() any {
	return &User{
		Person: Person{
			FirstName: "Vedran",
			LastName:  "Vuk",
			Age:       42,
		},
		Registered:  true,
		DateCreated: time.Date(1982, 6, 15, 17, 25, 1, 2, time.Local),
		LuckyNums: [3]int{
			42,
			69,
			1337,
		},
		Nicknames: []string{
			"veki",
		},
	}
}

type Person struct {
	FirstName string
	LastName  string
	Age       int
}

type User struct {
	Person
	Registered  bool
	DateCreated time.Time
	LuckyNums   [3]int
	Nicknames   []string
}
