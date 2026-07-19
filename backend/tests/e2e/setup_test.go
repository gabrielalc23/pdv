package e2e

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/app"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/mailer"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/platform/valkey"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	baseURL          string
	testOrgID        string
	testStoreID      string
	testMembershipID string
	testPool         *pgxpool.Pool
	testStore        *database.PostgresStore
	testValkey       *valkey.Client
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://pdv:pdv@localhost:5433/pdv?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6380"
	}
	if !tryConnect(ctx, dsn) || !tryValkey(ctx, valkeyAddr) {
		startDeps(ctx, dsn)
	}

	pool := mustConnect(ctx, dsn)
	defer pool.Close()

	// create separate test database and reconnect to it
	testDBName := "pdv_e2e_test"
	createTestDB(ctx, pool, testDBName)
	pool.Close()

	testDSN := replaceDBName(dsn, testDBName)
	testPool = mustConnect(ctx, testDSN)
	defer testPool.Close()

	runMigrations(ctx, testPool)
	testOrgID, testStoreID, testMembershipID = seedTenantData(ctx, testPool)
	seedPaymentMethods(ctx, testPool, testOrgID, testStoreID)

	store := database.NewStore(testPool)
	testStore = store
	lis := mustListen()
	baseURL = "http://" + lis.Addr().String()
	vk, err := valkey.NewClient(valkey.Config{Addr: valkeyAddr, DB: 15})
	if err != nil {
		log.Fatalf("valkey: %v", err)
	}
	defer vk.Close()
	testValkey = vk
	if _, err := vk.Do(ctx, vk.B().Flushdb().Build()); err != nil {
		log.Fatalf("flush isolated e2e valkey database: %v", err)
	}
	server := startServer(store, vk, lis.Addr().String())

	go func() {
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	waitForReady(ctx)

	code := m.Run()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)

	os.Exit(code)
}

func replaceDBName(dsn, newName string) string {
	// postgres://user:pass@host:port/dbname?params => replace dbname
	idx := strings.LastIndex(dsn, "/")
	if idx < 0 {
		return dsn
	}
	rest := dsn[idx+1:]
	idx2 := strings.Index(rest, "?")
	if idx2 < 0 {
		return dsn[:idx+1] + newName
	}
	return dsn[:idx+1] + newName + rest[idx2:]
}

func createTestDB(ctx context.Context, pool *pgxpool.Pool, name string) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	// terminate existing connections to the test DB if it exists
	_, _ = conn.Exec(ctx, `
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = $1 AND pid <> pg_backend_pid()
	`, name)

	_, err = conn.Exec(ctx, `DROP DATABASE IF EXISTS `+name)
	if err != nil {
		log.Fatalf("drop test db: %v", err)
	}
	_, err = conn.Exec(ctx, `CREATE DATABASE `+name)
	if err != nil {
		log.Fatalf("create test db: %v", err)
	}
}

func tryConnect(ctx context.Context, dsn string) bool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return false
	}
	defer pool.Close()
	return pool.Ping(ctx) == nil
}

func startDeps(ctx context.Context, dsn string) {
	cmd := exec.CommandContext(ctx,
		"docker", "compose",
		"-f", "../../docker-compose.test.yml",
		"up", "-d",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("docker compose up failed: %v\n%s", err, out)
	}

	for range 30 {
		if tryConnect(ctx, dsn) {
			return
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Fatal("dependency startup cancelled")
		case <-timer.C:
		}
	}
	log.Fatal("postgres not ready after 30s")
}

func tryValkey(ctx context.Context, addr string) bool {
	client, err := valkey.NewClient(valkey.Config{Addr: addr, DB: 15})
	if err != nil {
		return false
	}
	defer client.Close()
	return client.Ping(ctx) == nil
}

func mustConnect(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}
	return pool
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `DROP SCHEMA public CASCADE`)
	if err != nil {
		log.Fatalf("drop schema: %v", err)
	}
	_, err = conn.Exec(ctx, `CREATE SCHEMA public`)
	if err != nil {
		log.Fatalf("create schema: %v", err)
	}

	_, err = conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pg_trgm`)
	if err != nil {
		log.Fatalf("create extension pg_trgm: %v", err)
	}

	_, err = conn.Exec(ctx, `
		CREATE OR REPLACE FUNCTION uuidv7() RETURNS uuid
		LANGUAGE sql STABLE
		AS $$ SELECT gen_random_uuid() $$
	`)
	if err != nil {
		log.Fatalf("create uuidv7: %v", err)
	}

	matches, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		log.Fatalf("glob migrations: %v", err)
	}
	sort.Strings(matches)

	for _, path := range matches {
		sql, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", path, err)
		}
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("exec %s: %v\n%s", path, err, sql)
		}
	}
}

func seedPaymentMethods(ctx context.Context, pool *pgxpool.Pool, orgID, storeID string) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	type methodDef struct {
		code, name, kind string
		allowsChange     bool
		allowsInstall    bool
		maxInstall       int16
		sortOrder        int
	}
	methods := []methodDef{
		{"CASH", "Dinheiro", "CASH", true, false, 1, 1},
		{"PIX", "PIX", "PIX", false, false, 1, 2},
		{"DEBIT", "Cartão de Débito", "DEBIT_CARD", false, false, 1, 3},
		{"CREDIT", "Cartão de Crédito", "CREDIT_CARD", false, true, 12, 4},
		{"VOUCHER", "Vale", "VOUCHER", false, false, 1, 5},
	}

	now := time.Now()
	for _, m := range methods {
		var mid string
		err := conn.QueryRow(ctx, `
			INSERT INTO payment_methods (organization_id, code, name, kind, allows_change, allows_installments, max_installments, sort_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (organization_id, code) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, orgID, m.code, m.name, m.kind, m.allowsChange, m.allowsInstall, m.maxInstall, m.sortOrder, now, now).Scan(&mid)
		if err != nil {
			log.Fatalf("seed payment method %s: %v", m.code, err)
		}
		if mid == "" {
			continue
		}
		_, err = conn.Exec(ctx, `
			INSERT INTO store_payment_methods (organization_id, store_id, payment_method_id, is_active, sort_order, created_at, updated_at)
			VALUES ($1, $2, $3, TRUE, $4, $5, $6)
			ON CONFLICT (organization_id, store_id, payment_method_id) DO NOTHING
		`, orgID, storeID, mid, m.sortOrder, now, now)
		if err != nil {
			log.Fatalf("link payment method %s to store: %v", m.code, err)
		}
	}
}

func seedTenantData(ctx context.Context, pool *pgxpool.Pool) (orgID, storeID, membershipID string) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	var userID string
	err = conn.QueryRow(ctx, `
		INSERT INTO users (email, email_normalized, display_name)
		VALUES ('e2e@test.local', 'e2e@test.local', 'E2E Test')
		ON CONFLICT (email_normalized) DO UPDATE SET email = users.email
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, created_by_user_id)
		VALUES ('E2E Test Org', 'e2e-test-org', $1)
		ON CONFLICT (slug) DO UPDATE SET name = organizations.name
		RETURNING id
	`, userID).Scan(&orgID)
	if err != nil {
		log.Fatalf("insert organization: %v", err)
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO stores (organization_id, code, name, timezone, created_by_user_id)
		VALUES ($1, 'E2E', 'E2E Store', 'America/Sao_Paulo', $2)
		ON CONFLICT (organization_id, code) DO UPDATE SET name = stores.name
		RETURNING id
	`, orgID, userID).Scan(&storeID)
	if err != nil {
		log.Fatalf("insert store: %v", err)
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO organization_memberships (organization_id, user_id, created_by_user_id)
		VALUES ($1, $2, $2)
		ON CONFLICT (organization_id, user_id) WHERE status <> 'REMOVED' DO UPDATE SET user_id = organization_memberships.user_id
		RETURNING id
	`, orgID, userID).Scan(&membershipID)
	if err != nil {
		log.Fatalf("insert membership: %v", err)
	}

	return
}

func mustListen() net.Listener {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	return lis
}

func startServer(store *database.PostgresStore, vk *valkey.Client, addr string) *http.Server {
	logMailer, err := mailer.NewLogMailer("test")
	if err != nil {
		log.Fatalf("test log mailer: %v", err)
	}
	handler := app.New(testDependencies(store, vk, "http://"+addr, true, false, logMailer))
	return app.NewHTTPServer(addr, handler)
}

func testDependencies(store *database.PostgresStore, vk *valkey.Client, publicURL string, registrationEnabled, requireVerified bool, authMailer mailer.Mailer) app.Dependencies {
	hasher, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		log.Fatalf("password hasher: %v", err)
	}
	keyring, err := jwt.NewEphemeralKeyring("e2e-key")
	if err != nil {
		log.Fatalf("keyring: %v", err)
	}
	cookies, err := cookie.NewManager(cookie.Config{Secure: false, SameSite: "Lax", RefreshName: "pdv_refresh", CSRFName: "pdv_csrf", Env: "test"})
	if err != nil {
		log.Fatalf("cookies: %v", err)
	}
	csrfManager, err := csrf.NewManager([]byte("0123456789abcdef0123456789abcdef"), []string{publicURL})
	if err != nil {
		log.Fatalf("csrf: %v", err)
	}
	meta, err := requestmeta.NewResolver(nil)
	if err != nil {
		log.Fatalf("request metadata: %v", err)
	}
	fallback := ratelimit.NewFallbackLimiter(10000)
	return app.Dependencies{
		Store: store, Valkey: vk, PasswordHasher: hasher, PasswordPolicy: password.DefaultPolicy(), PasswordBlocklist: password.NewBuiltinBlocklist(),
		CookieManager: cookies, CSRFManager: csrfManager, JWTKeyring: keyring, JWTIssuer: "pdv-auth", JWTAudience: "pdv-api", AccessTokenTTL: 5 * time.Minute, JWTClockSkew: 30 * time.Second,
		RequestMeta: meta, RateLimiter: ratelimit.NewValkeyLimiter(vk, fallback), RefreshIdleTTL: 30 * 24 * time.Hour, SessionAbsoluteTTL: 90 * 24 * time.Hour,
		AuthTokenHashKey: []byte("abcdef0123456789abcdef0123456789"), RateLimitKeySecret: []byte("fedcba9876543210fedcba9876543210"),
		RegistrationEnabled: registrationEnabled, RequireVerifiedEmail: requireVerified, AppPublicURL: publicURL, AuthSessionCacheTTL: time.Minute, AuthSessionTouchInterval: 30 * time.Second, Mailer: authMailer,
	}
}

func waitForReady(ctx context.Context) {
	for range 30 {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		select {
		case <-ctx.Done():
			log.Fatal("server not ready before context deadline")
		case <-time.After(500 * time.Millisecond):
		}
	}
	log.Fatal("server not ready after 15s")
}

func getPaymentMethodID(ctx context.Context, pool *pgxpool.Pool, code string) string {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return ""
	}
	defer conn.Release()

	var id string
	if err := conn.QueryRow(ctx, `SELECT id::text FROM payment_methods WHERE code = $1`, code).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return ""
		}
		return ""
	}
	return id
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
