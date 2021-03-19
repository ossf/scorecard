package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jszwec/csvutil"
)

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

// Checks for duplicate item in the projects.txt
// This is used in the builds to validate there aren't duplicates in projects.txt.
func main() {
	projects, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer projects.Close()

	repos := []Repository{}
	data, err := ioutil.ReadAll(projects)
	if err != nil {
		panic(err)
	}
	err = csvutil.Unmarshal(data, &repos)
	if err != nil {
		panic(err)
	}

	m := make(map[string]bool)
	for _, item := range repos {
		if _, ok := m[item.Repo]; ok {
			log.Panicf("Item already in the list %s", item.Repo)
		}
		m[item.Repo] = true
	}
}
