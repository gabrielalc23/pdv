import type { ProductTileTone } from "../types/product-tile-tone.type";
import type { CatalogProductResponse } from "./catalog.interface";

export interface ProductTileProps {
  product: CatalogProductResponse;
  tone: ProductTileTone;
  quantity: number;
  onAdd: VoidFunction;
  onRemove: VoidFunction;
  isDisabled: boolean;
}
