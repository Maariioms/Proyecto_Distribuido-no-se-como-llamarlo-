export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:9000";

export interface Account {
  id: number;
  username: string;
  created_at: string;
}

export class ApiError extends Error {}

export async function login(username: string, password: string): Promise<Account> {
  const res = await fetch(`${API_URL}/api/accounts/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });

  if (!res.ok) {
    const message = await res.text();
    throw new ApiError(message || "No se pudo iniciar sesión.");
  }

  return res.json();
}

export interface Event {
  id: number;
  type: string;
  severity: "critical" | "high" | "medium" | "low" | string;
  created_at: string;
}

export async function listEvents(): Promise<Event[]> {
  const res = await fetch(`${API_URL}/api/events`, { cache: "no-store" });
  if (!res.ok) throw new ApiError("No se pudieron cargar los eventos.");
  return res.json();
}

export interface ServiceStatus {
  status: string;
  service_id: string;
  uptime: string;
  checks_done: number;
}

export async function getStatus(): Promise<ServiceStatus> {
  const res = await fetch(`${API_URL}/health`, { cache: "no-store" });
  if (!res.ok) throw new ApiError("No se pudo consultar el status.");
  return res.json();
}
