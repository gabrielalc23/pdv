import { SheetClose } from "./sheet-close";
import { SheetContent } from "./sheet-content";
import { SheetDescription } from "./sheet-description";
import { SheetFooter } from "./sheet-footer";
import { SheetHeader } from "./sheet-header";
import { SheetOverlay } from "./sheet-overlay";
import { SheetPortal } from "./sheet-portal";
import { SheetRoot } from "./sheet-root";
import { SheetTitle } from "./sheet-title";
import { SheetTrigger } from "./sheet-trigger";

export type SheetComponentType = typeof SheetRoot & {
  Close: typeof SheetClose;
  Content: typeof SheetContent;
  Description: typeof SheetDescription;
  Footer: typeof SheetFooter;
  Header: typeof SheetHeader;
  Overlay: typeof SheetOverlay;
  Portal: typeof SheetPortal;
  Title: typeof SheetTitle;
  Trigger: typeof SheetTrigger;
};

export const Sheet: SheetComponentType = Object.assign(SheetRoot, {
  Close: SheetClose,
  Content: SheetContent,
  Description: SheetDescription,
  Footer: SheetFooter,
  Header: SheetHeader,
  Overlay: SheetOverlay,
  Portal: SheetPortal,
  Title: SheetTitle,
  Trigger: SheetTrigger,
}) satisfies SheetComponentType;
