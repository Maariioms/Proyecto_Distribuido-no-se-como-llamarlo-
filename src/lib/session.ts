// sesion de mentiritas pa que el header sepa quien hizo login
// solo guarda el nombre en sessionStorage y no hay token ni nada seguro
// todo en try/catch pq el storage puede estar bloqueado o venir vacio

const KEY = "xsiam.username";

export function saveSession(username: string) {
  try {
    sessionStorage.setItem(KEY, username);
  } catch {
    // sin storage la app funciona igual, solo no recuerda el nombre
  }
}

export function readSession(): string | null {
  try {
    return sessionStorage.getItem(KEY);
  } catch {
    return null;
  }
}

export function clearSession() {
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    // nada que limpiar
  }
}
