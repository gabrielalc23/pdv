import { BreadcrumbEllipsis } from "./breadcrumb-ellipsis"
import { BreadcrumbItem } from "./breadcrumb-item"
import { BreadcrumbLink } from "./breadcrumb-link"
import { BreadcrumbList } from "./breadcrumb-list"
import { BreadcrumbPage } from "./breadcrumb-page"
import { BreadcrumbRoot } from "./breadcrumb-root"
import { BreadcrumbSeparator } from "./breadcrumb-separator"

type BreadcrumbComponentType = typeof BreadcrumbRoot & {
  Ellipsis: typeof BreadcrumbEllipsis
  Item: typeof BreadcrumbItem
  Link: typeof BreadcrumbLink
  List: typeof BreadcrumbList
  Page: typeof BreadcrumbPage
  Separator: typeof BreadcrumbSeparator
}

export const Breadcrumb: BreadcrumbComponentType = Object.assign(BreadcrumbRoot, {
  Ellipsis: BreadcrumbEllipsis,
  Item: BreadcrumbItem,
  Link: BreadcrumbLink,
  List: BreadcrumbList,
  Page: BreadcrumbPage,
  Separator: BreadcrumbSeparator,
}) satisfies BreadcrumbComponentType
