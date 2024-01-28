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
		// idx, err = bleve.New("index.bleve", bleve.NewIndexMapping())
		idx, err = bleveutils.Build("index.bleve", nil, imcb, dmcb, fmcb, doc())
		check(err)
		err = idx.Index("data", doc())
		check(err)
		idx.Mapping()
	case nil:
		// idx.Mapping().SetOnDetermineType(determineType)
		check(idx.Index("new", &SomeNewType{"Hello!"}))
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

func determineType(document interface{}) (s string) {
	s = bleveutils.DocType(document)
	fmt.Printf("type: %s\n", s)
	return
}

func imcb(m *mapping.IndexMappingImpl) *mapping.IndexMappingImpl {
	m.StoreDynamic = false
	m.IndexDynamic = false
	return m
}

func dmcb(typ reflect.Type, m *mapping.DocumentMapping) *mapping.DocumentMapping {
	m.Dynamic = false
	return m
}

func fmcb(typ reflect.Type, m *mapping.FieldMapping) *mapping.FieldMapping {
	if m.Type == "text" {
		m.Analyzer = "en"
	}
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
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Age       int    `json:"age,omitempty"`
}

func (self Person) BleveType() string {
	return bleveutils.DocType(self)
}

type User struct {
	Person      `json:"person,omitempty"`
	Registered  bool      `json:"registered,omitempty"`
	DateCreated time.Time `json:"dateCreated,omitempty"`
	LuckyNums   [3]int    `json:"luckyNums,omitempty"`
	Nicknames   []string  `json:"nicknames,omitempty"`
}

func (self User) BleveType() string {
	return bleveutils.DocType(self)
}

type SomeNewType struct {
	NewField string `json:"name,omitempty"`
}
