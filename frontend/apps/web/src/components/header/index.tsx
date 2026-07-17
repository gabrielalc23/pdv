import { useEffect, useRef, useState } from "react";
import type { JSX, KeyboardEvent as ReactKeyboardEvent } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  ArrowRight,
  CircleHelp,
  Menu,
  PanelLeftClose,
  PanelLeftOpen,
  Search,
  X,
} from "lucide-react";
import { Button } from "@pdv/ui-kit/components/button";
import { Separator } from "@pdv/ui-kit/components/separator";
import { Sheet } from "@pdv/ui-kit/components/sheet";
import type {
  HeaderProps,
  NavItem,
} from "../../interfaces/app-shell.interface";
import { useSheet } from "../../hooks/use-sheet.hook";
import { operationsNav, primaryNav } from "../sidebar/navigation.config";

const navigationGroups: { label: string; items: NavItem[] }[] = [
  { label: "Principal", items: primaryNav },
  { label: "Operações", items: operationsNav },
];

export function Header({
  current,
  isMobile,
  isCollapsed,
  onOpenMenu,
  onNavigate,
  onToggleSidebar,
}: HeaderProps): JSX.Element {
  const navigate = useNavigate();
  const [isCommandOpen, setIsCommandOpen] = useState<boolean>(false);
  const [query, setQuery] = useState<string>("");
  const [highlightedIndex, setHighlightedIndex] = useState<number>(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const helpSheet = useSheet("ajuda");

  const filteredGroups = navigationGroups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) =>
        item.label.toLowerCase().includes(query.toLowerCase()),
      ),
    }))
    .filter((group) => group.items.length > 0);
  const filteredItems = filteredGroups.flatMap((group) => group.items);

  function openCommandMenu(): void {
    setIsCommandOpen(true);
    setQuery("");
    setHighlightedIndex(0);
  }

  function closeCommandMenu(): void {
    setIsCommandOpen(false);
    setQuery("");
  }

  function goTo(item: NavItem): void {
    closeCommandMenu();
    onNavigate();
    navigate({ to: item.to });
  }

  function handleCommandKeyDown(
    event: ReactKeyboardEvent<HTMLInputElement>,
  ): void {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setHighlightedIndex(
        (index) => (index + 1) % Math.max(filteredItems.length, 1),
      );
    }

    if (event.key === "ArrowUp") {
      event.preventDefault();
      setHighlightedIndex(
        (index) =>
          (index - 1 + Math.max(filteredItems.length, 1)) %
          Math.max(filteredItems.length, 1),
      );
    }

    if (event.key === "Enter" && filteredItems[highlightedIndex]) {
      event.preventDefault();
      goTo(filteredItems[highlightedIndex]);
    }
  }

  useEffect(() => {
    function handleGlobalKeyDown(event: KeyboardEvent): void {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        openCommandMenu();
      }

      if (event.key === "Escape" && isCommandOpen) {
        closeCommandMenu();
      }
    }

    document.addEventListener("keydown", handleGlobalKeyDown);
    return () => document.removeEventListener("keydown", handleGlobalKeyDown);
  }, [isCommandOpen]);

  useEffect(() => {
    if (!isCommandOpen) return;

    inputRef.current?.focus();
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [isCommandOpen]);

  useEffect(() => {
    setHighlightedIndex(0);
  }, [query]);

  return (
    <>
      <header className="sticky top-0 z-20 flex h-16 items-center gap-3  bg-(--paper)/90 px-5 backdrop-blur-md sm:gap-4 sm:px-8 lg:px-10">
        <div className="flex min-w-0 shrink-0 items-center gap-3 sm:gap-4">
          <Button
            type="button"
            variant="outline"
            onClick={isMobile ? onOpenMenu : onToggleSidebar}
            aria-label={
              isMobile
                ? "Abrir menu lateral"
                : isCollapsed
                  ? "Expandir menu lateral"
                  : "Recolher menu lateral"
            }
            className="w-10 h-10 border-(--line) bg-(--surface)/70 text-(--ink-soft) shadow-none hover:bg-(--surface-muted) hover:text-(--ink)"
          >
            {isMobile ? (
              <Menu />
            ) : isCollapsed ? (
              <PanelLeftOpen />
            ) : (
              <PanelLeftClose />
            )}
          </Button>
          <Separator orientation="vertical" className="h-6 bg-(--line)" />
          <div className="hidden min-w-0 items-center gap-2 text-xs sm:flex">
            <span className="text-(--ink-soft)">Workspace</span>
            <span className="text-(--neutral-dark)/20">/</span>
            <span className="truncate font-semibold text-(--ink)">
              {current?.label ?? "Visão geral"}
            </span>
          </div>
        </div>

        <Button
          type="button"
          variant="outline"
          aria-keyshortcuts="Meta+K Control+K"
          onClick={openCommandMenu}
          className="group relative min-w-0 flex-1 justify-start border-(--line) bg-(--surface)/70 px-3 font-normal text-(--ink-soft) shadow-none hover:bg-(--surface-muted) hover:text-(--ink) sm:ml-2 sm:max-w-72 lg:max-w-96"
        >
          <Search className="size-4" aria-hidden="true" />
          <span className="truncate">Navegar para...</span>
          <kbd className="pointer-events-none absolute right-1.5 hidden h-5 items-center gap-1 rounded border border-(--line) bg-(--paper-deep) px-1.5 font-mono text-[10px] font-medium text-(--ink-soft) sm:flex">
            <span className="text-xs">⌘</span>K
          </kbd>
        </Button>

        <Sheet open={helpSheet.isOpen} onOpenChange={helpSheet.setOpen}>
          <Sheet.Trigger
            render={
              <Button
                type="button"
                variant="ghost"
                size="default"
                aria-label="Ajuda"
                className="ml-auto text-(--ink-soft) hover:bg-(--neutral-dark)/5 hover:text-(--ink)"
              />
            }
          >
            <CircleHelp className="size-4" />
          </Sheet.Trigger>
          <Sheet.Content
            side="right"
            className="w-full gap-0 border-(--line) bg-(--surface) p-0 text-(--ink) data-[side=right]:w-full sm:max-w-md"
          >
            <Sheet.Header className="px-6 py-5">
              <p className="mb-1 text-[12px] font-bold tracking-[0.2em] text-(--coral-dark)">
                Central de ajuda
              </p>
              <Sheet.Title className="font-serif text-2xl text-(--ink)">
                Como podemos ajudar?
              </Sheet.Title>
              <Sheet.Description className="text-(--ink-soft)">
                Use os atalhos abaixo para navegar pelo sistema com mais
                agilidade.
              </Sheet.Description>
            </Sheet.Header>
            <div className="flex-1 overflow-y-auto px-6 py-5">
              <div className="space-y-6">
                <section>
                  <h3 className="text-xs font-bold uppercase tracking-[0.16em] text-(--ink-soft)">
                    Atalhos de navegação
                  </h3>
                  <div className="mt-3 divide-y divide-(--line) rounded-md border border-(--line) bg-(--paper)">
                    <HelpShortcut keys="⌘ K" label="Abrir navegação" />
                    <HelpShortcut keys="↑ ↓" label="Navegar entre as páginas" />
                    <HelpShortcut
                      keys="Enter"
                      label="Abrir página selecionada"
                    />
                    <HelpShortcut keys="Esc" label="Fechar janela ou sheet" />
                  </div>
                </section>
                <section className="rounded-md bg-(--coral-wash) p-4">
                  <p className="text-sm font-semibold text-(--coral-dark)">
                    Dica rápida
                  </p>
                  <p className="mt-1 text-xs leading-relaxed text-(--ink-soft)">
                    Clique em “Navegar para...” no header ou pressione{" "}
                    <strong>⌘ K</strong> para encontrar qualquer área do
                    sistema.
                  </p>
                </section>
              </div>
            </div>
            <Sheet.Footer className="bg-(--surface) px-6 py-5">
              <p className="w-full text-xs text-(--ink-soft)">
                Precisa de mais ajuda? Fale com o responsável pelo sistema.
              </p>
            </Sheet.Footer>
          </Sheet.Content>
        </Sheet>
      </header>

      {isCommandOpen && (
        <div
          className="motion-safe:animate-command-overlay-enter fixed inset-0 z-50 bg-(--ink)/45 px-4 pt-[12vh] backdrop-blur-sm"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) closeCommandMenu();
          }}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Navegação"
            className="motion-safe:animate-command-menu-enter mx-auto w-full max-w-xl overflow-hidden rounded-2xl border border-(--line) bg-(--surface) text-(--ink) shadow-2xl"
            onMouseDown={(event) => event.stopPropagation()}
          >
            <div className="flex items-center gap-3  px-4">
              <Search
                className="size-4 shrink-0 text-(--ink-soft)"
                aria-hidden="true"
              />
              <input
                ref={inputRef}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                onKeyDown={handleCommandKeyDown}
                placeholder="Digite uma página para navegar..."
                aria-label="Pesquisar páginas"
                className="h-14 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-(--ink-soft)/70"
              />
              <button
                type="button"
                onClick={closeCommandMenu}
                aria-label="Fechar navegação"
                className="rounded-md p-1.5 text-(--ink-soft) transition-colors hover:bg-(--paper-deep) hover:text-(--ink)"
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="max-h-[min(22rem,55vh)] overflow-y-auto p-2">
              {filteredItems.length === 0 ? (
                <p className="px-3 py-10 text-center text-sm text-(--ink-soft)">
                  Nenhuma página encontrada.
                </p>
              ) : (
                filteredGroups.map((group) => (
                  <div key={group.label} className="mb-3 last:mb-0">
                    <p className="px-3 pb-1.5 pt-2 text-[10px] font-bold uppercase tracking-[0.16em] text-(--ink-soft)/70">
                      {group.label}
                    </p>
                    {group.items.map((item) => {
                      const itemIndex = filteredItems.indexOf(item);
                      const isHighlighted = itemIndex === highlightedIndex;
                      const Icon = item.icon;

                      return (
                        <button
                          key={item.to}
                          type="button"
                          onMouseEnter={() => setHighlightedIndex(itemIndex)}
                          onClick={() => goTo(item)}
                          className={`flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left text-sm transition-colors ${isHighlighted ? "bg-(--coral-wash) text-(--coral-dark)" : "text-(--ink-soft) hover:bg-(--paper-deep) hover:text-(--ink)"}`}
                        >
                          <span className="grid size-7 place-items-center rounded-md bg-(--paper-deep)">
                            <Icon className="size-4" />
                          </span>
                          <span className="flex-1 font-medium">
                            {item.label}
                          </span>
                          {isHighlighted && (
                            <ArrowRight className="size-4" aria-hidden="true" />
                          )}
                        </button>
                      );
                    })}
                  </div>
                ))
              )}
            </div>

            <div className="flex items-center gap-4  px-4 py-3 text-[11px] text-(--ink-soft)">
              <span>
                <kbd className="mr-1 rounded border border-(--line) bg-(--paper-deep) px-1 py-0.5 font-mono">
                  ↑↓
                </kbd>
                navegar
              </span>
              <span>
                <kbd className="mr-1 rounded border border-(--line) bg-(--paper-deep) px-1 py-0.5 font-mono">
                  Enter
                </kbd>
                abrir
              </span>
              <span>
                <kbd className="mr-1 rounded border border-(--line) bg-(--paper-deep) px-1 py-0.5 font-mono">
                  Esc
                </kbd>
                fechar
              </span>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

export function HelpShortcut({
  keys,
  label,
}: {
  keys: string;
  label: string;
}): JSX.Element {
  return (
    <div className="flex items-center justify-between gap-4 px-3 py-3 text-sm">
      <span className="text-(--ink-soft)">{label}</span>
      <kbd className="shrink-0 rounded-md border border-(--line) bg-(--surface) px-2 py-1 font-mono text-[11px] font-medium text-(--ink)">
        {keys}
      </kbd>
    </div>
  );
}
