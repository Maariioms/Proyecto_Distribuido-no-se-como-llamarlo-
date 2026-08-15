"use client";

import { AlertTriangle, ShieldHalf, Radio, Activity } from "lucide-react";

type Severidad = "critical" | "high" | "medium" | "low";

interface EventoSOC {
  id: string;
  fuente: string;
  tipo: string;
  severidad: Severidad;
  estado: "abierto" | "en revisión" | "cerrado";
  timestamp: string;
}

const SEVERIDAD_LABEL: Record<Severidad, string> = {
  critical: "Crítica",
  high: "Alta",
  medium: "Media",
  low: "Baja",
};

const SEVERIDAD_COLOR: Record<Severidad, string> = {
  critical: "var(--sev-critical)",
  high: "var(--sev-high)",
  medium: "var(--sev-medium)",
  low: "var(--sev-low)",
};

const EVENTOS_MOCK: EventoSOC[] = [
  { id: "EVT-1042", fuente: "fw-edge-01", tipo: "Tráfico saliente anómalo", severidad: "critical", estado: "abierto", timestamp: "2026-08-14 21:03" },
  { id: "EVT-1041", fuente: "endpoint-victor-laptop", tipo: "Proceso no firmado ejecutado", severidad: "high", estado: "en revisión", timestamp: "2026-08-14 20:47" },
  { id: "EVT-1040", fuente: "vpn-gateway", tipo: "Login desde IP no reconocida", severidad: "medium", estado: "en revisión", timestamp: "2026-08-14 20:12" },
  { id: "EVT-1039", fuente: "dns-resolver", tipo: "Consulta a dominio DGA sospechoso", severidad: "high", estado: "abierto", timestamp: "2026-08-14 19:58" },
  { id: "EVT-1038", fuente: "endpoint-ana-desktop", tipo: "Escaneo de puertos interno", severidad: "medium", estado: "cerrado", timestamp: "2026-08-14 18:30" },
  { id: "EVT-1037", fuente: "mail-gateway", tipo: "Adjunto con macro sospechosa bloqueado", severidad: "low", estado: "cerrado", timestamp: "2026-08-14 17:15" },
];

const RESUMEN = {
  eventosHoy: 37,
  alertasCriticas: EVENTOS_MOCK.filter((e) => e.severidad === "critical").length,
  agentesActivos: 14,
  agentesTotal: 16,
};

export default function HomePage() {
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
            Estado de la red en las últimas 24 horas.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard
            icon={<Activity size={18} />}
            label="Eventos hoy"
            value={RESUMEN.eventosHoy.toString()}
          />
          <StatCard
            icon={<AlertTriangle size={18} />}
            label="Alertas críticas abiertas"
            value={RESUMEN.alertasCriticas.toString()}
            accentColor="var(--sev-critical)"
          />
          <StatCard
            icon={<Radio size={18} />}
            label="Agentes activos"
            value={`${RESUMEN.agentesActivos} / ${RESUMEN.agentesTotal}`}
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
                    <th className="px-4 py-3 font-medium">Fuente</th>
                    <th className="px-4 py-3 font-medium">Tipo</th>
                    <th className="px-4 py-3 font-medium">Severidad</th>
                    <th className="px-4 py-3 font-medium">Estado</th>
                    <th className="px-4 py-3 font-medium">Hora</th>
                  </tr>
                </thead>
                <tbody>
                  {EVENTOS_MOCK.map((ev, i) => (
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
                      <td className="px-4 py-3 font-mono text-xs" style={{ color: "var(--c-text)" }}>
                        {ev.fuente}
                      </td>
                      <td className="px-4 py-3" style={{ color: "var(--c-text)" }}>
                        {ev.tipo}
                      </td>
                      <td className="px-4 py-3">
                        <SeverityPill severidad={ev.severidad} />
                      </td>
                      <td className="px-4 py-3 capitalize" style={{ color: "var(--c-text-2)" }}>
                        {ev.estado}
                      </td>
                      <td
                        className="px-4 py-3 font-mono text-xs tabular-nums"
                        style={{ color: "var(--c-text-3)" }}
                      >
                        {ev.timestamp}
                      </td>
                    </tr>
                  ))}
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

function SeverityPill({ severidad }: { severidad: Severidad }) {
  const color = SEVERIDAD_COLOR[severidad];
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
      {SEVERIDAD_LABEL[severidad]}
    </span>
  );
}
