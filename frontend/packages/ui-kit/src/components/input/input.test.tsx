import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Input } from "../input"

describe("Input", () => {
  it("renders with a placeholder", () => {
    render(<Input placeholder="Type here" />)
    const input = screen.getByPlaceholderText("Type here")
    expect(input).toBeInTheDocument()
    expect(input).toHaveAttribute("data-slot", "input")
  })

  it("updates value on typing", async () => {
    render(<Input placeholder="Name" />)
    const input = screen.getByPlaceholderText("Name") as HTMLInputElement
    await userEvent.type(input, "hello")
    expect(input.value).toBe("hello")
  })

  it("respects the type prop", () => {
    render(<Input type="email" placeholder="Email" />)
    expect(screen.getByPlaceholderText("Email")).toHaveAttribute("type", "email")
  })

  it("is disabled when disabled prop is set", async () => {
    const onChange = vi.fn()
    render(<Input disabled placeholder="Locked" onChange={onChange} />)
    const input = screen.getByPlaceholderText("Locked")
    expect(input).toBeDisabled()
    await userEvent.type(input, "x")
    expect(onChange).not.toHaveBeenCalled()
  })
})
