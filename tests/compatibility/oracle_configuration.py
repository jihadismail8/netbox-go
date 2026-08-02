"""Closed, test-only configuration for the pinned NetBox compatibility oracle."""

import os


ALLOWED_HOSTS = ["*"]
SECRET_KEY = os.environ["SECRET_KEY"]

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": os.environ["DB_NAME"],
        "USER": os.environ["DB_USER"],
        "PASSWORD": os.environ["DB_PASSWORD"],
        "HOST": os.environ["DB_HOST"],
        "PORT": int(os.getenv("DB_PORT", "5432")),
        "CONN_MAX_AGE": 0,
    }
}

REDIS = {
    "tasks": {
        "HOST": os.environ["REDIS_HOST"],
        "PORT": int(os.getenv("REDIS_PORT", "6379")),
        "USERNAME": "",
        "PASSWORD": "",
        "DATABASE": int(os.getenv("REDIS_DATABASE", "0")),
        "SSL": False,
    },
    "caching": {
        "HOST": os.environ["REDIS_HOST"],
        "PORT": int(os.getenv("REDIS_PORT", "6379")),
        "USERNAME": "",
        "PASSWORD": "",
        "DATABASE": int(os.getenv("REDIS_CACHE_DATABASE", "1")),
        "SSL": False,
    },
}

LOGIN_REQUIRED = True
ALLOW_TOKEN_RETRIEVAL = False
MAINTENANCE_MODE = False
ENFORCE_GLOBAL_UNIQUE = True
TIME_ZONE = "UTC"
CENSUS_REPORTING_ENABLED = False
METRICS_ENABLED = False
GRAPHQL_ENABLED = False
