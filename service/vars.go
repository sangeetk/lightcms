package service

import (
	"github.com/blevesearch/bleve"
)

// DBFile is the path of database file
var DBFile string

// IndexFile is the path of index file
var IndexFile string

// DefaultBucket name
const DefaultBucket = "default"

// Index map[ContentType]map[Language]bleve.Index
var Index map[string]map[string]bleve.Index
