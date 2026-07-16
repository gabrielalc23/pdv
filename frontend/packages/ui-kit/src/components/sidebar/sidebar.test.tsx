import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Sidebar } from "./index"

function renderSidebar() {
  return render(
    <Sidebar.Provider defaultOpen>
      <Sidebar>
        <Sidebar.Header>Main Menu</Sidebar.Header>
        <Sidebar.Menu>
          <Sidebar.MenuItem>
            <Sidebar.MenuButton>Dashboard</Sidebar.MenuButton>
          </Sidebar.MenuItem>
        </Sidebar.Menu>
      </Sidebar>
      <Sidebar.Trigger />
    </Sidebar.Provider>,
  )
}

describe("Sidebar", () => {
  it("renders the sidebar content and trigger", () => {
    renderSidebar()

    expect(screen.getByText("Main Menu")).toBeInTheDocument()
    expect(screen.getByText("Dashboard")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: /toggle sidebar/i })).toBeInTheDocument()
  })

  it("toggles open state when the trigger is clicked", async () => {
    const onOpenChange = vi.fn()
    render(
      <Sidebar.Provider defaultOpen onOpenChange={onOpenChange}>
        <Sidebar>
          <Sidebar.Header>Main Menu</Sidebar.Header>
        </Sidebar>
        <Sidebar.Trigger />
      </Sidebar.Provider>,
    )

    await userEvent.click(screen.getByRole("button", { name: /toggle sidebar/i }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
