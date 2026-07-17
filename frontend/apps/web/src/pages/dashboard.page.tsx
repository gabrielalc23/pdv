import {
  Activity,
  ArrowUpRight,
  Box,
  Grid2X2,
  Plus,
  Receipt,
  Search,
  ShoppingBag,
} from "lucide-react";
import { Badge } from "@pdv/ui-kit/components/badge";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { Link } from "@tanstack/react-router";

import { PageHeader } from "../components/page-header";
import { Metric } from "../components/metric";
import { useListInventoryQuery } from "../queries/inventory.query";
import { useListSalesQuery } from "../queries/sale.query";
import { formatCurrency } from "../utils/format-currency.util";
import { QuickAction } from "../components/quick-action";
import { ActivityRow } from "../components/activity-row";
import { SectionLabel } from "../components/section-label";
import { useSheet } from "../hooks/use-sheet.hook";

export function DashboardPage() {
  const salesQuery = useListSalesQuery({ page: 1, pageSize: 50 });
  const inventoryQuery = useListInventoryQuery({ page: 1, pageSize: 100 });
  const completedSales =
    salesQuery.data?.data.filter((sale) => sale.status === "COMPLETED") ?? [];
  const completedTotal = completedSales.reduce(
    (total, sale) => total + Number(sale.total),
    0,
  );
  const inventoryTotal =
    inventoryQuery.data?.data.reduce(
      (total, item) => total + Number(item.quantity),
      0,
    ) ?? 0;
  const newSaleSheet = useSheet("nova-venda");

  return (
    <div className="animate-[fade-in_500ms_ease-out_both]">
      <PageHeader
        breadcrumbs={[
          { label: "Workspace", to: "/" },
          { label: "Visão geral" },
        ]}
        title={
          <>
            O balcão está
            <br />
            <em className="text-(--coral)">aberto.</em>
          </>
        }
        description="Acompanhe a operação de hoje, retome atendimentos e tenha o essencial sempre à mão."
        action={
          <Sheet open={newSaleSheet.isOpen} onOpenChange={newSaleSheet.setOpen}>
            <Sheet.Trigger
              render={
                <Button className="h-11 rounded-md bg-(--coral) px-5 text-(--coral-foreground) hover:bg-(--coral-dark)" />
              }
            >
              <Plus className="size-4" />
              Nova venda
            </Sheet.Trigger>
            <Sheet.Content
              side="right"
              className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
            >
              <Sheet.Header className="px-6 py-5">
                <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                  Novo atendimento
                </p>
                <Sheet.Title className="font-serif text-2xl text-(--ink)">
                  Começar uma nova venda
                </Sheet.Title>
                <Sheet.Description className="text-(--ink-soft)">
                  Escolha como deseja iniciar o atendimento do cliente.
                </Sheet.Description>
              </Sheet.Header>
              <div className="flex-1 px-6 py-5">
                <div className="space-y-3">
                  <Sheet.Close
                    render={
                      <Link
                        to="/pos"
                        className="group flex items-center gap-4 rounded-md border border-(--coral-border) bg-(--coral-wash) p-4 transition-colors hover:bg-(--coral) hover:text-(--contrast-light)"
                      />
                    }
                  >
                    <span className="grid size-10 shrink-0 place-items-center rounded-md bg-(--coral) text-(--contrast-light) transition-colors group-hover:bg-(--contrast-light)/15">
                      <ShoppingBag className="size-4" />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-semibold">
                        Abrir o PDV
                      </span>
                      <span className="mt-1 block text-xs opacity-70">
                        Catálogo e carrinho lado a lado
                      </span>
                    </span>
                    <ArrowUpRight className="size-4 shrink-0" />
                  </Sheet.Close>
                  <Sheet.Close
                    render={
                      <Link
                        to="/catalog"
                        className="group flex items-center gap-4 rounded-md border border-(--line) bg-(--paper) p-4 transition-colors hover:border-(--coral-border) hover:bg-(--coral-wash)"
                      />
                    }
                  >
                    <span className="grid size-10 shrink-0 place-items-center rounded-md bg-(--mint) text-(--mint-foreground)">
                      <Grid2X2 className="size-4" />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-semibold">
                        Consultar catálogo
                      </span>
                      <span className="mt-1 block text-xs text-(--ink-soft)">
                        Encontre produtos antes de iniciar
                      </span>
                    </span>
                    <ArrowUpRight className="size-4 shrink-0 text-(--ink-soft)" />
                  </Sheet.Close>
                </div>
              </div>
              <Sheet.Footer className="bg-(--surface) px-6 py-5">
                <p className="w-full text-xs text-(--ink-soft)">
                  Você poderá adicionar produtos e revisar o pedido dentro do
                  PDV.
                </p>
              </Sheet.Footer>
            </Sheet.Content>
          </Sheet>
        }
      />

      <div className="mb-10 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Metric
          label="Vendas hoje"
          value={
            salesQuery.isLoading
              ? "..."
              : formatCurrency(String(completedTotal))
          }
          change={`${completedSales.length} concluídas`}
          tone="coral"
          icon={Receipt}
        />

        <Metric
          label="Atendimentos"
          value={
            salesQuery.isLoading
              ? "..."
              : String(salesQuery.data?.pagination.total ?? 0)
          }
          change="no período"
          tone="mint"
          icon={Activity}
        />

        <Metric
          label="Ticket médio"
          value={
            completedSales.length
              ? formatCurrency(String(completedTotal / completedSales.length))
              : "R$ 0,00"
          }
          change="ticket médio"
          tone="blue"
          icon={ShoppingBag}
        />

        <Metric
          label="Itens em estoque"
          value={inventoryQuery.isLoading ? "..." : String(inventoryTotal)}
          change={`${inventoryQuery.data?.data.filter((item) => Number(item.quantity) <= 0).length ?? 0} sem estoque`}
          tone="sand"
          icon={Box}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.35fr_0.65fr]">
        <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface) shadow-(--shadow)">
          <Card.Header className="flex flex-row items-start justify-between  px-6 py-5">
            <div>
              <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                Atalho de operação
              </p>

              <Card.Title className="font-serif text-2xl">
                Comece um novo atendimento
              </Card.Title>

              <Card.Description className="mt-1 text-sm text-(--ink-soft)">
                Busque um produto ou leia um código de barras para iniciar.
              </Card.Description>
            </div>

            <div className="grid size-11 place-items-center rounded-md bg-(--mint) text-(--mint-foreground)">
              <ShoppingBag className="size-5" />
            </div>
          </Card.Header>

          <Card.Content className="grid gap-3 p-6 sm:grid-cols-2">
            <QuickAction
              icon={Search}
              title="Buscar no catálogo"
              description="Por nome, SKU ou categoria"
            />

            <QuickAction
              icon={Grid2X2}
              title="Abrir PDV completo"
              description="Catálogo e carrinho lado a lado"
            />
          </Card.Content>
        </Card>

        <Card className="rounded-2xl border-(--mint-border) bg-(--mint) shadow-none">
          <Card.Header className="px-6 pt-6 pb-3">
            <div className="mb-4 flex items-center justify-between">
              <Badge className="bg-(--mint-foreground) text-(--contrast-light)">
                Operação saudável
              </Badge>

              <Activity className="size-5 text-(--mint-foreground)" />
            </div>

            <Card.Title className="font-serif text-2xl text-(--mint-deep)">
              Tudo sob controle.
            </Card.Title>

            <Card.Description className="mt-2 text-sm leading-relaxed text-(--mint-muted)">
              Nenhum alerta crítico no momento. Existem 12 produtos com estoque
              abaixo do mínimo.
            </Card.Description>
          </Card.Header>

          <Card.Footer className="px-6 pt-3 pb-6">
            <Button
              variant="outline"
              className="border-(--mint-accent-border) bg-(--transparent) text-(--mint-accent) hover:bg-(--contrast-light)/50"
            >
              Ver alertas
              <ArrowUpRight className="size-4" />
            </Button>
          </Card.Footer>
        </Card>
      </div>

      <div className="mt-10">
        <SectionLabel
          action={
            <Button variant="ghost" size="sm" className="text-(--coral-dark)">
              Ver todas
              <ArrowUpRight className="size-4" />
            </Button>
          }
        >
          Movimentações recentes
        </SectionLabel>

        <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface)">
          <div className="divide-y divide-(--line)">
            <ActivityRow
              title="Venda #1048 concluída"
              detail="3 itens · cartão de crédito"
              amount="R$ 184,90"
              time="há 4 min"
              tone="coral"
            />

            <ActivityRow
              title="Entrada de estoque registrada"
              detail="Café especial 250g · +24 un."
              amount="Estoque"
              time="há 18 min"
              tone="mint"
            />

            <ActivityRow
              title="Venda #1047 concluída"
              detail="1 item · PIX"
              amount="R$ 32,00"
              time="há 31 min"
              tone="blue"
            />
          </div>
        </Card>
      </div>
    </div>
  );
}
