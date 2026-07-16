import SidebarContent from "./sidebar-content"
import SidebarFooter from "./sidebar-footer"
import SidebarGroupAction from "./sidebar-group-action"
import SidebarGroupContent from "./sidebar-group-content"
import SidebarGroupLabel from "./sidebar-group-label"
import SidebarGroup from "./sidebar-group"
import SidebarHeader from "./sidebar-header"
import SidebarInset from "./sidebar-inset"
import SidebarInput from "./sidebar-input"
import SidebarMenuAction from "./sidebar-menu-action"
import SidebarMenuBadge from "./sidebar-menu-badge"
import SidebarMenuButton from "./sidebar-menu-button"
import SidebarMenuItem from "./sidebar-menu-item"
import SidebarMenuSkeleton from "./sidebar-menu-skeleton"
import SidebarMenuSubButton from "./sidebar-menu-sub-button"
import SidebarMenuSubItem from "./sidebar-menu-sub-item"
import SidebarMenuSub from "./sidebar-menu-sub"
import SidebarMenu from "./sidebar-menu"
import SidebarProvider from "./sidebar-provider"
import SidebarRail from "./sidebar-rail"
import SidebarRoot from "./sidebar-root"
import SidebarSeparator from "./sidebar-separator"
import SidebarTrigger from "./sidebar-trigger"

export type SidebarComponentType = typeof SidebarRoot & {
  Content: typeof SidebarContent
  Footer: typeof SidebarFooter
  Group: typeof SidebarGroup
  GroupAction: typeof SidebarGroupAction
  GroupContent: typeof SidebarGroupContent
  GroupLabel: typeof SidebarGroupLabel
  Header: typeof SidebarHeader
  Inset: typeof SidebarInset
  Input: typeof SidebarInput
  Menu: typeof SidebarMenu
  MenuAction: typeof SidebarMenuAction
  MenuBadge: typeof SidebarMenuBadge
  MenuButton: typeof SidebarMenuButton
  MenuItem: typeof SidebarMenuItem
  MenuSkeleton: typeof SidebarMenuSkeleton
  MenuSub: typeof SidebarMenuSub
  MenuSubButton: typeof SidebarMenuSubButton
  MenuSubItem: typeof SidebarMenuSubItem
  Provider: typeof SidebarProvider
  Rail: typeof SidebarRail
  Separator: typeof SidebarSeparator
  Trigger: typeof SidebarTrigger
}

export const Sidebar: SidebarComponentType = Object.assign(SidebarRoot, {
  Content: SidebarContent,
  Footer: SidebarFooter,
  Group: SidebarGroup,
  GroupAction: SidebarGroupAction,
  GroupContent: SidebarGroupContent,
  GroupLabel: SidebarGroupLabel,
  Header: SidebarHeader,
  Inset: SidebarInset,
  Input: SidebarInput,
  Menu: SidebarMenu,
  MenuAction: SidebarMenuAction,
  MenuBadge: SidebarMenuBadge,
  MenuButton: SidebarMenuButton,
  MenuItem: SidebarMenuItem,
  MenuSkeleton: SidebarMenuSkeleton,
  MenuSub: SidebarMenuSub,
  MenuSubButton: SidebarMenuSubButton,
  MenuSubItem: SidebarMenuSubItem,
  Provider: SidebarProvider,
  Rail: SidebarRail,
  Separator: SidebarSeparator,
  Trigger: SidebarTrigger,
}) satisfies SidebarComponentType
