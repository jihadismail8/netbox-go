// Package routers provides NetBox custom @action endpoints that do not fit the
// generic DRF model registry: system status, data-source sync, IPAM
// availability, and cable tracing.
//
// These endpoints are plain Gin routes (no proto/codegen), registered alongside
// the DRF routes so they inherit the same auth + envelope middleware.
package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"netbox-go/internal/database"
	"netbox-go/internal/routers/dcim"
	"netbox-go/internal/routers/ipam"
)

// netboxVersion is the version string surfaced by GET /api/status/.
// It intentionally mirrors a Python NetBox release so API consumers that
// gate on version numbers (e.g. netbox-python-client) keep working.
const netboxVersion = "4.4.0-go"

// installedApps is the list of NetBox "apps" reported by GET /api/status/.
var installedApps = []string{
	"dcim", "ipam", "circuits", "tenancy", "virtualization",
	"vpn", "wireless", "extras", "users", "core",
}

// registerCustomActionRoutes registers NetBox custom @action endpoints on the
// given router group. The group is expected to already carry auth + envelope
// middleware (it is the same apiGroup the DRF routes register on).
//
// Note: the public system routes (/api/status, /api/) are registered separately
// on the bare engine by registerPublicCustomRoutes — do NOT re-register them
// here, or Gin will panic on the duplicate path.
func registerCustomActionRoutes(r gin.IRouter) {
	// --- Data source sync (truthful "queued" stub; real ingestion deferred to asynq worker) ---
	r.POST("/api/core/data-sources/:id/sync", makeDataSourceSyncHandler())

	// --- IPAM availability endpoints (see ipam package) ---
	registerIPAMAvailableRoutes(r)

	// --- Cable trace / paths (see dcim package) ---
	registerTraceRoutes(r)

	// --- Connection lists (console / power / interface) ---
	registerConnectionRoutes(r)
}

// registerPublicCustomRoutes registers the small set of custom endpoints that
// must be reachable without authentication (status). Called on the bare engine.
func registerPublicCustomRoutes(r *gin.Engine) {
	registerSystemRoutes(r)
}

// ---- System status ----

// registerSystemRoutes registers GET /api/status/ and GET /api/.
func registerSystemRoutes(r gin.IRouter) {
	r.GET("/api/status", makeStatusHandler())
	r.GET("/api/", makeAPIRootHandler())
}

// makeStatusHandler returns NetBox's system status payload.
// DB liveness is probed but never fails the request (mirrors NetBox).
func makeStatusHandler() gin.HandlerFunc {
	hostname, _ := os.Hostname()
	return func(c *gin.Context) {
		dbStatus := "ok"
		if db := database.GetDB(); db != nil {
			var n int64
			// COUNT on django_content_type — a table that always exists.
			if err := db.Table("django_content_type").Limit(1).Count(&n).Error; err != nil {
				dbStatus = "error: " + err.Error()
			}
		} else {
			dbStatus = "not initialized"
		}

		c.JSON(http.StatusOK, gin.H{
			"netbox-version":     netboxVersion,
			"python-version":     nil,
			"go-version":         runtime.Version(),
			"django-version":     nil,
			"hostname":           hostname,
			"installed-apps":     installedApps,
			"plugins":            map[string]interface{}{},
			"rq-workers-running": 0,
			"database-status":    dbStatus,
		})
	}
}

// makeAPIRootHandler returns links to each module root, like NetBox's API root.
func makeAPIRootHandler() gin.HandlerFunc {
	modules := []string{
		"circuits", "core", "dcim", "extras", "ipam",
		"tenancy", "users", "virtualization", "vpn", "wireless",
	}
	return func(c *gin.Context) {
		base := schemeFromRequest(c) + "://" + c.Request.Host + "/api/"
		links := make(map[string]string, len(modules))
		for _, m := range modules {
			links[m] = base + m + "/"
		}
		c.JSON(http.StatusOK, links)
	}
}

// ---- Data source sync ----

// makeDataSourceSyncHandler marks a data source as "syncing" and returns 202.
//
// NOTE: NetBox performs the actual ingestion (git/scp/local fetch + file
// indexing) via a background job queue (RQ in Python, asynq in Go). The asynq
// worker is not yet wired up, so this handler only records the intent. This is
// a truthful "queued" response — it does not pretend the sync succeeded.
func makeDataSourceSyncHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		// Validate the data source exists.
		var exists int64
		if err := db.Table("core_datasource").Where("id = ?", id).Count(&exists).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if exists == 0 {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}

		// Record the sync attempt. last_synced is stamped now; status flips to
		// "syncing" and would be updated to "completed"/"failed" by the worker.
		now := time.Now()
		updates := map[string]interface{}{
			"status":      "syncing",
			"last_synced": now,
		}
		if err := db.Table("core_datasource").Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status": "queued",
			"detail": "Sync initiated. Actual ingestion is handled by an asynq worker (TODO).",
		})
	}
}

// ---- Delegation to dcim subpackage: trace & paths ----

// registerTraceRoutes wires the cable trace and paths endpoints onto the
// router group. The dcim package holds the actual resolution logic.
func registerTraceRoutes(r gin.IRouter) {
	traceSvc := dcim.NewTraceService()
	for _, model := range dcim.PathEndpointModels {
		r.GET("/api/dcim/"+model+"/:id/trace", makeTraceHandler(traceSvc, model))
	}
	for _, model := range dcim.PassThroughModels {
		r.GET("/api/dcim/"+model+"/:id/paths", makePathsHandler(traceSvc, model))
	}
}

// registerConnectionRoutes wires the three read-only connection-list endpoints
// (console / power / interface connections). The dcim package builds each row
// by joining through dcim_cabletermination.
func registerConnectionRoutes(r gin.IRouter) {
	for _, key := range []string{"console-connections", "power-connections", "interface-connections"} {
		ct, ok := dcim.ConnectionTypes[key]
		if !ok {
			continue
		}
		r.GET("/api/dcim/"+key, dcim.MakeConnectionsHandler(ct))
	}
}

func makeTraceHandler(svc *dcim.TraceService, model string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ?render=svg is not implemented (frontend never requests it).
		if c.Query("render") == "svg" {
			c.JSON(http.StatusNotImplemented, gin.H{
				"detail": "SVG rendering is not supported; omit ?render=svg or use the JSON response.",
			})
			return
		}
		id := c.Param("id")
		result, err := svc.Trace(c.Request.Context(), database.GetDB(), model, id)
		if err != nil {
			status := http.StatusInternalServerError
			if err == dcim.ErrNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func makePathsHandler(svc *dcim.TraceService, model string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := svc.Paths(c.Request.Context(), database.GetDB(), model, id)
		if err != nil {
			status := http.StatusInternalServerError
			if err == dcim.ErrNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// ---- Delegation to ipam subpackage: available IPs/prefixes/ASNs ----

func registerIPAMAvailableRoutes(r gin.IRouter) {
	svc := ipam.NewAvailableService()
	r.GET("/api/ipam/prefixes/:id/available-ips", svc.HandleGetAvailableIPs("prefix"))
	r.POST("/api/ipam/prefixes/:id/available-ips", svc.HandlePostAvailableIPs("prefix"))
	r.GET("/api/ipam/ip-ranges/:id/available-ips", svc.HandleGetAvailableIPs("range"))
	r.POST("/api/ipam/ip-ranges/:id/available-ips", svc.HandlePostAvailableIPs("range"))
	r.GET("/api/ipam/prefixes/:id/available-prefixes", svc.HandleGetAvailablePrefixes())
	r.POST("/api/ipam/prefixes/:id/available-prefixes", svc.HandlePostAvailablePrefixes())
	r.GET("/api/ipam/asn-ranges/:id/available-asns", svc.HandleGetAvailableASNs())
	r.POST("/api/ipam/asn-ranges/:id/available-asns", svc.HandlePostAvailableASNs())
}

// suppress unused-import warnings until json/datatypes are used by ipam wrapper
var _ = json.Marshal
var _ = datatypes.JSON{}
var _ = fmt.Sprintf
