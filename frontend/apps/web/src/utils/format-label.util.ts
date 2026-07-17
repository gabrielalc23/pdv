export function statusLabel(status: string): string {
  return status === "COMPLETED"
    ? "Concluída"
    : status === "CANCELLED"
      ? "Cancelada"
      : "Em aberto";
}
