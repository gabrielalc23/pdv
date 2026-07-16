"use client"

import FieldContent from "./field-content"
import FieldDescription from "./field-description"
import FieldError from "./field-error"
import FieldGroup from "./field-group"
import FieldLabel from "./field-label"
import FieldLegend from "./field-legend"
import FieldRoot from "./field-root"
import FieldSeparator from "./field-separator"
import FieldSet from "./field-set"
import FieldTitle from "./field-title"

export type FieldComponentType = typeof FieldRoot & {
  Content: typeof FieldContent
  Description: typeof FieldDescription
  Error: typeof FieldError
  Group: typeof FieldGroup
  Label: typeof FieldLabel
  Legend: typeof FieldLegend
  Separator: typeof FieldSeparator
  Set: typeof FieldSet
  Title: typeof FieldTitle
}

export const Field: FieldComponentType = Object.assign(FieldRoot, {
  Content: FieldContent,
  Description: FieldDescription,
  Error: FieldError,
  Group: FieldGroup,
  Label: FieldLabel,
  Legend: FieldLegend,
  Separator: FieldSeparator,
  Set: FieldSet,
  Title: FieldTitle,
}) satisfies FieldComponentType
