package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabrielalc23/pdv/config"
	"github.com/gabrielalc23/pdv/internal/app"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/platform/valkey"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer store.Close()

	var vk *valkey.Client
	if cfg.ValkeyAddr != "" {
		var err error
		vk, err = valkey.NewClient(valkey.Config{
			Addr:     cfg.ValkeyAddr,
			Password: cfg.ValkeyPassword,
			DB:       cfg.ValkeyDB,
		})
		if err != nil {
			log.Fatalf("valkey: %v", err)
		}
		defer vk.Close()
	}

	hasher, err := password.NewHasher(password.Params{
		MemoryKiB:   cfg.PasswordArgon2MemoryKiB,
		Iterations:  cfg.PasswordArgon2Iterations,
		Parallelism: cfg.PasswordArgon2Parallelism,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		log.Fatalf("password hasher: %v", err)
	}

	passwordPolicy := password.DefaultPolicy()

	var passwordBlocklist password.Blocklist
	if cfg.PasswordBlocklistPath != "" {
		var err error
		passwordBlocklist, err = password.NewBlocklistFromFile(cfg.PasswordBlocklistPath)
		if err != nil {
			log.Fatalf("password blocklist: %v", err)
		}
	} else {
		passwordBlocklist = password.NewBuiltinBlocklist()
	}

	var keyring *jwt.Keyring
	if cfg.JWTPrivateKeyPath != "" {
		keyring, err = jwt.LoadKeyring(cfg.JWTActiveKeyID, cfg.JWTPrivateKeyPath, cfg.JWTPublicKeysDir)
		if err != nil {
			log.Fatalf("jwt keyring: %v", err)
		}
	} else if cfg.AuthAllowEphemeralDevKey {
		kid := cfg.JWTActiveKeyID
		if kid == "" {
			kid = "dev-key"
		}
		keyring, err = jwt.NewEphemeralKeyring(kid)
		if err != nil {
			log.Fatalf("jwt ephemeral keyring: %v", err)
		}
	} else {
		log.Fatal("jwt: JWT_PRIVATE_KEY_PATH or AUTH_ALLOW_EPHEMERAL_DEV_KEY=true required")
	}

	cookieManager, err := cookie.NewManager(cookie.Config{
		Secure:      cfg.CookieSecure,
		SameSite:    cfg.CookieSameSite,
		RefreshName: cfg.CookieRefreshName,
		CSRFName:    cfg.CookieCSRFName,
		Env:         cfg.AppEnv,
	})
	if err != nil {
		log.Fatalf("cookie manager: %v", err)
	}

	csrfManager, err := csrf.NewManager(cfg.AuthCSRFSecret, cfg.CORSAllowedOrigins)
	if err != nil {
		log.Fatalf("csrf manager: %v", err)
	}

	requestMeta, err := requestmeta.NewResolver(cfg.TrustedProxyCIDRs)
	if err != nil {
		log.Fatalf("request meta resolver: %v", err)
	}

	rateLimiter := ratelimit.NewFallbackLimiter(10000)

	handler := app.New(app.Dependencies{
		Store:             store,
		Valkey:            vk,
		PasswordHasher:    hasher,
		PasswordPolicy:    passwordPolicy,
		PasswordBlocklist: passwordBlocklist,
		CookieManager:     cookieManager,
		CSRFManager:       csrfManager,
		JWTKeyring:        keyring,
		JWTIssuer:         cfg.JWTIssuer,
		JWTAudience:       cfg.JWTAudience,
		AccessTokenTTL:    cfg.AccessTokenTTL,
		JWTClockSkew:      cfg.JWTClockSkew,
		RequestMeta:       requestMeta,
		RateLimiter:       rateLimiter,
	})

	server := app.NewHTTPServer(cfg.Address, handler)

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
	}()

	slog.Info(fmt.Sprintf("listening on %s", cfg.Address))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
