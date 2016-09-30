package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/coditect/transloc-coding-exercise/model"
)

type Server struct {
	locations model.LocationStorage
	static    http.Handler
}

func NewServer(storage model.LocationStorage, rootDir string) *Server {
	return &Server{
		locations: storage,
		static:    http.FileServer(http.Dir(rootDir)),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/geoip" {
		var err error

		switch r.Method {
		case "GET", "HEAD":
			err = s.Get(w, r)
		case "POST":
			err = s.Post(w, r)
		default:
			err = s.MethodNotAllowed(w, r)
		}

		if err != nil {
			status := 500
			if httpError, ok := err.(model.HTTPError); ok {
				status = httpError.HTTPStatusCode
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(status)
			fmt.Fprintln(w, err)
		}

	} else {
		s.static.ServeHTTP(w, r)
	}
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) error {
	north, err := ParseFloatParameter(r, "north", true)
	if err != nil {
		return err
	}

	south, err := ParseFloatParameter(r, "south", true)
	if err != nil {
		return err
	}

	east, err := ParseFloatParameter(r, "east", true)
	if err != nil {
		return err
	}

	west, err := ParseFloatParameter(r, "west", true)
	if err != nil {
		return err
	}

	resolution, err := ParseFloatParameter(r, "resolution", false)
	if err != nil {
		return err
	}

	results, err := s.locations.Query(north, south, east, west)
	if err != nil {
		return model.HTTPError{err, 500}
	}

	if resolution > 0 {
		results = results.RoundLocations(resolution, resolution)
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(results.Logarithmic(10))
}

func (s *Server) Post(w http.ResponseWriter, r *http.Request) error {
	var reader io.Reader
	contentType := r.Header.Get("Content-Type")
	contentTypeParts := strings.SplitN(contentType, ";", 2)

	switch contentTypeParts[0] {
	case "text/csv":
		reader = r.Body
	case "multipart/form-data":
		file, _, err := r.FormFile("file")
		if err != nil {
			return err
		}
		reader = file
	default:
		return model.HTTPError{fmt.Errorf("Cannot accept requests of type %s", contentTypeParts[0]), 415}
	}

	locations, err := model.LocationTableFromCSV(reader)
	if err != nil {
		return err
	}

	return s.locations.Save(locations)
}

func (s *Server) MethodNotAllowed(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain")
	return model.HTTPError{fmt.Errorf("Method %s is not allowed", r.Method), 405}
}

func ParseFloatParameter(r *http.Request, name string, required bool) (float64, error) {
	raw := r.FormValue(name)
	if raw == "" {
		if required {
			return 0, model.HTTPError{fmt.Errorf("Missing required parameter %q", name), 400}
		}
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, model.HTTPError{err, 400}
	}
	return parsed, nil
}
