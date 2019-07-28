package lightcms

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	s "git.urantiatech.com/cloudcms/lightcms/service"
	"github.com/gorilla/mux"
	h "github.com/urantiatech/kit/transport/http"
	"golang.org/x/text/language"
)

// Languages supported
func Languages(languages []language.Tag) {
	s.Languages = languages
}

// Run method should be called from main function
func Run(port int) {
	// Parse command line parameters
	var dbFile string
	flag.StringVar(&dbFile, "dbFile", "db/cloudcms.db", "The database filename")
	flag.Parse()

	if err := s.Initialize(dbFile); err != nil {
		log.Fatal(err.Error())
	}
	if err := s.RebuildIndex(); err != nil {
		log.Fatal(err.Error())
	}

	var svc s.Service
	svc = s.Service{}

	r := mux.NewRouter()
	r.Handle("/create", h.NewServer(s.CreateEndpoint(svc), s.DecodeCreateReq, s.Encode))
	r.Handle("/read", h.NewServer(s.ReadEndpoint(svc), s.DecodeReadReq, s.Encode))
	r.Handle("/update", h.NewServer(s.UpdateEndpoint(svc), s.DecodeUpdateReq, s.Encode))
	r.Handle("/delete", h.NewServer(s.DeleteEndpoint(svc), s.DecodeDeleteReq, s.Encode))
	r.Handle("/search", h.NewServer(s.SearchEndpoint(svc), s.DecodeSearchReq, s.Encode))
	r.Handle("/facets", h.NewServer(s.FacetsSearchEndpoint(svc), s.DecodeFacetsSearchReq, s.Encode))
	r.Handle("/list", h.NewServer(s.ListEndpoint(svc), s.DecodeListReq, s.Encode))
	r.Handle("/schema", h.NewServer(s.SchemaEndpoint(svc), s.DecodeSchemaReq, s.Encode))

	r.PathPrefix("/drive/").Handler(http.StripPrefix("/drive/", http.FileServer(http.Dir("drive"))))

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
