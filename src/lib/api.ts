export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

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
