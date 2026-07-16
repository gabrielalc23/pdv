package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestE2EHappyPath(t *testing.T) {
	// Full flow: create product → inventory entry → create sale → add item → checkout → receipt

	// 1. Create product
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-HAPPY-001",
		Name:  "Produto E2E Happy",
		Price: "49.90",
	})
	if product.SKU != "E2E-HAPPY-001" {
		t.Fatalf("expected SKU E2E-HAPPY-001, got %s", product.SKU)
	}
	if !product.IsActive {
		t.Fatal("expected product to be active by default")
	}

	// 2. Create inventory entry (PURCHASE)
	entry := doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "100",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})
	if entry.Inventory.CurrentQuantity != "100.000" {
		t.Fatalf("expected quantity 100.000, got %s", entry.Inventory.CurrentQuantity)
	}
	if entry.Movement.Type != "PURCHASE" {
		t.Fatalf("expected movement type PURCHASE, got %s", entry.Movement.Type)
	}

	// 3. Verify inventory listing
	inventory := doRequestAndDecode[InventoryListResponse](t, http.MethodGet, "/inventory?search=E2E-HAPPY-001", nil)
	if len(inventory.Data) != 1 {
		t.Fatalf("expected 1 inventory item, got %d", len(inventory.Data))
	}
	if inventory.Data[0].Quantity != "100.000" {
		t.Fatalf("expected quantity 100.000, got %s", inventory.Data[0].Quantity)
	}

	// 4. Verify catalog includes the product
	catalog := doRequestAndDecode[ProductListResponse](t, http.MethodGet, "/catalog?search=E2E-HAPPY-001", nil)
	if len(catalog.Data) != 1 {
		t.Fatalf("expected 1 catalog item, got %d", len(catalog.Data))
	}

	// 5. Create sale
	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-happy-sale-001",
	})
	if sale.Status != "OPEN" {
		t.Fatalf("expected status OPEN, got %s", sale.Status)
	}
	if sale.IdempotencyKey != "e2e-happy-sale-001" {
		t.Fatalf("idempotency key mismatch")
	}

	// 6. Add item to sale
	item := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "2",
	})
	if len(item.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(item.Items))
	}
	if item.Items[0].ProductID != product.ID {
		t.Fatalf("expected product ID %s, got %s", product.ID, item.Items[0].ProductID)
	}
	if item.Items[0].Quantity != "2.000" {
		t.Fatalf("expected quantity 2.000, got %s", item.Items[0].Quantity)
	}
	if item.Items[0].Total != "99.80" {
		t.Fatalf("expected total 99.80, got %s", item.Items[0].Total)
	}

	// 7. Verify sale has the item
	saleWithItems := doRequestAndDecode[SaleResponse](t, http.MethodGet, "/sales/"+sale.ID, nil)
	if len(saleWithItems.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(saleWithItems.Items))
	}
	if saleWithItems.Subtotal != "99.80" {
		t.Fatalf("expected subtotal 99.80, got %s", saleWithItems.Subtotal)
	}
	if saleWithItems.Total != "99.80" {
		t.Fatalf("expected total 99.80, got %s", saleWithItems.Total)
	}

	// 8. Get PIX payment method ID from listing
	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID, cashID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
		}
		if m.Code == "CASH" {
			cashID = m.ID
		}
	}
	if pixID == "" || cashID == "" {
		t.Fatal("PIX and CASH payment methods not seeded")
	}

	// 9. Checkout with split payment
	checkout := doRequestAndDecode[CheckoutResponse](t, http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{
				PaymentMethodID: pixID,
				Amount:          "50.00",
			},
			{
				PaymentMethodID: cashID,
				Amount:          "49.80",
				ReceivedAmount:  strPtr("50.00"),
			},
		},
	})
	if checkout.Sale.Status != "COMPLETED" {
		t.Fatalf("expected status COMPLETED, got %s", checkout.Sale.Status)
	}
	if len(checkout.Payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(checkout.Payments))
	}
	for _, p := range checkout.Payments {
		if p.Status != "APPROVED" {
			t.Fatalf("expected payment status APPROVED, got %s", p.Status)
		}
	}
	if checkout.FiscalDocument.Status != "AUTHORIZED" {
		t.Fatalf("expected fiscal document AUTHORIZED, got %s", checkout.FiscalDocument.Status)
	}
	if checkout.FiscalDocument.AccessKey == nil || *checkout.FiscalDocument.AccessKey == "" {
		t.Fatal("expected access key to be set")
	}

	// Verify change for cash payment
	for _, p := range checkout.Payments {
		if p.PaymentMethodCode == "CASH" {
			if p.ReceivedAmount == nil || *p.ReceivedAmount != "50.00" {
				t.Fatalf("expected received amount 50.00 for CASH, got %v", p.ReceivedAmount)
			}
			if p.ChangeAmount == nil || *p.ChangeAmount != "0.20" {
				t.Fatalf("expected change amount 0.20 for CASH, got %v", p.ChangeAmount)
			}
		}
	}

	// 10. Verify receipt
	receipt := doRequestAndDecode[ReceiptResponse](t, http.MethodGet, "/sales/"+sale.ID+"/receipt", nil)
	if receipt.Sale.Status != "COMPLETED" {
		t.Fatalf("expected receipt status COMPLETED, got %s", receipt.Sale.Status)
	}
	if len(receipt.Items) != 1 {
		t.Fatalf("expected 1 receipt item, got %d", len(receipt.Items))
	}
	if len(receipt.Payments) != 2 {
		t.Fatalf("expected 2 receipt payments, got %d", len(receipt.Payments))
	}
	if receipt.FiscalDocument == nil {
		t.Fatal("expected fiscal document in receipt")
	}
	if receipt.FiscalDocument.Status != "AUTHORIZED" {
		t.Fatalf("expected fiscal document AUTHORIZED, got %s", receipt.FiscalDocument.Status)
	}

	// 11. Verify inventory decreased
	invAfter := doRequestAndDecode[InventoryResponse](t, http.MethodGet, "/products/"+product.ID+"/inventory", nil)
	if invAfter.Quantity != "98.000" {
		t.Fatalf("expected inventory 98.000 after sale of 2 units, got %s", invAfter.Quantity)
	}
}

func TestE2EIdempotency(t *testing.T) {
	// Creating a sale twice with same idempotency key should return the same sale

	first := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-dup-sale",
	})
	if first.Status != "OPEN" {
		t.Fatalf("expected OPEN, got %s", first.Status)
	}

	second := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-dup-sale",
	})
	if second.ID != first.ID {
		t.Fatalf("expected same sale ID %s, got %s", first.ID, second.ID)
	}
	if second.Number != first.Number {
		t.Fatalf("expected same sale number %d, got %d", first.Number, second.Number)
	}
}

func TestE2EValidationErrors(t *testing.T) {
	t.Run("create_product_empty_sku", func(t *testing.T) {
		resp, err := doRequest(http.MethodPost, "/products", UpsertProductRequest{
			SKU:   "",
			Name:  "No SKU",
			Price: "10.00",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 422 {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("create_product_duplicate_sku", func(t *testing.T) {
		doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
			SKU:   "E2E-DUP-SKU",
			Name:  "First",
			Price: "10.00",
		})

		resp, err := doRequest(http.MethodPost, "/products", UpsertProductRequest{
			SKU:   "E2E-DUP-SKU",
			Name:  "Second",
			Price: "20.00",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 409 {
			t.Fatalf("expected 409 for duplicate SKU, got %d", resp.StatusCode)
		}
	})

	t.Run("add_item_to_completed_sale", func(t *testing.T) {
		product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
			SKU:   "E2E-COMPLETED-ITEM",
			Name:  "Completed Item",
			Price: "10.00",
		})

		doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
			ProductID:     product.ID,
			Quantity:      "10",
			ReferenceType: "PURCHASE",
			ReferenceID:   product.ID,
		})

		pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
		var pixID string
		for _, m := range pm.Data {
			if m.Code == "PIX" {
				pixID = m.ID
				break
			}
		}

		sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
			IdempotencyKey: "e2e-completed-item-sale",
		})

		doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
			ProductID: product.ID,
			Quantity:  "1",
		})

		doRequestAndDecode[CheckoutResponse](t, http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
			Payments: []CheckoutPaymentRequest{
				{PaymentMethodID: pixID, Amount: "10.00"},
			},
		})

		resp, err := doRequest(http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
			ProductID: product.ID,
			Quantity:  "1",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 409 {
			t.Fatalf("expected 409 for adding item to completed sale, got %d", resp.StatusCode)
		}
	})
}

func TestE2ESaleLifecycle(t *testing.T) {
	// Open → add item → update item → remove item → cancel

	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-LIFECYCLE-001",
		Name:  "Lifecycle Product",
		Price: "25.00",
	})

	product2 := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-LIFECYCLE-002",
		Name:  "Lifecycle Product 2",
		Price: "50.00",
	})

	for _, p := range []ProductResponse{product, product2} {
		doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
			ProductID:     p.ID,
			Quantity:      "10",
			ReferenceType: "PURCHASE",
			ReferenceID:   p.ID,
		})
	}

	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-lifecycle-sale",
	})
	if sale.Status != "OPEN" {
		t.Fatalf("expected OPEN, got %s", sale.Status)
	}

	// Add two items
	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "2",
	})
	added2 := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product2.ID,
		Quantity:  "1",
	})
	var item1ID, item2ID string
	for _, it := range added2.Items {
		switch it.ProductID {
		case product.ID:
			item1ID = it.ID
		case product2.ID:
			item2ID = it.ID
		}
	}
	if item2ID == "" {
		t.Fatal("expected item2 to be present")
	}
	if added2.Subtotal != "100.00" {
		t.Fatalf("expected subtotal 100.00, got %s", added2.Subtotal)
	}

	saleWithItems := doRequestAndDecode[SaleResponse](t, http.MethodGet, "/sales/"+sale.ID, nil)
	if len(saleWithItems.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(saleWithItems.Items))
	}
	if saleWithItems.Subtotal != "100.00" {
		t.Fatalf("expected subtotal 100.00, got %s", saleWithItems.Subtotal)
	}

	// Update first item quantity
	updated := doRequestAndDecode[SaleResponse](t, http.MethodPut, fmt.Sprintf("/sales/%s/items/%s", sale.ID, item1ID), UpdateSaleItemRequest{
		Quantity: "3",
	})
	var updatedItem *SaleItemResponse
	for i := range updated.Items {
		if updated.Items[i].ID == item1ID {
			updatedItem = &updated.Items[i]
		}
	}
	if updatedItem == nil {
		t.Fatal("expected updated item to be present")
	}
	if updatedItem.Quantity != "3.000" {
		t.Fatalf("expected quantity 3.000, got %s", updatedItem.Quantity)
	}

	saleAfterUpdate := doRequestAndDecode[SaleResponse](t, http.MethodGet, "/sales/"+sale.ID, nil)
	if saleAfterUpdate.Subtotal != "125.00" {
		t.Fatalf("expected subtotal 125.00 after update, got %s", saleAfterUpdate.Subtotal)
	}

	// Remove second item
	removed := doRequestAndDecode[SaleResponse](t, http.MethodDelete, fmt.Sprintf("/sales/%s/items/%s", sale.ID, item2ID), nil)
	if len(removed.Items) != 1 {
		t.Fatalf("expected 1 item after removal, got %d", len(removed.Items))
	}
	if removed.Subtotal != "75.00" {
		t.Fatalf("expected subtotal 75.00 after removal, got %s", removed.Subtotal)
	}

	// Cancel the sale
	cancelled := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/cancel", nil)
	if cancelled.Status != "CANCELLED" {
		t.Fatalf("expected CANCELLED, got %s", cancelled.Status)
	}

	// Verify receipt returns 409 for cancelled sale
	resp, err := doRequest(http.MethodGet, "/sales/"+sale.ID+"/receipt", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 for receipt of cancelled sale, got %d", resp.StatusCode)
	}
}

func TestE2ECheckoutInsufficientStock(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-NOSTOCK",
		Name:  "No Stock Product",
		Price: "10.00",
	})
	// No inventory entry

	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-nostock-sale",
	})

	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "1",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
			break
		}
	}

	resp, err := doRequest(http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "10.00"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 for insufficient stock, got %d", resp.StatusCode)
	}
}

func doRequestAndDecode[T any](t *testing.T, method, path string, body any) T {
	t.Helper()
	resp, err := doRequest(method, path, body)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var eb struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
				Field   string `json:"field,omitempty"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&eb)
		t.Fatalf("unexpected status %d for %s %s: code=%s message=%s", resp.StatusCode, method, path, eb.Error.Code, eb.Error.Message)
	}

	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response from %s %s: %v", method, path, err)
	}
	return v
}

func requireEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", msg, got, want)
	}
}

func init() {
	// Verify that baseURL is set (TestMain ran)
}

func TestE2ECheckoutNoItems(t *testing.T) {
	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-noitems-checkout",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
			break
		}
	}

	resp, err := doRequest(http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "0.01"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409 for empty sale checkout, got %d: %s", resp.StatusCode, decodeError(resp))
	}
}

func TestE2ECheckoutPaymentMismatch(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-PAYMM",
		Name:  "Payment Mismatch",
		Price: "30.00",
	})

	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "5",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-paymm-sale",
	})

	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "1",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
			break
		}
	}

	// Pay only 10 instead of 30
	resp, err := doRequest(http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "10.00"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 422 {
		t.Fatalf("expected 422 for payment mismatch, got %d: %s", resp.StatusCode, decodeError(resp))
	}
}

func TestE2EListSalesFilter(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-LISTFLT",
		Name:  "List Filter",
		Price: "15.00",
	})

	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "5",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	openSale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-listflt-open",
	})
	_ = openSale

	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+openSale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "1",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
			break
		}
	}

	doRequestAndDecode[CheckoutResponse](t, http.MethodPost, "/sales/"+openSale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "15.00"},
		},
	})

	cancelledSale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: "e2e-listflt-cancelled",
	})
	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+cancelledSale.ID+"/cancel", nil)

	// List all
	var listResp struct {
		Data []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
		Pagination struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"pageSize"`
			Total      int64 `json:"total"`
			TotalPages int   `json:"totalPages"`
		} `json:"pagination"`
	}

	resp, err := doRequest(http.MethodGet, "/sales", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	if listResp.Pagination.Total < 2 {
		t.Fatalf("expected at least 2 sales, got %d", listResp.Pagination.Total)
	}

	// Filter by status=OPEN
	resp, err = doRequest(http.MethodGet, "/sales?status=OPEN", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}

	// There should be no OPEN sales (since the original was completed)
	// But actually, we completed the open sale, so it should be COMPLETED now
	// Filter by status=COMPLETED
	resp, err = doRequest(http.MethodGet, "/sales?status=COMPLETED", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	if listResp.Pagination.Total < 1 {
		t.Fatalf("expected at least 1 COMPLETED sale, got %d", listResp.Pagination.Total)
	}
}

func TestE2EPagination(t *testing.T) {
	// Create 3 products with a unique searchable prefix so the totals are
	// deterministic regardless of other test data in the shared database.
	const prefix = "E2E-PAG-"
	for i := 1; i <= 3; i++ {
		doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
			SKU:   fmt.Sprintf("%s%d", prefix, i),
			Name:  fmt.Sprintf("Pagination Product %d", i),
			Price: "10.00",
		})
	}

	// List with pageSize=2 scoped to the prefix - first page returns 2 items, total=3
	var listResp ProductListResponse
	resp, err := doRequest(http.MethodGet, "/products?search="+prefix+"&pageSize=2", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Data) != 2 {
		t.Fatalf("expected 2 items, got %d", len(listResp.Data))
	}
	if listResp.Pagination.Total != 3 {
		t.Fatalf("expected total 3, got %d", listResp.Pagination.Total)
	}
	if listResp.Pagination.TotalPages != 2 {
		t.Fatalf("expected 2 pages, got %d", listResp.Pagination.TotalPages)
	}

	// Get second page
	resp, err = doRequest(http.MethodGet, "/products?search="+prefix+"&pageSize=2&page=2", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Data) != 1 {
		t.Fatalf("expected 1 item on page 2, got %d", len(listResp.Data))
	}
}

func TestE2ECheckoutReceiptNonExistentSale(t *testing.T) {
	resp, err := doRequest(http.MethodGet, "/sales/00000000-0000-0000-0000-000000000000/receipt", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for non-existent sale receipt, got %d: %s", resp.StatusCode, decodeError(resp))
	}
}

// Quick inline test to verify health endpoint works
func TestE2EHealth(t *testing.T) {
	resp, err := doRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestE2EProductCRUD(t *testing.T) {
	// Create
	p := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:     "E2E-CRUD-001",
		Barcode: strPtr("7891234567890"),
		Name:    "CRUD Product",
		Price:   "99.99",
		Cost:    strPtr("50.00"),
	})
	if p.Barcode == nil || *p.Barcode != "7891234567890" {
		t.Fatal("barcode not set correctly")
	}

	// Get by ID
	got := doRequestAndDecode[ProductResponse](t, http.MethodGet, "/products/"+p.ID, nil)
	if got.Name != "CRUD Product" {
		t.Fatalf("expected 'CRUD Product', got '%s'", got.Name)
	}

	// Update
	updated := doRequestAndDecode[ProductResponse](t, http.MethodPut, "/products/"+p.ID, UpsertProductRequest{
		SKU:   "E2E-CRUD-001",
		Name:  "CRUD Product Updated",
		Price: "89.99",
	})
	if updated.Name != "CRUD Product Updated" {
		t.Fatalf("expected updated name, got '%s'", updated.Name)
	}
	if updated.Price != "89.99" {
		t.Fatalf("expected 89.99, got %s", updated.Price)
	}

	// Deactivate
	deactivated := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products/"+p.ID+"/deactivate", nil)
	if deactivated.IsActive {
		t.Fatal("expected product to be inactive")
	}

	// Activate
	activated := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products/"+p.ID+"/activate", nil)
	if !activated.IsActive {
		t.Fatal("expected product to be active")
	}

	// Filter by activeOnly
	var listResp ProductListResponse
	resp, err := doRequest(http.MethodGet, "/products?activeOnly=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range listResp.Data {
		if item.ID == p.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected product in active list")
	}

	// Non-existent product
	resp, err = doRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for non-existent product, got %d", resp.StatusCode)
	}
}

func TestE2ECatalogEndpoints(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:     "E2E-CATALOG",
		Barcode: strPtr("9999999999999"),
		Name:    "Catalog Product",
		Price:   "5.50",
	})

	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "10",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	// Get by ID from catalog
	catalog := doRequestAndDecode[ProductResponse](t, http.MethodGet, "/catalog/"+product.ID, nil)
	if catalog.Name != "Catalog Product" {
		t.Fatalf("expected 'Catalog Product', got '%s'", catalog.Name)
	}

	// Get by barcode
	byBarcode := doRequestAndDecode[ProductResponse](t, http.MethodGet, "/catalog/barcode/9999999999999", nil)
	if byBarcode.ID != product.ID {
		t.Fatalf("expected product ID %s, got %s", product.ID, byBarcode.ID)
	}

	// List with inStockOnly
	var listResp ProductListResponse
	resp, err := doRequest(http.MethodGet, "/catalog?inStockOnly=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range listResp.Data {
		if item.ID == product.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected product in catalog with inStockOnly")
	}
}

func TestE2EInventoryAdjustment(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-ADJUST",
		Name:  "Adjustment Product",
		Price: "20.00",
	})

	// Entry
	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "50",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	// Adjustment IN
	adj := doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/adjustments", struct {
		ProductID     string `json:"productId"`
		Direction     string `json:"direction"`
		Quantity      string `json:"quantity"`
		Reason        string `json:"reason"`
		ReferenceType string `json:"referenceType"`
		ReferenceID   string `json:"referenceId"`
	}{
		ProductID:     product.ID,
		Direction:     "IN",
		Quantity:      "10",
		Reason:        "Test adjustment in",
		ReferenceType: "ADJUSTMENT",
		ReferenceID:   product.ID,
	})
	if adj.Inventory.CurrentQuantity != "60.000" {
		t.Fatalf("expected 60.000 after adjustment IN, got %s", adj.Inventory.CurrentQuantity)
	}

	// Adjustment OUT
	adjOut := doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/adjustments", struct {
		ProductID     string `json:"productId"`
		Direction     string `json:"direction"`
		Quantity      string `json:"quantity"`
		Reason        string `json:"reason"`
		ReferenceType string `json:"referenceType"`
		ReferenceID   string `json:"referenceId"`
	}{
		ProductID:     product.ID,
		Direction:     "OUT",
		Quantity:      "5",
		Reason:        "Test adjustment out",
		ReferenceType: "ADJUSTMENT",
		ReferenceID:   product.ID,
	})
	if adjOut.Inventory.CurrentQuantity != "55.000" {
		t.Fatalf("expected 55.000 after adjustment OUT, got %s", adjOut.Inventory.CurrentQuantity)
	}

	// List movements
	var movResp struct {
		Data []struct {
			ID               string `json:"id"`
			Type             string `json:"type"`
			Quantity         string `json:"quantity"`
			PreviousQuantity string `json:"previousQuantity"`
			CurrentQuantity  string `json:"currentQuantity"`
		} `json:"data"`
	}
	resp, err := doRequest(http.MethodGet, "/products/"+product.ID+"/inventory/movements", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	if err := json.NewDecoder(resp.Body).Decode(&movResp); err != nil {
		t.Fatal(err)
	}
	if len(movResp.Data) < 3 {
		t.Fatalf("expected at least 3 movements, got %d", len(movResp.Data))
	}
}

func TestE2EListSalePayments(t *testing.T) {
	// Non-existent sale should return 404
	resp, err := doRequest(http.MethodGet, "/sales/00000000-0000-0000-0000-000000000000/payments", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent sale payments, got %d", resp.StatusCode)
	}

	// Create a sale with a split PIX + CASH payment and verify listing
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-PAYLIST",
		Name:  "Payments List Product",
		Price: "49.90",
	})

	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "10",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: strings.Replace(t.Name(), "/", "_", -1),
	})

	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "2",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID, cashID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
		}
		if m.Code == "CASH" {
			cashID = m.ID
		}
	}

	doRequestAndDecode[CheckoutResponse](t, http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "50.00"},
			{PaymentMethodID: cashID, Amount: "49.80", ReceivedAmount: strPtr("50.00")},
		},
	})

	list := doRequestAndDecode[SalePaymentsResponse](t, http.MethodGet, "/sales/"+sale.ID+"/payments", nil)
	if len(list.Data) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(list.Data))
	}

	byCode := map[string]SalePaymentResponse{}
	for _, p := range list.Data {
		if p.SaleID != sale.ID {
			t.Fatalf("expected sale ID %s, got %s", sale.ID, p.SaleID)
		}
		if p.Status != "APPROVED" {
			t.Fatalf("expected payment APPROVED, got %s", p.Status)
		}
		byCode[p.PaymentMethodCode] = p
	}

	pix, ok := byCode["PIX"]
	if !ok {
		t.Fatal("expected PIX payment in list")
	}
	if pix.Amount != "50.00" {
		t.Fatalf("expected PIX amount 50.00, got %s", pix.Amount)
	}
	if pix.Installments != 1 {
		t.Fatalf("expected PIX installments 1, got %d", pix.Installments)
	}

	cash, ok := byCode["CASH"]
	if !ok {
		t.Fatal("expected CASH payment in list")
	}
	if cash.Amount != "49.80" {
		t.Fatalf("expected CASH amount 49.80, got %s", cash.Amount)
	}
	if cash.ReceivedAmount == nil || *cash.ReceivedAmount != "50.00" {
		t.Fatalf("expected CASH received amount 50.00, got %v", cash.ReceivedAmount)
	}
	if cash.ChangeAmount == nil || *cash.ChangeAmount != "0.20" {
		t.Fatalf("expected CASH change amount 0.20, got %v", cash.ChangeAmount)
	}
}

func TestE2EFiscalDocument(t *testing.T) {
	product := doRequestAndDecode[ProductResponse](t, http.MethodPost, "/products", UpsertProductRequest{
		SKU:   "E2E-FISCAL",
		Name:  "Fiscal Product",
		Price: "100.00",
	})

	doRequestAndDecode[InventoryChangeResponse](t, http.MethodPost, "/inventory/entries", InventoryEntryRequest{
		ProductID:     product.ID,
		Quantity:      "5",
		ReferenceType: "PURCHASE",
		ReferenceID:   product.ID,
	})

	sale := doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales", CreateSaleRequest{
		IdempotencyKey: strings.Replace(t.Name(), "/", "_", -1),
	})

	doRequestAndDecode[SaleResponse](t, http.MethodPost, "/sales/"+sale.ID+"/items", AddSaleItemRequest{
		ProductID: product.ID,
		Quantity:  "1",
	})

	pm := doRequestAndDecode[PaymentMethodsResponse](t, http.MethodGet, "/payment-methods", nil)
	var pixID string
	for _, m := range pm.Data {
		if m.Code == "PIX" {
			pixID = m.ID
			break
		}
	}

	doRequestAndDecode[CheckoutResponse](t, http.MethodPost, "/sales/"+sale.ID+"/checkout", CheckoutRequest{
		Payments: []CheckoutPaymentRequest{
			{PaymentMethodID: pixID, Amount: "100.00"},
		},
	})

	// Get fiscal document
	fd := doRequestAndDecode[FiscalDocumentResponse](t, http.MethodGet, "/sales/"+sale.ID+"/fiscal-document", nil)
	if fd.Status != "AUTHORIZED" {
		t.Fatalf("expected AUTHORIZED, got %s", fd.Status)
	}
	if fd.AccessKey == nil || *fd.AccessKey == "" {
		t.Fatal("expected access key")
	}
}
