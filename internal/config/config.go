package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL  string
	Port         string
	CookieDomain string
	CookieSecure bool
	CORSOrigin   string
	S3Endpoint   string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string

	// OAuth
	GoogleClientID     string
	GoogleClientSecret string
	FacebookAppID      string
	FacebookAppSecret  string
	OAuthRedirectBase  string
	FrontendURL        string
}

func Load() (*Config, error) {
	c := &Config{}

	var err error
	if c.DatabaseURL, err = require("DATABASE_URL"); err != nil {
		return nil, err
	}
	if c.CookieDomain, err = require("COOKIE_DOMAIN"); err != nil {
		return nil, err
	}
	if c.CORSOrigin, err = require("CORS_ORIGIN"); err != nil {
		return nil, err
	}
	if c.S3Endpoint, err = require("S3_ENDPOINT"); err != nil {
		return nil, err
	}
	if c.S3Bucket, err = require("S3_BUCKET"); err != nil {
		return nil, err
	}
	if c.S3AccessKey, err = require("S3_ACCESS_KEY"); err != nil {
		return nil, err
	}
	if c.S3SecretKey, err = require("S3_SECRET_KEY"); err != nil {
		return nil, err
	}

	if c.OAuthRedirectBase, err = require("OAUTH_REDIRECT_BASE"); err != nil {
		return nil, err
	}
	if c.FrontendURL, err = require("FRONTEND_URL"); err != nil {
		return nil, err
	}

	// OAuth provider credentials are optional; missing = provider disabled.
	c.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	c.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	c.FacebookAppID = os.Getenv("FACEBOOK_APP_ID")
	c.FacebookAppSecret = os.Getenv("FACEBOOK_APP_SECRET")

	c.CookieSecure = os.Getenv("COOKIE_SECURE") == "true"
	if c.Port = os.Getenv("PORT"); c.Port == "" {
		c.Port = "8080"
	}

	return c, nil
}

func require(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("required env var %s is not set", key)
}
