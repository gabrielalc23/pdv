import { useState } from "react";
import type { ChangeEvent, FC, JSX } from "react";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Input } from "@pdv/ui-kit/components/input";
import { CatalogCard } from "../components/catalog-card";
import { PageHeader } from "../components/page-header";
import { useListCatalogQuery } from "../queries/catalog.query";
import { useListCategoriesQuery } from "../queries/category.query";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { Search, Sliders } from "lucide-react";
import { formatCurrency } from "../utils/format-currency.util";
import type { UseQueryResult } from "@tanstack/react-query";
import type {
  CatalogListResponse,
  CatalogProductResponse,
} from "../interfaces/catalog.interface";
import { useSheet } from "../hooks/use-sheet.hook";

const tones: readonly ["coral", "mint", "blue", "sand"] = [
  "coral",
  "mint",
  "blue",
  "sand",
] as const;

export const CatalogPage: FC = (): JSX.Element => {
  const [search, setSearch] = useState<string>("");
  const [availabilityFilter, setAvailabilityFilter] = useState<
    "all" | "in-stock" | "out-of-stock"
  >("all");
  const [categoryFilter, setCategoryFilter] = useState<string>("");
  const filtersSheet = useSheet("filtros-catalogo");

  const catalogQuery: UseQueryResult<CatalogListResponse> = useListCatalogQuery(
    {
      search: search || undefined,
      activeOnly: true,
      inStockOnly: availabilityFilter === "in-stock",
      categoryId: categoryFilter || undefined,
      page: 1,
      pageSize: 50,
    },
  );
  const categoriesQuery = useListCategoriesQuery({ activeOnly: true });

  const products: CatalogProductResponse[] = (
    catalogQuery.data?.data ?? []
  ).filter(
    (product) => availabilityFilter !== "out-of-stock" || !product.inStock,
  );

  function handleSearch(
    event: ChangeEvent<HTMLInputElement, HTMLInputElement>,
  ): void {
    setSearch(event.target.value);
  }

  return (
    <div>
      <PageHeader
        breadcrumbs={[{ label: "Visão geral", to: "/" }, { label: "Catálogo" }]}
        title="Catálogo"
        description="A mesma visão de produtos que alimenta o caixa, pronta para consulta rápida."
        action={
          <Sheet open={filtersSheet.isOpen} onOpenChange={filtersSheet.setOpen}>
            <Sheet.Trigger
              render={
                <Button
                  variant="outline"
                  className="rounded-md border-(--line)"
                />
              }
            >
              <Sliders className="size-4" /> Filtros
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
                  Filtrar catálogo
                </Sheet.Title>
                <Sheet.Description className="text-(--ink-soft)">
                  Encontre rapidamente produtos disponíveis para venda.
                </Sheet.Description>
              </Sheet.Header>
              <div className="flex-1 px-6 py-5">
                <label
                  htmlFor="catalog-availability-filter"
                  className="text-sm font-semibold"
                >
                  Disponibilidade
                </label>
                <select
                  id="catalog-availability-filter"
                  value={availabilityFilter}
                  onChange={(event) =>
                    setAvailabilityFilter(
                      event.target.value as typeof availabilityFilter,
                    )
                  }
                  className="mt-2 h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                >
                  <option value="all">Todos os produtos</option>
                  <option value="in-stock">Em estoque</option>
                  <option value="out-of-stock">Sem estoque</option>
                </select>
                <label
                  htmlFor="catalog-category-filter"
                  className="mt-5 block text-sm font-semibold"
                >
                  Categoria
                </label>
                <select
                  id="catalog-category-filter"
                  value={categoryFilter}
                  onChange={(event) => setCategoryFilter(event.target.value)}
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
                    setAvailabilityFilter("all");
                    setCategoryFilter("");
                  }}
                >
                  Limpar filtros
                </Button>
              </Sheet.Footer>
            </Sheet.Content>
          </Sheet>
        }
      />
      <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface)">
        <div className="flex items-center gap-3 px-5 py-3 text-xs text-(--ink-soft)">
          <Search className="size-4" />
          <Input
            className="border-0 bg-(--transparent) px-0 shadow-none focus-visible:ring-0"
            placeholder="Buscar produto por nome, SKU ou código de barras"
            value={search}
            onChange={handleSearch}
          />
        </div>
        <Card.Content className="p-6">
          {catalogQuery.error && (
            <p className="mb-4 text-sm text-(--coral-dark)">
              {catalogQuery.error.message}
            </p>
          )}
          {catalogQuery.isLoading && (
            <p className="mb-4 text-sm text-(--ink-soft)">
              Carregando catálogo...
            </p>
          )}
          {!catalogQuery.isLoading &&
            !catalogQuery.error &&
            products.length === 0 && (
              <p className="mb-4 text-sm text-(--ink-soft)">
                Nenhum produto encontrado.
              </p>
            )}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {products.map(
              (product: CatalogProductResponse, index: number): JSX.Element => (
                <CatalogCard
                  key={product.id}
                  name={product.name}
                  category={`SKU ${product.sku} · ${product.inStock ? `${product.quantity} un.` : "Sem estoque"}`}
                  price={formatCurrency(product.price)}
                  tone={tones[index % tones.length]}
                />
              ),
            )}
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};
