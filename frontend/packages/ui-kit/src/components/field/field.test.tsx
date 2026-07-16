import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { Field } from "./index"

describe("Field", () => {
  it("renders label and description", () => {
    render(
      <Field>
        <Field.Label htmlFor="name">Name</Field.Label>
        <Field.Description>Enter your full name</Field.Description>
      </Field>,
    )

    expect(screen.getByText("Name")).toBeInTheDocument()
    expect(screen.getByText("Enter your full name")).toBeInTheDocument()
    expect(screen.getByText("Name").closest("[data-slot='field-label']")).toBeInTheDocument()
  })

  it("renders a single error message", () => {
    render(
      <Field>
        <Field.Label>Email</Field.Label>
        <Field.Error errors={[{ message: "Invalid email" }]} />
      </Field>,
    )

    const alert = screen.getByRole("alert")
    expect(alert).toHaveAttribute("data-slot", "field-error")
    expect(alert).toHaveTextContent("Invalid email")
  })

  it("renders multiple errors as a list", () => {
    render(
      <Field>
        <Field.Label>Password</Field.Label>
        <Field.Error errors={[{ message: "Too short" }, { message: "Missing number" }]} />
      </Field>,
    )

    const alert = screen.getByRole("alert")
    expect(alert).toHaveTextContent("Too short")
    expect(alert).toHaveTextContent("Missing number")
    expect(alert.querySelectorAll("li")).toHaveLength(2)
  })

  it("renders nothing when there are no errors", () => {
    render(
      <Field>
        <Field.Label>Age</Field.Label>
        <Field.Error errors={[undefined, undefined]} />
      </Field>,
    )
    expect(screen.queryByRole("alert")).not.toBeInTheDocument()
  })
})
