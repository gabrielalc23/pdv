import { useEffect, useRef, useState } from "react";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Input } from "@pdv/ui-kit/components/input";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { Link } from "@tanstack/react-router";
import { PageHeader } from "../components/page-header";
import {
  useAddSaleItemMutation,
  useCreateSaleMutation,
  useRemoveSaleItemMutation,
  useUpdateSaleItemMutation,
} from "../mutations/sale.mutation";
import { useCheckoutSaleMutation } from "../mutations/checkout.mutation";
import { useListCatalogQuery } from "../queries/catalog.query";
import { useListPaymentMethodsQuery } from "../queries/payment.query";
import { useGetSaleQuery, useListSalesQuery } from "../queries/sale.query";
import type { CatalogProductResponse } from "../interfaces/catalog.interface";
import type { SaleResponse } from "../interfaces/sale.interface";
import { formatCurrency } from "../utils/format-currency.util";
import { Plus, Receipt, Search, ShoppingBag, Sliders } from "lucide-react";
import { ProductTile } from "../components/product-title";
import { formatDate } from "../utils/format-date.util";
import { useSheet } from "../hooks/use-sheet.hook";

const tones = ["coral", "mint", "blue", "sand"] as const;

export function PosPage() {
  const [search, setSearch] = useState("");
  const [sale, setSale] = useState<SaleResponse | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [selectedOpenSaleId, setSelectedOpenSaleId] = useState<string | null>(
    null,
  );
  const catalogSearchRef = useRef<HTMLInputElement>(null);
  const cartSheet = useSheet("carrinho");
  const shortcutsSheet = useSheet("atalhos");
  const openOrdersSheet = useSheet("pedidos-abertos");
  const catalogQuery = useListCatalogQuery({
    activeOnly: true,
    inStockOnly: true,
    page: 1,
    pageSize: 100,
  });
  const paymentMethodsQuery = useListPaymentMethodsQuery();
  const openSalesQuery = useListSalesQuery({
    status: "OPEN",
    page: 1,
    pageSize: 50,
  });
  const selectedOpenSaleQuery = useGetSaleQuery(selectedOpenSaleId ?? "");
  const createSale = useCreateSaleMutation();
  const addSaleItem = useAddSaleItemMutation();
  const removeSaleItem = useRemoveSaleItemMutation();
  const updateSaleItem = useUpdateSaleItemMutation();
  const checkoutSale = useCheckoutSaleMutation();
  const products = (catalogQuery.data?.data ?? []).filter((product) => {
    const normalizedSearch = search.toLowerCase();
    return [product.name, product.sku, product.barcode ?? ""].some((value) =>
      value.toLowerCase().includes(normalizedSearch),
    );
  });
  const isBusy: boolean =
    createSale.isPending ||
    addSaleItem.isPending ||
    removeSaleItem.isPending ||
    updateSaleItem.isPending ||
    checkoutSale.isPending;

  useEffect(() => {
    if (!selectedOpenSaleQuery.data) return;

    setSale(selectedOpenSaleQuery.data);
    setSelectedOpenSaleId(null);
    openOrdersSheet.setOpen(false);
    setFeedback(null);
  }, [selectedOpenSaleQuery.data]);

  useEffect(() => {
    function handleShortcut(event: KeyboardEvent): void {
      if (event.key === "F1") {
        event.preventDefault();
        shortcutsSheet.setOpen(true);
      }

      if (event.key === "F2") {
        event.preventDefault();
        catalogSearchRef.current?.focus();
      }

      if (event.key === "F4") {
        event.preventDefault();
        cartSheet.setOpen(true);
      }
    }

    document.addEventListener("keydown", handleShortcut);
    return () => document.removeEventListener("keydown", handleShortcut);
  }, []);

  async function addProduct(product: CatalogProductResponse) {
    setFeedback(null);
    try {
      const currentSale =
        sale ??
        (await createSale.mutateAsync({ idempotencyKey: crypto.randomUUID() }));
      if (!sale) setSale(currentSale);
      const updatedSale = await addSaleItem.mutateAsync({
        saleId: currentSale.id,
        data: { productId: product.id, quantity: "1", discount: null },
      });
      setSale(updatedSale);
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Não foi possível adicionar o produto.",
      );
    }
  }

  async function increaseProduct(product: CatalogProductResponse) {
    const item = sale?.items.find(
      (saleItem) => saleItem.productId === product.id,
    );
    if (!item) {
      await addProduct(product);
      return;
    }

    setFeedback(null);
    try {
      setSale(
        await updateSaleItem.mutateAsync({
          saleId: item.saleId,
          itemId: item.id,
          data: {
            quantity: String(Number(item.quantity) + 1),
            discount: item.discount,
          },
        }),
      );
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Não foi possível adicionar o produto.",
      );
    }
  }

  async function decreaseProduct(product: CatalogProductResponse) {
    const item = sale?.items.find(
      (saleItem) => saleItem.productId === product.id,
    );
    if (!item || !sale) return;

    const quantity = Number(item.quantity);
    if (quantity <= 1) {
      await removeProduct(item.id);
      return;
    }

    setFeedback(null);
    try {
      setSale(
        await updateSaleItem.mutateAsync({
          saleId: sale.id,
          itemId: item.id,
          data: { quantity: String(quantity - 1), discount: item.discount },
        }),
      );
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Não foi possível remover o produto.",
      );
    }
  }

  function productQuantity(productId: string): number {
    return Number(
      sale?.items.find((item) => item.productId === productId)?.quantity ?? 0,
    );
  }

  async function removeProduct(itemId: string) {
    if (!sale) return;
    setFeedback(null);
    try {
      setSale(await removeSaleItem.mutateAsync({ saleId: sale.id, itemId }));
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Não foi possível remover o produto.",
      );
    }
  }

  async function handleCheckout() {
    if (!sale) return;
    const paymentMethod = paymentMethodsQuery.data?.data.find(
      (method) => method.isActive,
    );
    if (!paymentMethod) {
      setFeedback("Nenhum meio de pagamento ativo foi encontrado.");
      return;
    }
    setFeedback(null);
    try {
      await checkoutSale.mutateAsync({
        saleId: sale.id,
        data: {
          payments: [
            {
              paymentMethodId: paymentMethod.id,
              amount: sale.total,
              installments: 1,
            },
          ],
        },
      });
      setSale(null);
      setFeedback("Venda finalizada com sucesso.");
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Não foi possível finalizar a venda.",
      );
    }
  }

  return (
    <Sheet open={cartSheet.isOpen} onOpenChange={cartSheet.setOpen}>
      <div>
        <PageHeader
          breadcrumbs={[
            { label: "Visão geral", to: "/" },
            { label: "Novo atendimento" },
          ]}
          title="Novo atendimento"
          description="Encontre produtos no catálogo e monte o pedido do cliente."
          actionPlacement="below"
          action={
            <div className="flex w-full flex-col gap-2 sm:flex-row sm:items-center">
              <div className="flex h-10 min-w-0 flex-1 items-center gap-2 rounded-md border border-(--line) bg-(--contrast-light)/70 px-3">
                <Search className="size-4 text-(--ink-soft)" />
                <Input
                  ref={catalogSearchRef}
                  className="border-0 bg-(--transparent) px-0 shadow-none focus-visible:ring-0"
                  placeholder="Buscar produto..."
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                />
              </div>
              <div className="flex shrink-0 flex-wrap gap-2">
                <Sheet.Trigger
                  render={
                    <Button className="rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)" />
                  }
                >
                  <ShoppingBag className="size-4" />
                  Carrinho
                  <span className="grid size-5 place-items-center rounded-full bg-(--contrast-light)/15 text-[11px]">
                    {sale?.items.length ?? 0}
                  </span>
                </Sheet.Trigger>
                <Sheet
                  open={shortcutsSheet.isOpen}
                  onOpenChange={shortcutsSheet.setOpen}
                >
                  <Sheet.Trigger
                    render={
                      <Button
                        variant="outline"
                        className="rounded-md border-(--line)"
                      />
                    }
                  >
                    <Sliders className="size-4" /> Atalhos
                  </Sheet.Trigger>
                  <Sheet.Content
                    side="right"
                    className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
                  >
                    <Sheet.Header className="px-6 py-5">
                      <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                        Operação rápida
                      </p>
                      <Sheet.Title className="font-serif text-2xl text-(--ink)">
                        Atalhos do atendimento
                      </Sheet.Title>
                      <Sheet.Description className="text-(--ink-soft)">
                        Acesse as ações mais usadas sem tirar as mãos do
                        teclado.
                      </Sheet.Description>
                    </Sheet.Header>
                    <div className="flex-1 overflow-y-auto px-6 py-5">
                      <div className="divide-y divide-(--line) rounded-md border border-(--line) bg-(--paper)">
                        <ShortcutRow
                          keys="F1"
                          label="Abrir esta lista de atalhos"
                        />
                        <ShortcutRow
                          keys="F2"
                          label="Focar na busca de produtos"
                        />
                        <ShortcutRow keys="F4" label="Abrir o carrinho atual" />
                        <ShortcutRow
                          keys="⌘ K"
                          label="Abrir a navegação do sistema"
                        />
                        <ShortcutRow keys="Esc" label="Fechar o painel atual" />
                      </div>
                      <div className="mt-6 rounded-md bg-(--coral-wash) p-4">
                        <p className="text-sm font-semibold text-(--coral-dark)">
                          Dica rápida
                        </p>
                        <p className="mt-1 text-xs leading-relaxed text-(--ink-soft)">
                          Use <strong>F2</strong> para localizar um produto
                          rapidamente e<strong> F4</strong> para revisar o
                          pedido antes do pagamento.
                        </p>
                      </div>
                    </div>
                    <Sheet.Footer className="bg-(--surface) px-6 py-5">
                      <p className="w-full text-xs text-(--ink-soft)">
                        Os atalhos funcionam enquanto o atendimento estiver
                        aberto.
                      </p>
                    </Sheet.Footer>
                  </Sheet.Content>
                </Sheet>
                <Sheet
                  open={openOrdersSheet.isOpen}
                  onOpenChange={openOrdersSheet.setOpen}
                >
                  <Sheet.Trigger
                    render={
                      <Button className="rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)" />
                    }
                  >
                    <Receipt className="size-4" /> Pedidos abertos
                  </Sheet.Trigger>
                  <Sheet.Content
                    side="right"
                    className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
                  >
                    <Sheet.Header className=" px-6 py-5">
                      <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                        Atendimento
                      </p>
                      <Sheet.Title className="font-serif text-2xl text-(--ink)">
                        Pedidos abertos
                      </Sheet.Title>
                      <Sheet.Description className="text-(--ink-soft)">
                        Retome um pedido pausado para continuar o atendimento.
                      </Sheet.Description>
                    </Sheet.Header>
                    <div className="flex-1 overflow-y-auto px-6 py-5">
                      {openSalesQuery.error && (
                        <p className="text-sm text-(--coral-dark)">
                          {openSalesQuery.error.message}
                        </p>
                      )}
                      {openSalesQuery.isLoading && (
                        <p className="text-sm text-(--ink-soft)">
                          Carregando pedidos...
                        </p>
                      )}
                      {!openSalesQuery.isLoading &&
                        !openSalesQuery.error &&
                        openSalesQuery.data?.data.length === 0 && (
                          <div className="flex h-full min-h-64 flex-col items-center justify-center text-center">
                            <div className="mb-4 grid size-14 place-items-center rounded-full border border-dashed border-(--line) text-(--ink-soft)/40">
                              <Receipt className="size-5" />
                            </div>
                            <p className="text-sm font-semibold text-(--ink)">
                              Nenhum pedido em aberto
                            </p>
                            <p className="mt-2 max-w-55 text-xs leading-relaxed text-(--ink-soft)">
                              Os pedidos pausados aparecerão aqui para você
                              continuar o atendimento.
                            </p>
                          </div>
                        )}
                      <div className="space-y-3">
                        {openSalesQuery.data?.data.map((openSale) => (
                          <button
                            key={openSale.id}
                            type="button"
                            onClick={() => setSelectedOpenSaleId(openSale.id)}
                            disabled={selectedOpenSaleQuery.isLoading}
                            className="flex w-full items-center gap-3 rounded-md border border-(--line) bg-(--paper) p-4 text-left transition-colors hover:border-(--coral-border) hover:bg-(--coral-wash) disabled:pointer-events-none disabled:opacity-60"
                          >
                            <span className="grid size-10 shrink-0 place-items-center rounded-md bg-(--coral-wash) text-(--coral-dark)">
                              <Receipt className="size-4" />
                            </span>
                            <span className="min-w-0 flex-1">
                              <span className="block text-sm font-semibold">
                                Pedido #{openSale.number}
                              </span>
                              <span className="mt-1 block text-xs text-(--ink-soft)">
                                Aberto em {formatDate(openSale.openedAt)}
                              </span>
                            </span>
                            <span className="text-right">
                              <strong className="block text-sm">
                                {formatCurrency(openSale.total)}
                              </strong>
                              <span className="mt-1 block text-xs text-(--coral-dark)">
                                Retomar
                              </span>
                            </span>
                          </button>
                        ))}
                      </div>
                      {selectedOpenSaleQuery.isLoading && (
                        <p className="mt-4 text-center text-xs text-(--ink-soft)">
                          Abrindo pedido...
                        </p>
                      )}
                      {selectedOpenSaleQuery.error && (
                        <p className="mt-4 text-xs text-(--coral-dark)">
                          {selectedOpenSaleQuery.error.message}
                        </p>
                      )}
                    </div>
                    <Sheet.Footer className=" bg-(--surface) px-6 py-5">
                      <p className="w-full text-xs text-(--ink-soft)">
                        {openSalesQuery.data?.data.length ?? 0} pedido(s)
                        aguardando atendimento.
                      </p>
                    </Sheet.Footer>
                  </Sheet.Content>
                </Sheet>
              </div>
            </div>
          }
        />
        <div className="grid gap-6">
          <Card className="min-h-127.5 rounded-2xl border-(--line) bg-(--surface)">
            <Card.Header className=" px-6 py-5">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <Card.Title className="font-serif text-2xl">
                    Catálogo
                  </Card.Title>
                  <Card.Description className="mt-1">
                    Produtos disponíveis para venda
                  </Card.Description>
                </div>
              </div>
            </Card.Header>
            <Card.Content className="grid grid-cols-2 gap-3 p-6 sm:grid-cols-3">
              {catalogQuery.error && (
                <p className="col-span-full text-sm text-(--coral-dark)">
                  {catalogQuery.error.message}
                </p>
              )}
              {catalogQuery.isLoading && (
                <p className="col-span-full text-sm text-(--ink-soft)">
                  Carregando catálogo...
                </p>
              )}
              {!catalogQuery.isLoading &&
                !catalogQuery.error &&
                products.length === 0 && (
                  <p className="col-span-full text-sm text-(--ink-soft)">
                    Nenhum produto disponível.
                  </p>
                )}
              {products.map((product, index) => (
                <ProductTile
                  key={product.id}
                  product={product}
                  tone={tones[index % tones.length]}
                  quantity={productQuantity(product.id)}
                  onAdd={() => increaseProduct(product)}
                  onRemove={() => decreaseProduct(product)}
                  isDisabled={isBusy}
                />
              ))}
            </Card.Content>
          </Card>
        </div>
        <div className="mt-5 text-center text-xs text-(--ink-soft)">
          <Link to="/" className="hover:text-(--coral-dark)">
            ← Voltar para visão geral
          </Link>
        </div>
        <Sheet.Content
          side="right"
          className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
        >
          <Sheet.Header className="px-6 py-5">
            <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
              Pedido atual
            </p>
            <Sheet.Title className="font-serif text-2xl text-(--ink)">
              {sale ? `Pedido #${sale.number}` : "Carrinho vazio"}
            </Sheet.Title>
            <Sheet.Description className="text-(--ink-soft)">
              Revise os produtos antes de finalizar o atendimento.
            </Sheet.Description>
          </Sheet.Header>
          <div className="flex-1 overflow-y-auto px-6 py-5">
            {!sale || sale.items.length === 0 ? (
              <div className="flex h-full flex-col items-center justify-center text-center">
                <div className="mb-4 grid size-16 place-items-center rounded-full border border-dashed border-(--line) text-(--ink-soft)/40">
                  <Plus className="size-6" />
                </div>
                <p className="text-sm font-semibold text-(--ink)">
                  Adicione o primeiro produto
                </p>
                <p className="mt-2 max-w-55 text-xs leading-relaxed text-(--ink-soft)">
                  Use a busca acima ou escaneie um código para começar o
                  atendimento.
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {sale.items.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center gap-3 rounded-md border border-(--line) bg-(--contrast-light)/60 p-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold">
                        {item.productName}
                      </p>
                      <p className="mt-1 text-xs text-(--ink-soft)">
                        {item.quantity} × {formatCurrency(item.unitPrice)}
                      </p>
                    </div>
                    <strong className="text-sm">
                      {formatCurrency(item.total)}
                    </strong>
                    <button
                      type="button"
                      aria-label={`Remover ${item.productName}`}
                      onClick={() => removeProduct(item.id)}
                      disabled={isBusy}
                      className="rounded-md px-1 text-lg leading-none text-(--ink-soft) hover:bg-(--coral-wash) hover:text-(--coral-dark) disabled:pointer-events-none disabled:opacity-50"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
          <Sheet.Footer className="bg-(--surface) px-6 py-5">
            <div className="w-full">
              <div className="mb-4 flex justify-between text-sm">
                <span className="text-(--ink-soft)">Subtotal</span>
                <strong>{formatCurrency(sale?.total ?? "0")}</strong>
              </div>
              {feedback && (
                <p className="mb-3 rounded-md bg-(--coral-wash) px-3 py-2 text-xs text-(--coral-dark)">
                  {feedback}
                </p>
              )}
              <Button
                disabled={!sale || sale.items.length === 0 || isBusy}
                onClick={handleCheckout}
                className="h-10 w-full rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)"
              >
                {checkoutSale.isPending
                  ? "Finalizando..."
                  : "Ir para pagamento"}
              </Button>
            </div>
          </Sheet.Footer>
        </Sheet.Content>
      </div>
    </Sheet>
  );
}

function ShortcutRow({ keys, label }: { keys: string; label: string }) {
  return (
    <div className="flex items-center justify-between gap-4 px-3 py-3 text-sm">
      <span className="text-(--ink-soft)">{label}</span>
      <kbd className="shrink-0 rounded-md border border-(--line) bg-(--surface) px-2 py-1 font-mono text-[11px] font-medium text-(--ink)">
        {keys}
      </kbd>
    </div>
  );
}
