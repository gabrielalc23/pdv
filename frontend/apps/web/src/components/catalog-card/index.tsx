import { ArrowUpRight, Package } from "lucide-react";
import type { JSX } from "react";
import type { CatalogCardProps } from "../../interfaces/catalog-card.interface";

export function CatalogCard({
  name,
  category,
  price,
  tone,
}: CatalogCardProps): JSX.Element {
  const bg: string = {
    coral: "bg-(--coral-wash)",
    mint: "bg-(--mint)",
    blue: "bg-(--blue-wash)",
    sand: "bg-(--sand)",
  }[tone];
  return (
    <button
      type="button"
      className="group rounded-2xl border border-(--line) bg-(--contrast-light)/40 p-3 text-left transition-all hover:-translate-y-1 hover:bg-(--contrast-light) hover:shadow-(--shadow)"
    >
      <div
        className={`mb-4 grid aspect-[1.7] place-items-center rounded-md ${bg}`}
      >
        <Package className="size-9 text-(--blue-foreground)/40 transition-transform group-hover:scale-110" />
      </div>
      <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-(--coral-dark)">
        {category}
      </p>
      <p className="mt-1 font-serif text-xl font-semibold">{name}</p>
      <div className="mt-4 flex items-center justify-between">
        <strong className="text-sm">{price}</strong>
        <span className="grid size-7 place-items-center rounded-full bg-(--paper-deep) text-(--ink-soft)">
          <ArrowUpRight className="size-3.5" />
        </span>
      </div>
    </button>
  );
}
