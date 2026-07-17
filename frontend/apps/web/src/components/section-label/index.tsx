import type { JSX } from "react";
import type { SectionLabelProps } from "../../interfaces/section-label-props.interface";

export function SectionLabel({
  children,
  action,
}: SectionLabelProps): JSX.Element {
  return (
    <div className="mb-4 flex items-center justify-between">
      <h2 className="text-[1.35rem]">{children}</h2>
      {action}
    </div>
  );
}
