import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { Card } from "./index"

describe("Card", () => {
  it("renders header, title, content and footer", () => {
    render(
      <Card>
        <Card.Header>
          <Card.Title>Profile</Card.Title>
        </Card.Header>
        <Card.Content>Body content</Card.Content>
        <Card.Footer>Footer content</Card.Footer>
      </Card>,
    )

    const root = screen.getByText("Profile").closest("[data-slot='card']")
    expect(root).toBeInTheDocument()
    expect(screen.getByText("Body content")).toBeInTheDocument()
    expect(screen.getByText("Footer content")).toBeInTheDocument()
    expect(screen.getByText("Profile").closest("[data-slot='card-title']")).toBeInTheDocument()
    expect(screen.getByText("Body content").closest("[data-slot='card-content']")).toBeInTheDocument()
    expect(
      screen.getByText("Footer content").closest("[data-slot='card-footer']"),
    ).toBeInTheDocument()
  })

  it("supports the sm size variant", () => {
    render(
      <Card size="sm">
        <Card.Content>Small</Card.Content>
      </Card>,
    )
    const root = screen.getByText("Small").closest("[data-slot='card']")
    expect(root).toHaveAttribute("data-size", "sm")
  })
})
