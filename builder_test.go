package bleveutils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
)

type Root struct {
	Time      time.Time
	Name      string
	Friends   []string
	Addresses [5]int
}

type Child struct {
	Root
	Age      int
	Employed bool
}

func TestBuild(t *testing.T) {
	os.RemoveAll("index.bleve")
	var idx, err = Build("index.bleve", nil, nil, Child{})
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	var p = &Child{
		Root{
			time.Now(),
			"Vedran",
			[]string{"Marty", "Dylan", "HAL"},
			[5]int{2, 5, 9, 17, 83},
		},
		41,
		true,
	}
	if err = idx.Index("yes", p); err != nil {
		t.Fatal(err)
	}

	var fields []string
	if fields, err = idx.Fields(); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Indexed fields: %v\n", fields)

	
	var res *bleve.SearchResult
	res, err = idx.Search(bleve.NewSearchRequest(bleve.NewQueryStringQuery("Name:vedran")))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
