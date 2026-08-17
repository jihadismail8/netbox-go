// Package initial is the package that starts the service to initialize the service, including
// the initialization configuration, service configuration, connecting to the database, and
// resource release needed when shutting down the service.
package initial

import (
	"flag"
	"strconv"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/stat"
	"github.com/go-dev-frame/sponge/pkg/tracer"

	"netbox-go/configs"
	"netbox-go/internal/config"
	"netbox-go/internal/contenttype"
	"netbox-go/internal/database"
	runtimeconfig "netbox-go/internal/platform/config"
)

var (
	version    string
	configFile string
)

// InitApp initializes the application and returns its validated HTTP runtime
// configuration for explicit service composition.
func InitApp() runtimeconfig.HTTPRuntime {
	return initApp(database.InitDB)
}

func initApp(initDatabase func()) runtimeconfig.HTTPRuntime {
	httpRuntime := initConfig()
	cfg := config.Get()

	// initializing log
	_, err := logger.Init(
		logger.WithLevel(cfg.Logger.Level),
		logger.WithFormat(cfg.Logger.Format),
		logger.WithSave(
			cfg.Logger.IsSave,
			//logger.WithFileName(cfg.Logger.LogFileConfig.Filename),
			//logger.WithFileMaxSize(cfg.Logger.LogFileConfig.MaxSize),
			//logger.WithFileMaxBackups(cfg.Logger.LogFileConfig.MaxBackups),
			//logger.WithFileMaxAge(cfg.Logger.LogFileConfig.MaxAge),
			//logger.WithFileIsCompression(cfg.Logger.LogFileConfig.IsCompression),
		),
	)
	if err != nil {
		panic(err)
	}
	logger.Info("[logger] was initialized")

	// initializing tracing
	if cfg.App.EnableTrace {
		tracer.InitWithConfig(
			cfg.App.Name,
			cfg.App.Env,
			cfg.App.Version,
			cfg.Jaeger.AgentHost,
			strconv.Itoa(cfg.Jaeger.AgentPort),
			cfg.App.TracingSamplingRate,
		)
		logger.Info("[tracer] was initialized")
	}

	// initializing the print system and process resources
	if cfg.App.EnableStat {
		stat.Init(
			stat.WithLog(logger.Get()),
			stat.WithAlarm(), // invalid if it is windows, the default threshold for cpu and memory is 0.8, you can modify them
			stat.WithPrintField(logger.String("service_name", cfg.App.Name), logger.String("host", cfg.App.Host)),
		)
		logger.Info("[resource statistics] was initialized")
	}

	// initializing database
	initDatabase()
	logger.Infof("[%s] was initialized", cfg.Database.Driver)

	// auto-migrate all models (creates tables if they don't exist)
	if err := database.AutoMigrate(); err != nil {
		panic("database bootstrap failed: " + err.Error())
	}

	// Seed django_content_type with every concrete NetBox model. This table
	// powers all GenericFKs (cables, custom fields, tags, change log, ...).
	// Python populates it via `manage.py migrate`; we seed it ourselves.
	if count, err := contenttype.Seed(database.GetDB()); err != nil {
		panic("content-type seed failed: " + err.Error())
	} else {
		logger.Infof("[content-type] %d rows in django_content_type", count)
	}

	database.InitCache(cfg.App.CacheType)
	if cfg.App.CacheType != "" {
		logger.Infof("[%s] was initialized", cfg.App.CacheType)
	}

	return httpRuntime
}

func initConfig() runtimeconfig.HTTPRuntime {
	flag.StringVar(&version, "version", "", "service Version Number")
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.Parse()

	httpRuntime := getConfigFromLocal()

	if version != "" {
		config.Get().App.Version = version
	}

	return httpRuntime
}

// get configuration from local configuration file
func getConfigFromLocal() runtimeconfig.HTTPRuntime {
	if configFile == "" {
		configFile = configs.Location("netbox_go.yml")
	}
	err := config.Init(configFile)
	if err != nil {
		panic("init config error: " + err.Error())
	}
	if err := config.ApplyEnvironmentOverrides(); err != nil {
		panic("invalid environment configuration: " + err.Error())
	}

	httpRuntime, err := runtimeconfig.LoadHTTPRuntimeFromEnvironment()
	if err != nil {
		panic("invalid environment configuration: " + err.Error())
	}
	return httpRuntime
}
