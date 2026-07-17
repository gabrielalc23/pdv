import { SelectContent } from "./select-content";
import { SelectGroup } from "./select-group";
import { SelectItem } from "./select-item";
import { SelectLabel } from "./select-label";
import { SelectRoot } from "./select-root";
import { SelectScrollDownButton } from "./select-scroll-down-button";
import { SelectScrollUpButton } from "./select-scroll-up-button";
import { SelectSeparator } from "./select-separator";
import { SelectTrigger } from "./select-trigger";
import { SelectValue } from "./select-value";

export type SelectComponentType = typeof SelectRoot & {
  Content: typeof SelectContent;
  Group: typeof SelectGroup;
  Item: typeof SelectItem;
  Label: typeof SelectLabel;
  ScrollDownButton: typeof SelectScrollDownButton;
  ScrollUpButton: typeof SelectScrollUpButton;
  Separator: typeof SelectSeparator;
  Trigger: typeof SelectTrigger;
  Value: typeof SelectValue;
};

export const Select: SelectComponentType = Object.assign(SelectRoot, {
  Content: SelectContent,
  Group: SelectGroup,
  Item: SelectItem,
  Label: SelectLabel,
  ScrollDownButton: SelectScrollDownButton,
  ScrollUpButton: SelectScrollUpButton,
  Separator: SelectSeparator,
  Trigger: SelectTrigger,
  Value: SelectValue,
}) satisfies SelectComponentType;
