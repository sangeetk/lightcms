package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"git.urantiatech.com/cloudcms/cloudcms/api"
	"git.urantiatech.com/cloudcms/cloudcms/item"
	"github.com/boltdb/bolt"
	"github.com/patrickmn/go-cache"
	"github.com/urantiatech/kit/endpoint"
)

// Update - creates a single item
func (s *Service) Update(ctx context.Context, req *api.UpdateRequest, sync bool) (*api.Response, error) {
	var resp = api.Response{Type: req.Type, Language: req.Language}
	var db *bolt.DB
	var err error

	if _, ok := Index[req.Type]; !ok {
		resp.Err = api.ErrorInvalidContentType.Error()
		return &resp, nil
	}

	// Open database in read-write mode
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

		var content map[string]interface{}
		// Get the existing value
		val := bb.Get([]byte(req.Slug))
		if val == nil {
			return api.ErrorNotFound
		}
		err = json.Unmarshal(val, &content)
		if err != nil {
			return err
		}

		// Update values
		if req.Content == nil {
			return api.ErrorNullContent
		}

		var fields = (req.Content).(map[string]interface{})
		for k, v := range fields {

			// Update file
			if strings.HasPrefix(k, "file:") {
				var file item.File
				id := int64(content["id"].(float64))

				if b, err := json.Marshal(v); err != nil {
					return err
				} else if err := json.Unmarshal(b, &file); err != nil {
					return err
				}

				// Update only if new file was uploaded
				if len(file.Bytes) > 0 {
					file.URI = fmt.Sprintf("/drive/%s/%s/%d/%s", req.Type, req.Language, id, file.Name)

					filemap := v.(map[string]interface{})
					filemap["uri"] = file.URI
					filemap["bytes"] = nil

					// Create path
					path := fmt.Sprintf("drive/%s/%s/%d", req.Type, req.Language, id)
					if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
						return err
					}

					// Create file
					dst, err := os.Create(path + "/" + file.Name)
					if err != nil {
						return err
					}
					defer dst.Close()

					// Copy the uploaded file to the destination file
					buff := bytes.NewBuffer(file.Bytes)
					if _, err := io.Copy(dst, buff); err != nil {
						return err
					}
				}

			}

			// Update field
			content[k] = v

		}
		content["updated_at"] = time.Now().Unix()

		// Commit to database
		j, err := json.Marshal(content)
		if err != nil {
			return err
		}
		err = bb.Put([]byte(req.Slug), j)
		if err != nil {
			return err
		}

		resp.Content = content

		index, err := getIndex(req.Type, req.Language)
		if err != nil {
			return err
		}
		err = index.Index(req.Slug, content)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		resp.Err = err.Error()
		return &resp, nil
	}

	// Update the cache
	key := fmt.Sprintf("%s.%s.%s", req.Language, req.Type, req.Slug)
	RespCache.Set(key, &resp, cache.DefaultExpiration)

	return &resp, nil
}

// UpdateEndpoint - creates endpoint for Update service
func UpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(api.UpdateRequest)
		return svc.Update(ctx, &req, false)
	}
}

// DecodeUpdateReq - decodes the incoming request
func DecodeUpdateReq(ctx context.Context, r *http.Request) (interface{}, error) {
	var request api.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}
