import { useState } from "react";
import { ArrowUpRight, Package, Plus, Search, Sliders } from "lucide-react";
import { Badge } from "@pdv/ui-kit/components/badge";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Input } from "@pdv/ui-kit/components/input";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { PageHeader } from "../components/page-header";
import {
  useActivateProductMutation,
  useCreateProductMutation,
  useDeactivateProductMutation,
} from "../mutations/product.mutation";
import { useListProductsQuery } from "../queries/product.query";
import { useListCategoriesQuery } from "../queries/category.query";
import type { ProductResponse } from "../interfaces/product.interface";
import { formatCurrency } from "../utils/format-currency.util";
import { useSheet } from "../hooks/use-sheet.hook";

const tones = ["coral", "mint", "blue", "sand"] as const;

export function ProductsPage() {
  const [search, setSearch] = useState("");
  const [productFilter, setProductFilter] = useState<
    "all" | "active" | "inactive"
  >("all");
  const [categoryFilter, setCategoryFilter] = useState<string>("");
  const [form, setForm] = useState({
    sku: "",
    barcode: "",
    name: "",
    price: "",
    cost: "",
    categoryId: "",
  });
  const productSheet = useSheet("novo-produto");
  const productFiltersSheet = useSheet("filtros-produtos");
  const productsQuery = useListProductsQuery({
    search: search || undefined,
    page: 1,
    pageSize: 50,
    activeOnly: productFilter === "active" ? true : undefined,
    categoryId: categoryFilter || undefined,
  });
  const categoriesQuery = useListCategoriesQuery({ activeOnly: true });
  const createProduct = useCreateProductMutation();
  const activateProduct = useActivateProductMutation();
  const deactivateProduct = useDeactivateProductMutation();
  const products = (productsQuery.data?.data ?? []).filter(
    (product) => productFilter !== "inactive" || !product.isActive,
  );

  function updateForm(field: keyof typeof form, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    createProduct.mutate(
      {
        sku: form.sku,
        barcode: form.barcode || null,
        name: form.name,
        price: form.price,
        cost: form.cost || null,
        categoryId: form.categoryId || null,
      },
      {
        onSuccess: () => {
          setForm({
            sku: "",
            barcode: "",
            name: "",
            price: "",
            cost: "",
            categoryId: "",
          });
          productSheet.setOpen(false);
        },
      },
    );
  }

  return (
    <Sheet open={productSheet.isOpen} onOpenChange={productSheet.setOpen}>
      <div>
        <PageHeader
          breadcrumbs={[
            { label: "Visão geral", to: "/" },
            { label: "Produtos" },
          ]}
          title="Produtos"
          description="Mantenha preços, códigos e disponibilidade do seu catálogo sempre atualizados."
          action={
            <div className="flex gap-2">
              <Sheet.Trigger
                render={
                  <Button className="rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)" />
                }
              >
                <Plus /> Novo produto
              </Sheet.Trigger>
              <Sheet
                open={productFiltersSheet.isOpen}
                onOpenChange={productFiltersSheet.setOpen}
              >
                <Sheet.Trigger
                  render={
                    <Button variant="outline" className="border-(--line)" />
                  }
                >
                  <Sliders /> Filtros
                </Sheet.Trigger>
                <Sheet.Content
                  side="right"
                  className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
                >
                  <Sheet.Header className="border-b border-(--line) px-6 py-5">
                    <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                      Catálogo
                    </p>
                    <Sheet.Title className="font-serif text-2xl">
                      Filtrar produtos
                    </Sheet.Title>
                    <Sheet.Description className="text-(--ink-soft)">
                      Escolha quais produtos devem aparecer na lista.
                    </Sheet.Description>
                  </Sheet.Header>
                  <div className="flex-1 px-6 py-5">
                    <label
                      htmlFor="product-status-filter"
                      className="text-sm font-semibold"
                    >
                      Status
                    </label>
                    <select
                      id="product-status-filter"
                      value={productFilter}
                      onChange={(event) =>
                        setProductFilter(
                          event.target.value as typeof productFilter,
                        )
                      }
                      className="mt-2 h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                    >
                      <option value="all">Todos os produtos</option>
                      <option value="active">Somente ativos</option>
                      <option value="inactive">Somente inativos</option>
                    </select>
                    <label
                      htmlFor="product-category-filter"
                      className="mt-5 block text-sm font-semibold"
                    >
                      Categoria
                    </label>
                    <select
                      id="product-category-filter"
                      value={categoryFilter}
                      onChange={(event) =>
                        setCategoryFilter(event.target.value)
                      }
                      className="mt-2 h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                    >
                      <option value="">Todas as categorias</option>
                      {categoriesQuery.data?.data.map((category) => (
                        <option key={category.id} value={category.id}>
                          {category.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <Sheet.Footer className="border-t border-(--line) bg-(--surface) px-6 py-5">
                    <Button
                      type="button"
                      variant="outline"
                      className="w-full border-(--line)"
                      onClick={() => {
                        setProductFilter("all");
                        setCategoryFilter("");
                      }}
                    >
                      Limpar filtros
                    </Button>
                  </Sheet.Footer>
                </Sheet.Content>
              </Sheet>
            </div>
          }
        />
        <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface)">
          <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-3">
            <div className="flex min-w-60 flex-1 items-center gap-2 text-sm text-(--ink-soft)">
              <Search className="size-4" />
              <Input
                className="border-0 bg-(--transparent) px-0 shadow-none focus-visible:ring-0"
                placeholder="Buscar por nome ou SKU..."
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
            </div>
            <span className="text-xs text-(--ink-soft)">
              {productsQuery.isLoading
                ? "Carregando..."
                : `${productsQuery.data?.pagination.total ?? 0} produtos`}
            </span>
          </div>
          <div className="divide-y divide-(--line)">
            {productsQuery.error && (
              <p className="px-5 py-8 text-sm text-(--coral-dark)">
                {productsQuery.error.message}
              </p>
            )}
            {!productsQuery.isLoading &&
              !productsQuery.error &&
              products.length === 0 && (
                <p className="px-5 py-8 text-sm text-(--ink-soft)">
                  Nenhum produto encontrado.
                </p>
              )}
            {products.map((product, index) => (
              <ProductRow
                key={product.id}
                product={product}
                tone={tones[index % tones.length]}
                onToggle={() =>
                  product.isActive
                    ? deactivateProduct.mutate(product.id)
                    : activateProduct.mutate(product.id)
                }
                isPending={
                  activateProduct.isPending || deactivateProduct.isPending
                }
              />
            ))}
          </div>
        </Card>

        <Sheet.Content
          side="right"
          className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-lg"
        >
          <Sheet.Header className=" px-6 py-6">
            <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
              Catálogo
            </p>
            <Sheet.Title className="font-serif text-3xl">
              Novo produto
            </Sheet.Title>
            <Sheet.Description className="mt-2 text-(--ink-soft)">
              Cadastre os dados comerciais e defina como este item aparecerá no
              caixa.
            </Sheet.Description>
          </Sheet.Header>
          <form
            className="flex min-h-0 flex-1 flex-col"
            onSubmit={handleSubmit}
          >
            <div className="flex-1 space-y-5 overflow-y-auto px-6 py-6">
              <div className="space-y-2">
                <label htmlFor="product-name" className="text-sm font-semibold">
                  Nome do produto
                </label>
                <Input
                  id="product-name"
                  required
                  placeholder="Ex.: Café especial 250g"
                  value={form.name}
                  onChange={(event) => updateForm("name", event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label
                  htmlFor="product-category"
                  className="text-sm font-semibold"
                >
                  Categoria
                </label>
                <select
                  id="product-category"
                  value={form.categoryId}
                  onChange={(event) =>
                    updateForm("categoryId", event.target.value)
                  }
                  className="h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                >
                  <option value="">Sem categoria</option>
                  {categoriesQuery.data?.data.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid gap-5 sm:grid-cols-2">
                <div className="space-y-2">
                  <label
                    htmlFor="product-sku"
                    className="text-sm font-semibold"
                  >
                    SKU
                  </label>
                  <Input
                    id="product-sku"
                    required
                    placeholder="CAF-250"
                    value={form.sku}
                    onChange={(event) => updateForm("sku", event.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <label
                    htmlFor="product-barcode"
                    className="text-sm font-semibold"
                  >
                    Código de barras
                  </label>
                  <Input
                    id="product-barcode"
                    placeholder="Opcional"
                    value={form.barcode}
                    onChange={(event) =>
                      updateForm("barcode", event.target.value)
                    }
                  />
                </div>
              </div>
              <div className="grid gap-5 sm:grid-cols-2">
                <div className="space-y-2">
                  <label
                    htmlFor="product-price"
                    className="text-sm font-semibold"
                  >
                    Preço de venda
                  </label>
                  <Input
                    id="product-price"
                    required
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0,00"
                    value={form.price}
                    onChange={(event) =>
                      updateForm("price", event.target.value)
                    }
                  />
                </div>
                <div className="space-y-2">
                  <label
                    htmlFor="product-cost"
                    className="text-sm font-semibold"
                  >
                    Custo
                  </label>
                  <Input
                    id="product-cost"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="Opcional"
                    value={form.cost}
                    onChange={(event) => updateForm("cost", event.target.value)}
                  />
                </div>
              </div>
              {createProduct.error && (
                <p className="rounded-md bg-(--coral-wash) px-3 py-2 text-xs text-(--coral-dark)">
                  {createProduct.error.message}
                </p>
              )}
            </div>
            <Sheet.Footer className=" bg-(--surface) px-6 py-5">
              <div className="flex w-full gap-3">
                <Sheet.Close
                  render={
                    <Button
                      type="button"
                      variant="outline"
                      className="flex-1 border-(--line)"
                    />
                  }
                >
                  Cancelar
                </Sheet.Close>
                <Button
                  type="submit"
                  disabled={createProduct.isPending}
                  className="flex-1 rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)"
                >
                  {createProduct.isPending ? "Salvando..." : "Salvar produto"}
                </Button>
              </div>
            </Sheet.Footer>
          </form>
        </Sheet.Content>
      </div>
    </Sheet>
  );
}

function ProductRow({
  product,
  tone,
  onToggle,
  isPending,
}: {
  product: ProductResponse;
  tone: "coral" | "mint" | "blue" | "sand";
  onToggle: () => void;
  isPending: boolean;
}) {
  const bg = {
    coral: "bg-(--coral-wash)",
    mint: "bg-(--mint)",
    blue: "bg-(--blue-wash)",
    sand: "bg-(--sand)",
  }[tone];
  return (
    <div className="flex flex-wrap items-center gap-4 px-5 py-4 sm:flex-nowrap">
      <span
        className={`grid size-10 shrink-0 place-items-center rounded-md ${bg}`}
      >
        <Package className="text-(--blue-foreground)/60" />
      </span>
      <div className="min-w-45 flex-1">
        <p className="text-sm font-semibold">{product.name}</p>
        <p className="mt-0.5 text-xs text-(--ink-soft)">
          SKU {product.sku} {product.barcode ? `· ${product.barcode}` : ""}
        </p>
      </div>
      <div className="w-24 text-sm font-semibold">
        {formatCurrency(product.price)}
        <span className="mt-0.5 block text-[11px] font-normal text-(--ink-soft)">
          preço
        </span>
      </div>
      <div className="w-24 text-sm">
        —
        <span className="mt-0.5 block text-[11px] text-(--ink-soft)">
          no estoque
        </span>
      </div>
      <Badge
        variant={product.isActive ? "default" : "secondary"}
        className={
          product.isActive ? "bg-(--mint) text-(--mint-foreground)" : ""
        }
      >
        {product.isActive ? "Ativo" : "Inativo"}
      </Badge>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        disabled={isPending}
        onClick={onToggle}
      >
        {product.isActive ? "Desativar" : "Ativar"}
      </Button>
      <button
        type="button"
        aria-label={`Abrir ${product.name}`}
        className="rounded-md p-2 text-(--ink-soft) hover:bg-(--neutral-dark)/5"
      >
        <ArrowUpRight />
      </button>
    </div>
  );
}
