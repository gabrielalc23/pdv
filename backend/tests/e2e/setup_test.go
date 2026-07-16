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

	"github.com/gabrielalc23/pdv/internal/catalog"
	"github.com/gabrielalc23/pdv/internal/checkout"
	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/inventory"
	"github.com/gabrielalc23/pdv/internal/payments"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/internal/sales"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var baseURL string

func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://pdv:pdv@localhost:5432/pdv?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if !tryConnect(ctx, dsn) {
		startDeps(ctx, dsn)
	}

	pool := mustConnect(ctx, dsn)
	defer pool.Close()

	// create separate test database and reconnect to it
	testDBName := "pdv_e2e_test"
	createTestDB(ctx, pool, testDBName)
	pool.Close()

	testDSN := replaceDBName(dsn, testDBName)
	testPool := mustConnect(ctx, testDSN)
	defer testPool.Close()

	runMigrations(ctx, testPool)
	seedPaymentMethods(ctx, testPool)

	store := database.NewStore(testPool)
	lis := mustListen()
	baseURL = "http://" + lis.Addr().String()
	server := startServer(store, lis.Addr().String())

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
		"-f", "../docker-compose.test.yml",
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
		time.Sleep(time.Second)
	}
	log.Fatal("postgres not ready after 30s")
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

func seedPaymentMethods(ctx context.Context, pool *pgxpool.Pool) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM payment_methods`).Scan(&count); err != nil {
		log.Fatalf("count payment_methods: %v", err)
	}
	if count > 0 {
		return
	}

	methods := []struct {
		code, name, kind string
		allowsChange     bool
		allowsInstall    bool
		maxInstall       int16
		sortOrder        int
	}{
		{"CASH", "Dinheiro", "CASH", true, false, 1, 1},
		{"PIX", "PIX", "PIX", false, false, 1, 2},
		{"DEBIT", "Cartão de Débito", "DEBIT_CARD", false, false, 1, 3},
		{"CREDIT", "Cartão de Crédito", "CREDIT_CARD", false, true, 12, 4},
		{"VOUCHER", "Vale", "VOUCHER", false, false, 1, 5},
	}

	for _, m := range methods {
		_, err := conn.Exec(ctx, `
			INSERT INTO payment_methods (code, name, kind, allows_change, allows_installments, max_installments, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (code) DO NOTHING
		`, m.code, m.name, m.kind, m.allowsChange, m.allowsInstall, m.maxInstall, m.sortOrder)
		if err != nil {
			log.Fatalf("seed payment method %s: %v", m.code, err)
		}
	}
}

func mustListen() net.Listener {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	return lis
}

func startServer(store *database.Store, addr string) *http.Server {
	router := apphttp.NewRouter(apphttp.Dependencies{
		HealthHandler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		},
	})

	productService := products.NewService(store.Queries)
	productHandler := products.NewHandler(productService)
	products.RegisterRoutes(router, productHandler)

	inventoryService := inventory.NewService(store.Queries, inventory.NewTxManager(store))
	inventoryHandler := inventory.NewHandler(inventoryService)
	inventory.RegisterRoutes(router, inventoryHandler)

	catalogService := catalog.NewService(store.Queries)
	catalogHandler := catalog.NewHandler(catalogService)
	catalog.RegisterRoutes(router, catalogHandler)

	salesService := sales.NewService(store.Queries, sales.NewTxManager(store))
	salesHandler := sales.NewHandler(salesService)
	sales.RegisterRoutes(router, salesHandler)

	fiscalProvider := &fiscal.MockProvider{}

	checkoutService := checkout.NewService(checkout.NewTxManager(store), fiscalProvider, store)
	checkoutHandler := checkout.NewHandler(checkoutService)
	checkout.RegisterRoutes(router, checkoutHandler)

	paymentsService := payments.NewService(store.Queries)
	paymentsHandler := payments.NewHandler(paymentsService)
	payments.RegisterRoutes(router, paymentsHandler)

	fiscalService := fiscal.NewService(store.Queries, fiscalProvider)
	fiscalHandler := fiscal.NewHandler(fiscalService)
	fiscal.RegisterRoutes(router, fiscalHandler)

	receiptService := receipt.NewService(store.Queries)
	receiptHandler := receipt.NewHandler(receiptService)
	receipt.RegisterRoutes(router, receiptHandler)

	return &http.Server{
		Addr:    addr,
		Handler: router,
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
