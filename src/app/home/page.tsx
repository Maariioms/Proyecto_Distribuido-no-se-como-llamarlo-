"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, ShieldHalf, Radio, Activity } from "lucide-react";
import { listEvents, getStatus, type Event, type ServiceStatus } from "@/lib/api";

const SEVERIDAD_LABEL: Record<string, string> = {
  critical: "Crítica",
  high: "Alta",
  medium: "Media",
  low: "Baja",
};

const SEVERIDAD_COLOR: Record<string, string> = {
  critical: "var(--sev-critical)",
  high: "var(--sev-high)",
  medium: "var(--sev-medium)",
  low: "var(--sev-low)",
};

export default function HomePage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [status, setStatus] = useState<ServiceStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const [eventsData, statusData] = await Promise.all([listEvents(), getStatus()]);
        setEvents(eventsData);
        setStatus(statusData);
      } catch {
        setError("No se pudo conectar con el middleware.");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const criticas = events.filter((e) => e.severity === "critical").length;

  return (
    <div className="min-h-screen">
      <header
        className="sticky top-0 z-10 flex items-center justify-between px-6 py-4 border-b"
        style={{ background: "var(--c-surface)", borderColor: "var(--c-border)" }}
      >
        <div className="flex items-center gap-2.5">
          <ShieldHalf size={22} style={{ color: "var(--c-accent)" }} />
          <span className="text-sm font-semibold tracking-tight" style={{ color: "var(--c-text)" }}>
            XSIAM-sito
          </span>
        </div>
        <div className="flex items-center gap-3 text-xs" style={{ color: "var(--c-text-2)" }}>
          <span className="font-mono">analista.demo</span>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-8 flex flex-col gap-8">
        <div>
          <h1 className="text-xl font-semibold tracking-tight" style={{ color: "var(--c-text)" }}>
            Resumen
          </h1>
          <p className="mt-1 text-sm" style={{ color: "var(--c-text-2)" }}>
            Datos en vivo desde el backend, vía middleware.
          </p>
        </div>

        {error && (
          <p className="text-sm" style={{ color: "var(--sev-critical)" }}>
            {error}
          </p>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard
            icon={<Activity size={18} />}
            label="Eventos registrados"
            value={loading ? "…" : events.length.toString()}
          />
          <StatCard
            icon={<AlertTriangle size={18} />}
            label="Eventos críticos"
            value={loading ? "…" : criticas.toString()}
            accentColor="var(--sev-critical)"
          />
          <StatCard
            icon={<Radio size={18} />}
            label={status ? `Nodo ${status.service_id}` : "Nodo status"}
            value={status ? status.uptime : "…"}
          />
        </div>

        <div>
          <h2 className="text-sm font-semibold tracking-wide mb-3" style={{ color: "var(--c-text)" }}>
            Eventos recientes
          </h2>

          <div
            className="rounded-lg border overflow-hidden"
            style={{ borderColor: "var(--c-border)", background: "var(--c-surface)" }}
          >
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr
                    className="text-left text-xs font-mono tracking-wider"
                    style={{ color: "var(--c-text-3)", borderBottom: "1px solid var(--c-border)" }}
                  >
                    <th className="px-4 py-3 font-medium">ID</th>
                    <th className="px-4 py-3 font-medium">Tipo</th>
                    <th className="px-4 py-3 font-medium">Severidad</th>
                    <th className="px-4 py-3 font-medium">Hora</th>
                  </tr>
                </thead>
                <tbody>
                  {events.map((ev, i) => (
                    <tr
                      key={ev.id}
                      className="card-in"
                      style={{
                        borderTop: i === 0 ? "none" : "1px solid var(--c-border)",
                        // @ts-expect-error --i es un custom property, no una propiedad CSS estándar
                        "--i": i,
                      }}
                    >
                      <td className="px-4 py-3 font-mono text-xs" style={{ color: "var(--c-text-2)" }}>
                        {ev.id}
                      </td>
                      <td className="px-4 py-3" style={{ color: "var(--c-text)" }}>
                        {ev.type}
                      </td>
                      <td className="px-4 py-3">
                        <SeverityPill severidad={ev.severity} />
                      </td>
                      <td
                        className="px-4 py-3 font-mono text-xs tabular-nums"
                        style={{ color: "var(--c-text-3)" }}
                      >
                        {new Date(ev.created_at).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                  {!loading && events.length === 0 && (
                    <tr>
                      <td colSpan={4} className="px-4 py-6 text-center text-xs" style={{ color: "var(--c-text-3)" }}>
                        Sin eventos todavía.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
  accentColor,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  accentColor?: string;
}) {
  return (
    <div
      className="flex items-center gap-3 rounded-lg border p-4"
      style={{ borderColor: "var(--c-border)", background: "var(--c-surface)" }}
    >
      <div
        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md"
        style={{
          background: `color-mix(in srgb, ${accentColor ?? "var(--c-accent)"} 16%, var(--c-surface))`,
          color: accentColor ?? "var(--c-accent)",
        }}
      >
        {icon}
      </div>
      <div>
        <p className="text-xs" style={{ color: "var(--c-text-3)" }}>
          {label}
        </p>
        <p
          className="text-lg font-semibold tabular-nums"
          style={{ color: "var(--c-text)" }}
        >
          {value}
        </p>
      </div>
    </div>
  );
}

function SeverityPill({ severidad }: { severidad: string }) {
  const color = SEVERIDAD_COLOR[severidad] ?? "var(--c-text-3)";
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
      style={{
        background: `color-mix(in srgb, ${color} 14%, var(--c-surface))`,
        border: `1px solid color-mix(in srgb, ${color} 40%, var(--c-border))`,
        color,
      }}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: color }}
        aria-hidden="true"
      />
      {SEVERIDAD_LABEL[severidad] ?? severidad}
    </span>
  );
}
