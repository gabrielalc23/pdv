import { useLocation, useRouter, useSearch } from "@tanstack/react-router";
import type { NavigateOptions, ParsedLocation } from "@tanstack/react-router";

export type SheetId =
  | "ajuda"
  | "nova-venda"
  | "carrinho"
  | "atalhos"
  | "pedidos-abertos"
  | "novo-produto"
  | "entrada-estoque"
  | "ajuste-estoque"
  | "filtros-produtos"
  | "filtros-vendas"
  | "filtros-catalogo";

interface SheetState {
  isOpen: boolean;
  setOpen: (isOpen: boolean) => void;
}

interface SheetLocation {
  sheet?: string | undefined;
  sidebar_mobile?: boolean | undefined;
}

interface SheetSearch {
  sheet?: string | undefined;
  sidebar_mobile?: boolean | undefined;
}

export function useSheet(sheetId: SheetId): SheetState {
  const router = useRouter();
  const location: ParsedLocation<SheetLocation> = useLocation();
  const search: SheetSearch = useSearch({ strict: false });

  function setOpen(isOpen: boolean): void {
    const nextUrl: URL = new URL(window.location.href);

    if (isOpen) {
      nextUrl.searchParams.set("sheet", sheetId);
    } else {
      nextUrl.searchParams.delete("sheet");
    }

    const nextPath = `${location.pathname}${nextUrl.search}${nextUrl.hash}`;
    const updateHistory: (
      path: string,
      state?: any,
      navigateOpts?: NavigateOptions,
    ) => void = isOpen ? router.history.push : router.history.replace;
    updateHistory.call(router.history, nextPath);
  }

  return { isOpen: search.sheet === sheetId, setOpen };
}
