package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"git.urantiatech.com/cloudcms/cloudcms/api"
	"github.com/patrickmn/go-cache"
	"golang.org/x/text/language"
)

// Languages supported
var Languages []language.Tag

// Interface definition
type Interface interface {
	// Normal DB operations
	Create(context.Context, *api.CreateRequest, bool) (*api.Response, error)
	Read(context.Context, *api.ReadRequest) (*api.Response, error)
	Update(context.Context, *api.UpdateRequest, bool) (*api.Response, error)
	Delete(context.Context, *api.DeleteRequest, bool) (*api.Response, error)
	Search(context.Context, *api.SearchRequest) (*api.SearchResults, error)
	List(context.Context, *api.ListRequest) (*api.ListResults, error)

	// Schema request from admin interface
	Schema(context.Context, *api.SchemaRequest) (*api.SchemaResponse, error)
}

// Service struct for accessing services
type Service struct{}

// RespCache caches the response
var RespCache *cache.Cache

// Encode the response
func Encode(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func PrintCreateReq(r *api.CreateRequest) {
	json, err := json.MarshalIndent(r, "  ", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))
}
