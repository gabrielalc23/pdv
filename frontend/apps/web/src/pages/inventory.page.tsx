import { useState } from "react";
import { Badge } from "@pdv/ui-kit/components/badge";
import { Button } from "@pdv/ui-kit/components/button";
import { Card } from "@pdv/ui-kit/components/card";
import { Input } from "@pdv/ui-kit/components/input";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import { PageHeader } from "../components/page-header";
import {
  useCreateInventoryAdjustmentMutation,
  useCreateInventoryEntryMutation,
} from "../mutations/inventory.mutation";
import { useListInventoryQuery } from "../queries/inventory.query";
import { Box, Plus, Search, Sliders } from "lucide-react";
import type { Nullable } from "@pdv/types";
import { SectionLabel } from "../components/section-label";
import { useSheet } from "../hooks/use-sheet.hook";

export function InventoryPage() {
  const [search, setSearch] = useState<string>("");
  const [form, setForm] = useState({
    productId: "",
    quantity: "",
    reason: "",
    direction: "IN" as "IN" | "OUT",
  });
  const inventoryQuery = useListInventoryQuery({
    search: search || undefined,
    page: 1,
    pageSize: 50,
  });
  const createEntry = useCreateInventoryEntryMutation();
  const createAdjustment = useCreateInventoryAdjustmentMutation();
  const entrySheet = useSheet("entrada-estoque");
  const adjustmentSheet = useSheet("ajuste-estoque");
  const formMode: Nullable<"entry" | "adjustment"> = entrySheet.isOpen
    ? "entry"
    : adjustmentSheet.isOpen
      ? "adjustment"
      : null;
  const items = inventoryQuery.data?.data ?? [];
  const lowStock = items.filter((item) => Number(item.quantity) <= 0).length;

  function openForm(mode: "entry" | "adjustment") {
    const sheet = mode === "entry" ? entrySheet : adjustmentSheet;
    sheet.setOpen(!sheet.isOpen);
    if (!form.productId && items[0])
      setForm((current) => ({ ...current, productId: items[0].productId }));
  }

  function closeForm(): void {
    if (entrySheet.isOpen) entrySheet.setOpen(false);
    if (adjustmentSheet.isOpen) adjustmentSheet.setOpen(false);
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const referenceId = crypto.randomUUID();
    if (formMode === "entry") {
      createEntry.mutate(
        {
          productId: form.productId,
          quantity: form.quantity,
          reason: form.reason || null,
          referenceType: "WEB",
          referenceId,
        },
        { onSuccess: closeForm },
      );
    } else if (formMode === "adjustment") {
      createAdjustment.mutate(
        {
          productId: form.productId,
          direction: form.direction,
          quantity: form.quantity,
          reason: form.reason,
          referenceType: "WEB",
          referenceId,
        },
        { onSuccess: closeForm },
      );
    }
  }

  const isSaving = createEntry.isPending || createAdjustment.isPending;

  return (
    <Sheet
      open={formMode !== null}
      onOpenChange={(isOpen) => {
        if (!isOpen) closeForm();
      }}
    >
      <div>
        <PageHeader
          breadcrumbs={[
            { label: "Visão geral", to: "/" },
            { label: "Estoque" },
          ]}
          title="Estoque"
          description="Acompanhe níveis, entradas e ajustes para que a operação nunca pare."
          action={
            <div className="flex gap-2">
              <Button
                variant="outline"
                className="rounded-md border-(--line)"
                onClick={() => openForm("adjustment")}
              >
                <Sliders className="size-4" /> Ajustar
              </Button>
              <Button
                className="rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)"
                onClick={() => openForm("entry")}
              >
                <Plus className="size-4" /> Registrar entrada
              </Button>
            </div>
          }
        />
        <div className="mb-8 grid gap-4 sm:grid-cols-3">
          <StockStat
            label="Itens monitorados"
            value={
              inventoryQuery.isLoading
                ? "..."
                : String(inventoryQuery.data?.pagination.total ?? 0)
            }
            tone="blue"
          />
          <StockStat
            label="Sem estoque"
            value={inventoryQuery.isLoading ? "..." : String(lowStock)}
            tone="coral"
          />
          <StockStat label="Movimentações" value="—" tone="mint" />
        </div>
        <SectionLabel>Visão de estoque</SectionLabel>
        <Card className="overflow-hidden rounded-2xl border-(--line) bg-(--surface)">
          <div className="flex items-center gap-2  px-5 py-3 text-sm text-(--ink-soft)">
            <Search className="size-4" />
            <Input
              className="border-0 bg-(--transparent) px-0 shadow-none focus-visible:ring-0"
              placeholder="Buscar produto ou código..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <div className="divide-y divide-(--line)">
            {inventoryQuery.error && (
              <p className="px-5 py-8 text-sm text-(--coral-dark)">
                {inventoryQuery.error.message}
              </p>
            )}
            {!inventoryQuery.isLoading &&
              !inventoryQuery.error &&
              items.length === 0 && (
                <p className="px-5 py-8 text-sm text-(--ink-soft)">
                  Nenhum item encontrado.
                </p>
              )}
            {items.map((item) => {
              const isEmpty = Number(item.quantity) <= 0;
              return (
                <InventoryRow
                  key={item.productId}
                  name={item.name}
                  location={`SKU ${item.sku}`}
                  quantity={item.quantity}
                  minimum="—"
                  state={isEmpty ? "Sem estoque" : "Disponível"}
                  progress={isEmpty ? "0%" : "100%"}
                  isWarning={isEmpty}
                />
              );
            })}
          </div>
        </Card>

        <Sheet.Content
          side="right"
          className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-lg"
        >
          <Sheet.Header className=" px-6 py-6">
            <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
              Estoque
            </p>
            <Sheet.Title className="font-serif text-3xl">
              {formMode === "entry" ? "Registrar entrada" : "Ajustar estoque"}
            </Sheet.Title>
            <Sheet.Description className="mt-2 text-(--ink-soft)">
              {formMode === "entry"
                ? "Registre uma nova entrada para atualizar o saldo do produto."
                : "Corrija o saldo informando a direção e o motivo da movimentação."}
            </Sheet.Description>
          </Sheet.Header>
          <form
            className="flex min-h-0 flex-1 flex-col"
            onSubmit={handleSubmit}
          >
            <div className="flex-1 space-y-5 overflow-y-auto px-6 py-6">
              <div className="space-y-2">
                <label
                  htmlFor="inventory-product"
                  className="text-sm font-semibold"
                >
                  Produto
                </label>
                <select
                  id="inventory-product"
                  required
                  className="h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                  value={form.productId}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      productId: event.target.value,
                    }))
                  }
                >
                  <option value="">Selecione o produto</option>
                  {items.map((item) => (
                    <option key={item.productId} value={item.productId}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid gap-5 sm:grid-cols-2">
                <div className="space-y-2">
                  <label
                    htmlFor="inventory-quantity"
                    className="text-sm font-semibold"
                  >
                    Quantidade
                  </label>
                  <Input
                    id="inventory-quantity"
                    required
                    type="number"
                    min="0.01"
                    step="0.01"
                    placeholder="0"
                    value={form.quantity}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        quantity: event.target.value,
                      }))
                    }
                  />
                </div>
                {formMode === "adjustment" && (
                  <div className="space-y-2">
                    <label
                      htmlFor="inventory-direction"
                      className="text-sm font-semibold"
                    >
                      Movimento
                    </label>
                    <select
                      id="inventory-direction"
                      className="h-10 w-full rounded-md border border-(--line) bg-(--contrast-light) px-2.5 text-sm"
                      value={form.direction}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          direction: event.target.value as "IN" | "OUT",
                        }))
                      }
                    >
                      <option value="IN">Adicionar</option>
                      <option value="OUT">Retirar</option>
                    </select>
                  </div>
                )}
              </div>
              <div className="space-y-2">
                <label
                  htmlFor="inventory-reason"
                  className="text-sm font-semibold"
                >
                  Motivo{" "}
                  {formMode === "entry" && (
                    <span className="font-normal text-(--ink-soft)">
                      (opcional)
                    </span>
                  )}
                </label>
                <Input
                  id="inventory-reason"
                  required={formMode === "adjustment"}
                  placeholder={
                    formMode === "entry"
                      ? "Ex.: Compra do fornecedor"
                      : "Descreva o motivo"
                  }
                  value={form.reason}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      reason: event.target.value,
                    }))
                  }
                />
              </div>
              {(createEntry.error || createAdjustment.error) && (
                <p className="rounded-md bg-(--coral-wash) px-3 py-2 text-xs text-(--coral-dark)">
                  {(createEntry.error || createAdjustment.error)?.message}
                </p>
              )}
            </div>
            <Sheet.Footer className="bg-(--surface) px-6 py-5">
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
                  disabled={isSaving || !form.productId}
                  className="flex-1 rounded-md bg-(--coral) text-(--contrast-light) hover:bg-(--coral-dark)"
                >
                  {isSaving
                    ? "Salvando..."
                    : formMode === "entry"
                      ? "Registrar entrada"
                      : "Salvar ajuste"}
                </Button>
              </div>
            </Sheet.Footer>
          </form>
        </Sheet.Content>
      </div>
    </Sheet>
  );
}

function StockStat({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "blue" | "coral" | "mint";
}) {
  const styles = {
    blue: "bg-(--blue-wash) text-(--blue-foreground)",
    coral: "bg-(--coral-wash) text-(--coral-dark)",
    mint: "bg-(--mint) text-(--mint-foreground)",
  };
  return (
    <Card className="rounded-2xl border-(--line) bg-(--surface) shadow-none">
      <Card.Content className="flex items-center justify-between p-5">
        <div>
          <p className="text-xs text-(--ink-soft)">{label}</p>
          <p className="mt-2 font-serif text-3xl font-semibold">{value}</p>
        </div>
        <span
          className={`grid size-10 place-items-center rounded-md ${styles[tone]}`}
        >
          <Box className="size-4" />
        </span>
      </Card.Content>
    </Card>
  );
}
function InventoryRow({
  name,
  location,
  quantity,
  minimum,
  state,
  progress,
  isWarning,
}: {
  name: string;
  location: string;
  quantity: string;
  minimum: string;
  state: string;
  progress: string;
  isWarning?: boolean;
}) {
  return (
    <div className="grid gap-4 px-5 py-4 sm:grid-cols-[1fr_180px_110px_150px] sm:items-center">
      <div>
        <p className="text-sm font-semibold">{name}</p>
        <p className="mt-0.5 text-xs text-(--ink-soft)">{location}</p>
      </div>
      <div>
        <div className="mb-1 flex justify-between text-[11px] text-(--ink-soft)">
          <span>Nível atual</span>
          <span>
            {quantity} / mín. {minimum}
          </span>
        </div>
        <div className="h-1.5 overflow-hidden rounded-full bg-(--paper-deep)">
          <div
            className={`h-full rounded-full ${isWarning ? "bg-(--coral)" : "bg-(--mint-progress)"}`}
            style={{ width: progress }}
          />
        </div>
      </div>
      <p className="text-sm font-semibold">{quantity} un.</p>
      <Badge
        variant={isWarning ? "outline" : "secondary"}
        className={
          isWarning
            ? "w-fit border-(--coral-border) text-(--coral-dark)"
            : "w-fit bg-(--mint) text-(--mint-foreground)"
        }
      >
        {state}
      </Badge>
    </div>
  );
}
