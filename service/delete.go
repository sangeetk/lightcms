package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"git.urantiatech.com/cloudcms/cloudcms/api"
	"github.com/boltdb/bolt"
	"github.com/urantiatech/kit/endpoint"
)

// Delete - creates a single item
func (s *Service) Delete(ctx context.Context, req *api.DeleteRequest, sync bool) (*api.Response, error) {
	var resp = api.Response{Type: req.Type, Language: req.Language}
	var db *bolt.DB
	var err error

	if _, ok := Index[req.Type]; !ok {
		resp.Err = api.ErrorInvalidContentType.Error()
		return &resp, nil
	}

	// Open database in read-write mode
	// It will be created if it doesn't exist.
	//options := bolt.Options{ReadOnly: false}

	db, err = bolt.Open(DBFile, 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bb, err := getBucket(tx, req.Type, req.Language)
		if err != nil {
			return err
		}

		// Get the existing value
		val := bb.Get([]byte(req.Slug))
		if val == nil {
			return api.ErrorNotFound
		}

		err = json.Unmarshal(val, &resp.Content)
		if err != nil {
			return err
		}

		err = bb.Delete([]byte(req.Slug))
		if err != nil {
			return err
		}

		index, err := getIndex(req.Type, req.Language)
		if err != nil {
			return err
		}
		err = index.Delete(req.Slug)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		resp.Err = err.Error()
		return &resp, nil
	}

	return &resp, nil
}

// DeleteEndpoint - creates endpoint for Delete service
func DeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(api.DeleteRequest)
		return svc.Delete(ctx, &req, false)
	}
}

// DecodeDeleteReq - decodes the incoming request
func DecodeDeleteReq(ctx context.Context, r *http.Request) (interface{}, error) {
	var request api.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}
