package routers

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter wraps gin.ResponseWriter to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// NetboxEnvelopeMiddleware handles request key conversion (snake -> camel)
// and response formatting (camel -> snake, enveloping lists, unwrapping detail replies).
func NetboxEnvelopeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only intercept requests under /api/
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// 1. Process Request Body (snake_case -> camelCase)
		if c.Request.Body != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil && len(bodyBytes) > 0 {
					var parsed interface{}
					if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
						converted := snakeToCamelContainer(parsed)
						newBodyBytes, err := json.Marshal(converted)
						if err == nil {
							c.Request.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
							c.Request.ContentLength = int64(len(newBodyBytes))
						}
					}
				}
			}
		}

		// 2. Intercept Response Body (camelCase -> snake_case, envelope list, unwrap detail)
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// If response is not JSON or is empty, just write it out
		contentType := blw.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") || blw.body.Len() == 0 {
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// Parse the captured JSON response
		var responseData interface{}
		if err := json.Unmarshal(blw.body.Bytes(), &responseData); err != nil {
			// Write original body if unmarshal fails
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// Format and transform the response
		transformed := transformResponse(responseData, blw.ResponseWriter.Status())

		// Convert keys to snake_case
		transformed = camelToSnakeContainer(transformed)

		// Serialize and write response
		newBytes, err := json.Marshal(transformed)
		if err != nil {
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// Update content-type to application/json and write
		blw.Header().Set("Content-Type", "application/json; charset=utf-8")
		blw.ResponseWriter.Write(newBytes)
	}
}

// transformResponse unwraps single-key replies and formats list envelopes
func transformResponse(data interface{}, status int) interface{} {
	if data == nil {
		return nil
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	// If it's an error response (e.g. has code, msg, and status is not 2xx)
	if status >= 400 {
		if msg, ok := m["msg"].(string); ok && msg != "" {
			return map[string]interface{}{"detail": msg}
		}
		if detail, ok := m["detail"]; ok {
			return map[string]interface{}{"detail": detail}
		}
		return m
	}

	// If it is Sponge's standard envelope: {"code":0, "msg":"...", "data":...}
	if _, hasCode := m["code"]; hasCode {
		if _, hasMsg := m["msg"]; hasMsg {
			if val, hasData := m["data"]; hasData {
				return transformResponse(val, status)
			}
		}
	}

	// If the response is a list reply with total count: {"total": 10, "dcimSites": [...]}
	if total, hasTotal := m["total"]; hasTotal {
		for key, val := range m {
			if key != "total" {
				if arr, isArr := val.([]interface{}); isArr {
					return map[string]interface{}{
						"count":    total,
						"next":     nil,
						"previous": nil,
						"results":  arr,
					}
				}
			}
		}
	}

	// Check if this is a single-key object or array wrapper, e.g. {"dcimSite": {...}} or {"dcimSites": [...]}
	if len(m) == 1 {
		for key, val := range m {
			// Do not unwrap if key is already a standard envelope key
			if key == "id" || key == "count" || key == "results" {
				break
			}
			// If value is a nested object, unwrap it (e.g. GET detail)
			if _, isObj := val.(map[string]interface{}); isObj {
				return val
			}
			// If value is a slice, wrap as list results (e.g. GET list)
			if arr, isArr := val.([]interface{}); isArr {
				return map[string]interface{}{
					"count":    len(arr),
					"next":     nil,
					"previous": nil,
					"results":  arr,
				}
			}
		}
	}

	return m
}

// snakeToCamelContainer converts keys recursively
func snakeToCamelContainer(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		res := make(map[string]interface{}, len(v))
		for k, val := range v {
			res[snakeToCamel(k)] = snakeToCamelContainer(val)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, val := range v {
			res[i] = snakeToCamelContainer(val)
		}
		return res
	default:
		return data
	}
}

// camelToSnakeContainer converts keys recursively
func camelToSnakeContainer(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		res := make(map[string]interface{}, len(v))
		for k, val := range v {
			res[camelToSnake(k)] = camelToSnakeContainer(val)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, val := range v {
			res[i] = camelToSnakeContainer(val)
		}
		return res
	default:
		return data
	}
}

func camelToSnake(s string) string {
	if strings.HasSuffix(s, "ID") {
		return camelToSnake(s[:len(s)-2]) + "_id"
	}
	var res []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res = append(res, '_')
		}
		res = append(res, unicode.ToLower(r))
	}
	return string(res)
}

func snakeToCamel(s string) string {
	if s == "id" {
		return "id"
	}
	if s == "_abs_distance" {
		return "xAbsDistance"
	}
	if strings.HasSuffix(s, "_id") {
		return snakeToCamel(s[:len(s)-3]) + "ID"
	}
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
