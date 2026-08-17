package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTransformResponsePreservesNetBoxEnvelopeSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   interface{}
		status int
		want   interface{}
	}{
		{
			name:   "nil",
			data:   nil,
			status: http.StatusOK,
			want:   nil,
		},
		{
			name:   "non object response",
			data:   []interface{}{"unchanged"},
			status: http.StatusOK,
			want:   []interface{}{"unchanged"},
		},
		{
			name:   "error message becomes detail",
			data:   map[string]interface{}{"code": float64(7), "msg": "invalid input"},
			status: http.StatusBadRequest,
			want:   map[string]interface{}{"detail": "invalid input"},
		},
		{
			name:   "existing error detail is retained",
			data:   map[string]interface{}{"detail": "not found"},
			status: http.StatusNotFound,
			want:   map[string]interface{}{"detail": "not found"},
		},
		{
			name:   "unrecognized error is not discarded",
			data:   map[string]interface{}{"errors": []interface{}{"conflict"}},
			status: http.StatusConflict,
			want:   map[string]interface{}{"errors": []interface{}{"conflict"}},
		},
		{
			name: "sponge envelope is recursively removed",
			data: map[string]interface{}{
				"code": float64(0),
				"msg":  "ok",
				"data": map[string]interface{}{
					"dcimSite": map[string]interface{}{"id": float64(3), "name": "AMS"},
				},
			},
			status: http.StatusOK,
			want:   map[string]interface{}{"id": float64(3), "name": "AMS"},
		},
		{
			name: "total and list become a paginated result",
			data: map[string]interface{}{
				"total":     float64(2),
				"dcimSites": []interface{}{map[string]interface{}{"id": float64(1)}},
			},
			status: http.StatusOK,
			want: map[string]interface{}{
				"count":    float64(2),
				"next":     nil,
				"previous": nil,
				"results":  []interface{}{map[string]interface{}{"id": float64(1)}},
			},
		},
		{
			name:   "single object wrapper is removed",
			data:   map[string]interface{}{"dcimSite": map[string]interface{}{"id": float64(4)}},
			status: http.StatusOK,
			want:   map[string]interface{}{"id": float64(4)},
		},
		{
			name:   "single list wrapper becomes a paginated result",
			data:   map[string]interface{}{"dcimSites": []interface{}{map[string]interface{}{"id": float64(5)}}},
			status: http.StatusOK,
			want: map[string]interface{}{
				"count":    1,
				"next":     nil,
				"previous": nil,
				"results":  []interface{}{map[string]interface{}{"id": float64(5)}},
			},
		},
		{
			name:   "standard single key is not unwrapped",
			data:   map[string]interface{}{"id": float64(6)},
			status: http.StatusOK,
			want:   map[string]interface{}{"id": float64(6)},
		},
		{
			name: "ordinary object is unchanged",
			data: map[string]interface{}{
				"id":   float64(7),
				"name": "LON",
			},
			status: http.StatusOK,
			want: map[string]interface{}{
				"id":   float64(7),
				"name": "LON",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := transformResponse(tt.data, tt.status)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("transformResponse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestEnvelopeKeyConversionIsRecursive(t *testing.T) {
	t.Parallel()

	snake := map[string]interface{}{
		"site_name": "AMS",
		"region_id": float64(2),
		"nested_items": []interface{}{
			map[string]interface{}{"rack_id": float64(9)},
		},
	}
	wantCamel := map[string]interface{}{
		"siteName": "AMS",
		"regionID": float64(2),
		"nestedItems": []interface{}{
			map[string]interface{}{"rackID": float64(9)},
		},
	}
	if got := snakeToCamelContainer(snake); !reflect.DeepEqual(got, wantCamel) {
		t.Fatalf("snakeToCamelContainer() = %#v, want %#v", got, wantCamel)
	}

	wantSnake := map[string]interface{}{
		"site_name": "AMS",
		"region_id": float64(2),
		"nested_items": []interface{}{
			map[string]interface{}{"rack_id": float64(9)},
		},
	}
	if got := camelToSnakeContainer(wantCamel); !reflect.DeepEqual(got, wantSnake) {
		t.Fatalf("camelToSnakeContainer() = %#v, want %#v", got, wantSnake)
	}

	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{name: "id remains id", got: snakeToCamel("id"), want: "id"},
		{name: "foreign key suffix", got: snakeToCamel("assigned_object_id"), want: "assignedObjectID"},
		{name: "legacy distance field", got: snakeToCamel("_abs_distance"), want: "xAbsDistance"},
		{name: "empty segment", got: snakeToCamel("site__name"), want: "siteName"},
		{name: "camel foreign key suffix", got: camelToSnake("assignedObjectID"), want: "assigned_object_id"},
		{name: "ordinary camel key", got: camelToSnake("lastUpdated"), want: "last_updated"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Fatalf("conversion = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestNetboxEnvelopeMiddlewareConvertsJSONRequestAndResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NetboxEnvelopeMiddleware())

	var requestBody map[string]interface{}
	var contentLength int64
	router.POST("/api/dcim/sites", func(c *gin.Context) {
		contentLength = c.Request.ContentLength
		if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode converted request: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "ok",
			"data": gin.H{
				"total": 1,
				"dcimSites": []gin.H{
					{"siteID": 4, "lastUpdated": "2026-08-02T00:00:00Z"},
				},
			},
		})
	})

	body := []byte(`{"site_name":"AMS","region_id":2,"nested_items":[{"rack_id":9}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/dcim/sites", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	wantRequest := map[string]interface{}{
		"siteName": "AMS",
		"regionID": float64(2),
		"nestedItems": []interface{}{
			map[string]interface{}{"rackID": float64(9)},
		},
	}
	if !reflect.DeepEqual(requestBody, wantRequest) {
		t.Fatalf("converted request = %#v, want %#v", requestBody, wantRequest)
	}
	if contentLength <= 0 {
		t.Fatalf("converted request content length = %d, want positive", contentLength)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", got)
	}

	var gotResponse interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &gotResponse); err != nil {
		t.Fatalf("decode transformed response: %v", err)
	}
	wantResponse := map[string]interface{}{
		"count":    float64(1),
		"next":     nil,
		"previous": nil,
		"results": []interface{}{
			map[string]interface{}{
				"site_id":      float64(4),
				"last_updated": "2026-08-02T00:00:00Z",
			},
		},
	}
	if !reflect.DeepEqual(gotResponse, wantResponse) {
		t.Fatalf("transformed response = %#v, want %#v", gotResponse, wantResponse)
	}
}

func TestNetboxEnvelopeMiddlewareResponseFallbacks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("outside API is untouched", func(t *testing.T) {
		router := gin.New()
		router.Use(NetboxEnvelopeMiddleware())
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"camelKey": "unchanged"})
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
		if got := response.Body.String(); got != `{"camelKey":"unchanged"}` {
			t.Fatalf("body = %q, want untouched JSON", got)
		}
	})

	t.Run("plain text is copied through", func(t *testing.T) {
		router := gin.New()
		router.Use(NetboxEnvelopeMiddleware())
		router.GET("/api/plain", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/plain", []byte("plain response"))
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/plain", nil))
		if got := response.Body.String(); got != "plain response" {
			t.Fatalf("body = %q, want plain response", got)
		}
	})

	t.Run("malformed JSON response is copied through", func(t *testing.T) {
		router := gin.New()
		router.Use(NetboxEnvelopeMiddleware())
		router.GET("/api/broken", func(c *gin.Context) {
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusOK)
			_, _ = c.Writer.WriteString("{broken")
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/broken", nil))
		if got := response.Body.String(); got != "{broken" {
			t.Fatalf("body = %q, want original malformed JSON", got)
		}
	})

	t.Run("empty JSON response remains empty", func(t *testing.T) {
		router := gin.New()
		router.Use(NetboxEnvelopeMiddleware())
		router.DELETE("/api/items/:id", func(c *gin.Context) {
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusNoContent)
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/items/1", nil))
		if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
			t.Fatalf("status/body = %d/%q, want 204 with empty body", response.Code, response.Body.String())
		}
	})

	t.Run("error message becomes NetBox detail", func(t *testing.T) {
		router := gin.New()
		router.Use(NetboxEnvelopeMiddleware())
		router.POST("/api/items", func(c *gin.Context) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 7, "msg": "invalid field"})
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/items", nil))
		if got := response.Body.String(); got != `{"detail":"invalid field"}` {
			t.Fatalf("body = %q, want NetBox detail response", got)
		}
	})
}
