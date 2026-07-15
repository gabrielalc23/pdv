import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: AdminHome,
})

function AdminHome() {
  return <main className="p-2">Admin</main>
}
