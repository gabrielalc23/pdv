import { useState } from "react";
import type { JSX } from "react";
import { ArrowUpRight, Receipt, Search, Sliders } from "lucide-react";
import { Badge } from "@pdv/ui-kit/components/badge";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Input } from "@pdv/ui-kit/components/input";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { PageHeader } from "../components/page-header";
import { useListSalesQuery } from "../queries/sale.query";
import { formatDate } from "../utils/format-date.util";
import { formatCurrency } from "../utils/format-currency.util";
import { statusLabel } from "../utils/format-label.util";
import type {
  SaleListItemResponse,
  SaleListResponse,
} from "../interfaces/sale.interface";
import type { SaleStatus } from "../types/sale.type";
import type { UseQueryResult } from "@tanstack/react-query";
import { SectionLabel } from "../components/section-label";
import { useSheet } from "../hooks/use-sheet.hook";

export function SalesPage(): JSX.Element {
  const [search, setSearch] = useState<string>("");
  const [statusFilter, setStatusFilter] = useState<SaleStatus | "ALL">("ALL");
  const filtersSheet = useSheet("filtros-vendas");

  const salesQuery: UseQueryResult<SaleListResponse> = useListSalesQuery({
    page: 1,
    pageSize: 50,
  });

  const sales: SaleListItemResponse[] = (salesQuery.data?.data ?? []).filter(
    (sale: SaleListItemResponse): boolean =>
      (String(sale.number).includes(search) || !search) &&
      (statusFilter === "ALL" || sale.status === statusFilter),
  );

  const completedSales: SaleListItemResponse[] = sales.filter(
    (sale: SaleListItemResponse) => sale.status === "COMPLETED",
  );
  const openSales: SaleListItemResponse[] = sales.filter(
    (sale: SaleListItemResponse) => sale.status === "OPEN",
  );
  const cancelledSales: SaleListItemResponse[] = sales.filter(
    (sale: SaleListItemResponse) => sale.status === "CANCELLED",
  );
  const completedTotal: number = completedSales.reduce(
    (total: number, sale: SaleListItemResponse) => total + Number(sale.total),
    0,
  );

  return (
    <div>
      <PageHeader
        breadcrumbs={[{ label: "Visão geral", to: "/" }, { label: "Vendas" }]}
        title="Vendas"
        description="Consulte atendimentos, pagamentos e documentos fiscais em um só lugar."
        action={
          <Button className="rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)">
            <ArrowUpRight /> Exportar relatório
          </Button>
        }
      />
      <div className="mb-6 grid gap-4 md:grid-cols-3">
        <MiniStat
          label="Concluídas"
          value={`${completedSales.length} vendas`}
          extra={formatCurrency(String(completedTotal))}
        />
        <MiniStat
          label="Em aberto"
          value={`${openSales.length} pedidos`}
          extra="Aguardando checkout"
        />
        <MiniStat
          label="Canceladas"
          value={`${cancelledSales.length} vendas`}
          extra="No período carregado"
        />
      </div>
      <SectionLabel
        action={
          <Sheet open={filtersSheet.isOpen} onOpenChange={filtersSheet.setOpen}>
            <Sheet.Trigger
              render={<Button variant="outline" className="border-(--line)" />}
            >
              <Sliders /> Filtros
            </Sheet.Trigger>
            <Sheet.Content
              side="right"
              className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
            >
              <Sheet.Header className="border-b border-(--line) px-6 py-5">
                <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                  Vendas
                </p>
                <Sheet.Title className="font-serif text-2xl">
                  Filtrar vendas
                </Sheet.Title>
                <Sheet.Description className="text-(--ink-soft)">
                  Mostre apenas vendas com o status escolhido.
                </Sheet.Description>
              </Sheet.Header>
              <div className="flex-1 px-6 py-5">
                <label
                  htmlFor="sales-status-filter"
                  className="text-sm font-semibold"
                >
                  Status da venda
                </label>
                <select
                  id="sales-status-filter"
                  value={statusFilter}
                  onChange={(event) =>
                    setStatusFilter(event.target.value as SaleStatus | "ALL")
                  }
                  className="mt-2 h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                >
                  <option value="ALL">Todas</option>
                  <option value="OPEN">Em aberto</option>
                  <option value="COMPLETED">Concluídas</option>
                  <option value="CANCELLED">Canceladas</option>
                </select>
              </div>
              <Sheet.Footer className="border-t border-(--line) bg-(--surface) px-6 py-5">
                <Button
                  type="button"
                  variant="outline"
                  className="w-full border-(--line)"
                  onClick={() => setStatusFilter("ALL")}
                >
                  Limpar filtros
                </Button>
              </Sheet.Footer>
            </Sheet.Content>
          </Sheet>
        }
      >
        Todas as vendas
      </SectionLabel>
      <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface)">
        <div className="flex items-center gap-3 px-5 py-3 text-xs text-(--ink-soft)">
          <Search className="size-4" />
          <Input
            className="border-0 bg-(--transparent) px-0 shadow-none focus-visible:ring-0"
            placeholder="Buscar por número..."
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="divide-y divide-(--line)">
          {salesQuery.error && (
            <p className="px-5 py-8 text-sm text-(--coral-dark)">
              {salesQuery.error.message}
            </p>
          )}
          {salesQuery.isLoading && (
            <p className="px-5 py-8 text-sm text-(--ink-soft)">
              Carregando vendas...
            </p>
          )}
          {!salesQuery.isLoading && !salesQuery.error && sales.length === 0 && (
            <p className="px-5 py-8 text-sm text-(--ink-soft)">
              Nenhuma venda encontrada.
            </p>
          )}
          {sales.map((sale) => (
            <SaleRow
              key={sale.id}
              id={`#${sale.number}`}
              time={formatDate(sale.openedAt)}
              items="Detalhes disponíveis"
              amount={formatCurrency(sale.total)}
              status={statusLabel(sale.status)}
            />
          ))}
        </div>
      </Card>
    </div>
  );
}

function MiniStat({
  label,
  value,
  extra,
}: {
  label: string;
  value: string;
  extra: string;
}) {
  return (
    <Card className="rounded-2xl border-(--line) bg-(--surface) shadow-none">
      <Card.Content className="p-5">
        <p className="text-xs text-(--ink-soft)">{label}</p>
        <p className="mt-2 font-serif text-2xl font-semibold">{value}</p>
        <p className="mt-1 text-xs text-(--coral-dark)">{extra}</p>
      </Card.Content>
    </Card>
  );
}
function SaleRow({
  id,
  time,
  items,
  amount,
  status,
}: {
  id: string;
  time: string;
  items: string;
  amount: string;
  status: string;
}) {
  return (
    <div className="flex items-center gap-4 px-5 py-4">
      <span className="grid size-9 place-items-center rounded-md bg-(--coral-wash) text-(--coral-dark)">
        <Receipt />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold">Venda {id}</p>
        <p className="mt-0.5 text-xs text-(--ink-soft)">
          {time} · {items}
        </p>
      </div>
      <Badge
        variant={status === "Em aberto" ? "outline" : "secondary"}
        className={
          status === "Em aberto"
            ? "border-(--coral-border) text-(--coral-dark)"
            : "bg-(--mint) text-(--mint-foreground)"
        }
      >
        {status}
      </Badge>
      <p className="hidden w-24 text-right text-sm font-semibold sm:block">
        {amount}
      </p>
      <ArrowUpRight className="text-(--neutral-dark)/25" />
    </div>
  );
}
