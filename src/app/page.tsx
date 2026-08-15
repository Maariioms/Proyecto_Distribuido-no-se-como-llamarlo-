"use client";

import { useEffect, useState } from "react";
import { ChevronDown, Loader2, LogIn, ShieldHalf } from "lucide-react";

// 12 cometas a 30° entre sí, con duración/retraso variados para que no se
// vean sincronizadas.
const COMETAS_BASE = Array.from({ length: 12 }, (_, i) => ({
  angle: i * 30,
  delay: (i % 6) * 0.5,
  duration: 2.6 + (i % 4) * 0.4,
}));

interface Cometa {
  angle: number;
  delay: number;
  duration: number;
}

const ANALISTAS_PLACEHOLDER = ["analista.demo"];

export default function Home() {
  const [nombre, setNombre] = useState(ANALISTAS_PLACEHOLDER[0]);
  const [pin, setPin] = useState("");
  const [error, setError] = useState("");
  const [cargando, setCargando] = useState(false);

  const [cometas, setCometas] = useState<Cometa[]>([]);

  useEffect(() => {
    setCometas(
      COMETAS_BASE.map((c) => ({
        ...c,
        angle: c.angle + (Math.random() * 24 - 12),
      }))
    );
  }, []);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setCargando(true);
    // Placeholder: sin autenticación real todavía.
    setTimeout(() => {
      setCargando(false);
      setError("Login no implementado — esto es solo el placeholder visual.");
    }, 500);
  }

  return (
    <div className="relative min-h-screen flex items-center justify-center px-4">
      <div className="comet-field" aria-hidden="true">
        {cometas.map((c, i) => (
          <div
            key={i}
            className="comet-arm"
            style={{ transform: `rotate(${c.angle}deg)` }}
          >
            <span
              className="comet"
              style={{ animationDelay: `${c.delay}s`, animationDuration: `${c.duration}s` }}
            />
          </div>
        ))}
      </div>

      <div className="relative w-full max-w-sm">
        <div className="flex flex-col items-center gap-3 mb-8">
          <ShieldHalf
            className="glow-pulse"
            size={72}
            strokeWidth={1.5}
            style={{ color: "var(--c-accent)" }}
          />
          <p className="text-xs tracking-widest" style={{ color: "var(--c-text-3)" }}>
            Iniciar sesión
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-4 p-6 rounded-lg border"
          style={{ background: "var(--c-surface)", borderColor: "var(--c-border)" }}
        >
          <div className="flex flex-col gap-1.5">
            <label
              className="text-xs font-mono font-semibold tracking-wider"
              style={{ color: "var(--c-text-2)" }}
            >
              ANALISTA
            </label>
            <div className="relative">
              <select
                value={nombre}
                onChange={(e) => setNombre(e.target.value)}
                className="appearance-none w-full rounded-lg pl-3 pr-9 py-2.5 text-sm font-mono outline-none border transition cursor-pointer"
                style={{
                  background: "var(--c-surface-raised)",
                  borderColor: "var(--c-border)",
                  color: "var(--c-text)",
                }}
              >
                {ANALISTAS_PLACEHOLDER.map((a) => (
                  <option key={a} value={a}>
                    {a}
                  </option>
                ))}
              </select>
              <ChevronDown
                className="w-4 h-4 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none"
                style={{ color: "var(--c-text-3)" }}
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              className="text-xs font-mono font-semibold tracking-wider"
              style={{ color: "var(--c-text-2)" }}
            >
              PIN
            </label>
            <input
              type="password"
              inputMode="numeric"
              autoComplete="current-password"
              maxLength={6}
              value={pin}
              onChange={(e) => setPin(e.target.value.replace(/\D/g, ""))}
              placeholder="••••••"
              className="w-full rounded-lg px-3 py-2.5 text-sm font-mono outline-none border tracking-[0.4em] transition"
              style={{
                background: "var(--c-surface-raised)",
                borderColor: "var(--c-border)",
                color: "var(--c-text)",
              }}
            />
          </div>

          {error && (
            <div
              className="px-3 py-2 rounded-lg text-xs"
              style={{
                background: "color-mix(in srgb, var(--c-error) 12%, transparent)",
                border: "1.5px solid color-mix(in srgb, var(--c-error) 55%, transparent)",
                color: "var(--c-error)",
              }}
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={cargando || pin.length === 0}
            className="btn-press flex items-center justify-center gap-2 px-4 py-2.5 mt-1 rounded-lg text-sm font-bold transition disabled:opacity-50"
            style={{ background: "var(--c-accent)", color: "var(--c-on-accent)" }}
          >
            {cargando ? <Loader2 className="w-4 h-4 animate-spin" /> : <LogIn className="w-4 h-4" />}
            Entrar
          </button>
        </form>
      </div>
    </div>
  );
}
