package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KristinaBu/Go_url_shortener/internal/domain"
)

type testLinkService struct {
	createFunc func(ctx context.Context, originalURL string) (domain.Link, error)
	getFunc    func(ctx context.Context, shortCode string) (domain.Link, error)
}

func (s *testLinkService) Create(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	return s.createFunc(ctx, originalURL)
}

func (s *testLinkService) Get(
	ctx context.Context,
	shortCode string,
) (domain.Link, error) {
	return s.getFunc(ctx, shortCode)
}

func TestHandler_CreateLink(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		createFunc func(context.Context, string) (domain.Link, error)
		wantStatus int
		wantURL    string
	}{
		{
			name: "success",
			body: `{"url":"https://example.com"}`,
			createFunc: func(
				_ context.Context,
				originalURL string,
			) (domain.Link, error) {
				return domain.Link{
					ShortCode:   "abc123_XYZ",
					OriginalURL: originalURL,
				}, nil
			},
			wantStatus: http.StatusCreated,
			wantURL:    "/links/abc123_XYZ",
		},
		{
			name: "invalid JSON",
			body: `{"url":`,
			createFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				t.Fatal("Create should not be called")
				return domain.Link{}, nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid URL",
			body: `{"url":"invalid-url"}`,
			createFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				return domain.Link{}, domain.ErrInvalidURL
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal error",
			body: `{"url":"https://example.com"}`,
			createFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				return domain.Link{}, errors.New("database error")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &testLinkService{
				createFunc: tt.createFunc,
			}

			handler := New(service)

			request := httptest.NewRequest(
				http.MethodPost,
				"/links",
				strings.NewReader(tt.body),
			)

			recorder := httptest.NewRecorder()

			handler.CreateLink(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.wantStatus,
				)
			}

			if tt.wantURL == "" {
				return
			}

			var response createLinkResponse

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.ShortURL != tt.wantURL {
				t.Fatalf(
					"short URL = %q, want %q",
					response.ShortURL,
					tt.wantURL,
				)
			}
		})
	}

}

func TestHandler_CreateLinkMethodNotAllowed(t *testing.T) {
	service := &testLinkService{
		createFunc: func(
			_ context.Context,
			_ string,
		) (domain.Link, error) {
			t.Fatal("Create should not be called")
			return domain.Link{}, nil
		},
	}

	handler := New(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/links",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.CreateLink(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status code = %d, want %d",
			recorder.Code,
			http.StatusMethodNotAllowed,
		)
	}

}

func TestHandler_GetLink(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		getFunc func(context.Context, string) (domain.Link, error)
		status  int
		wantURL string
	}{
		{
			name: "success",
			path: "/links/abc123_XYZ",
			getFunc: func(
				_ context.Context,
				shortCode string,
			) (domain.Link, error) {
				if shortCode != "abc123_XYZ" {
					t.Fatalf(
						"short code = %q, want %q",
						shortCode,
						"abc123_XYZ",
					)
				}

				return domain.Link{
					ShortCode:   shortCode,
					OriginalURL: "https://example.com",
				}, nil
			},
			status:  http.StatusOK,
			wantURL: "https://example.com",
		},
		{
			name: "not found",
			path: "/links/unknown",
			getFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				return domain.Link{}, domain.ErrNotFound
			},
			status: http.StatusNotFound,
		},
		{
			name: "invalid path",
			path: "/wrong/abc123_XYZ",
			getFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				t.Fatal("Get should not be called")
				return domain.Link{}, nil
			},
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			path: "/links/abc123_XYZ",
			getFunc: func(
				_ context.Context,
				_ string,
			) (domain.Link, error) {
				return domain.Link{}, errors.New("database error")
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &testLinkService{
				getFunc: tt.getFunc,
			}

			handler := New(service)

			request := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)

			recorder := httptest.NewRecorder()

			handler.GetLink(recorder, request)

			if recorder.Code != tt.status {
				t.Fatalf(
					"status code = %d, want %d",
					recorder.Code,
					tt.status,
				)
			}

			if tt.wantURL == "" {
				return
			}

			var response getLinkResponse

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.URL != tt.wantURL {
				t.Fatalf(
					"url = %q, want %q",
					response.URL,
					tt.wantURL,
				)
			}
		})
	}

}

func TestHandler_GetLinkMethodNotAllowed(t *testing.T) {
	service := &testLinkService{
		getFunc: func(
			_ context.Context,
			_ string,
		) (domain.Link, error) {
			t.Fatal("Get should not be called")
			return domain.Link{}, nil
		},
	}

	handler := New(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/links/abc123_XYZ",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetLink(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status code = %d, want %d",
			recorder.Code,
			http.StatusMethodNotAllowed,
		)
	}

}
