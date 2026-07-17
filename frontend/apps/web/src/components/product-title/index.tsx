import { Minus, Package, Plus } from "lucide-react";
import type { JSX } from "react";
import type { ProductTileProps } from "../../interfaces/product-tile-props.interface";
import { formatCurrency } from "../../utils/format-currency.util";

export function ProductTile({
  product,
  tone,
  quantity,
  onAdd,
  onRemove,
  isDisabled,
}: ProductTileProps): JSX.Element {
  const background: string = {
    coral: "bg-(--coral-wash)",
    mint: "bg-(--mint)",
    blue: "bg-(--blue-wash)",
    sand: "bg-(--sand)",
  }[tone];

  return (
    <div
      className={`group rounded-md border bg-(--contrast-light)/50 p-3 text-left transition-all hover:-translate-y-0.5 hover:bg-(--contrast-light) hover:shadow-md ${quantity > 0 ? "border-(--coral-border) shadow-sm" : "border-(--line)"}`}
    >
      <div
        className={`mb-3 grid aspect-[1.35] place-items-center rounded-md ${background}`}
      >
        <Package className="size-7 text-(--blue-foreground)/45 transition-transform group-hover:scale-110" />
      </div>
      <p className="truncate text-sm font-semibold">{product.name}</p>
      <p className="mt-1 truncate text-[11px] text-(--ink-soft)">
        SKU {product.sku} · {product.quantity} un.
      </p>
      <div className="mt-3 flex items-center justify-between gap-2">
        <p className="text-sm font-bold text-(--coral-dark)">
          {formatCurrency(product.price)}
        </p>
        <div className="flex items-center rounded-md border border-(--line) bg-(--contrast-light)">
          <button
            type="button"
            aria-label={`Remover ${product.name} do carrinho`}
            disabled={isDisabled || quantity === 0}
            onClick={onRemove}
            className="grid size-8 place-items-center text-(--ink-soft) transition-colors hover:bg-(--surface-muted) hover:text-(--ink) disabled:pointer-events-none disabled:opacity-35"
          >
            <Minus className="size-3.5" />
          </button>
          <span
            aria-label={`${quantity} ${product.name} no carrinho`}
            className="grid min-w-7 place-items-center border-x border-(--line) px-1 text-xs font-semibold"
          >
            {quantity}
          </span>
          <button
            type="button"
            aria-label={`Adicionar ${product.name} ao carrinho`}
            disabled={isDisabled}
            onClick={onAdd}
            className="grid size-8 place-items-center text-(--coral-dark) transition-colors hover:bg-(--coral-wash) disabled:pointer-events-none disabled:opacity-35"
          >
            <Plus className="size-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
}
