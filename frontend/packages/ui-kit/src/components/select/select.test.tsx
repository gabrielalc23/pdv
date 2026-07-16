import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Select } from "./index"

describe("Select", () => {
  it("renders a trigger with a placeholder", () => {
    render(
      <Select.Root>
        <Select.Trigger>
          <Select.Value placeholder="Choose a fruit" />
        </Select.Trigger>
        <Select.Content>
          <Select.Item value="apple">Apple</Select.Item>
        </Select.Content>
      </Select.Root>,
    )

    const trigger = screen.getByRole("button", { name: /choose a fruit/i })
    expect(trigger).toBeInTheDocument()
    expect(trigger.closest("[data-slot='select']")).toBeInTheDocument()
  })

  it("opens the list and exposes items", async () => {
    render(
      <Select.Root>
        <Select.Trigger>
          <Select.Value placeholder="Choose a fruit" />
        </Select.Trigger>
        <Select.Content>
          <Select.Item value="apple">Apple</Select.Item>
          <Select.Item value="banana">Banana</Select.Item>
        </Select.Content>
      </Select.Root>,
    )

    await userEvent.click(screen.getByRole("button", { name: /choose a fruit/i }))

    expect(await screen.findByText("Apple")).toBeInTheDocument()
    expect(screen.getByText("Banana")).toBeInTheDocument()
  })

  it("calls onValueChange when an item is selected", async () => {
    const onValueChange = vi.fn()
    render(
      <Select.Root onValueChange={onValueChange}>
        <Select.Trigger>
          <Select.Value placeholder="Choose a fruit" />
        </Select.Trigger>
        <Select.Content>
          <Select.Item value="apple">Apple</Select.Item>
        </Select.Content>
      </Select.Root>,
    )

    await userEvent.click(screen.getByRole("button", { name: /choose a fruit/i }))
    await userEvent.click(await screen.findByText("Apple"))

    expect(onValueChange).toHaveBeenCalledWith("apple", expect.anything())
  })
})
