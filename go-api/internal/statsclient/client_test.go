package statsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComputeForwardsAuthAndDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stats" || r.Header.Get("Authorization") != "Bearer abc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body struct{ Matrices [][][]float64 }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Matrices) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"max":3,"min":-5,"average":-0.95,"sum":-7.6,"anyDiagonal":false,"diagonal":[false,false]}`))
	}))
	defer srv.Close()

	s, err := New(srv.URL).Compute(context.Background(), "Bearer abc", [][]float64{{1}}, [][]float64{{2}})
	if err != nil {
		t.Fatal(err)
	}
	if s.Max != 3 || s.Min != -5 || s.Sum != -7.6 || s.AnyDiagonal || len(s.Diagonal) != 2 {
		t.Fatalf("unexpected stats %+v", s)
	}
}

func TestComputeNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	if _, err := New(srv.URL).Compute(context.Background(), "", [][]float64{{1}}); err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if err := New(srv.URL).Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	srv.Close()
	if err := New(srv.URL).Ping(context.Background()); err == nil {
		t.Fatal("expected error when server is down")
	}
}
