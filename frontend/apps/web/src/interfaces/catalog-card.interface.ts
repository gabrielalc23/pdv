import type { CatalogCardTone } from "../types/catalog-card.type";

export interface CatalogCardProps {
  name: string;
  category: string;
  price: string;
  tone: CatalogCardTone;
}
