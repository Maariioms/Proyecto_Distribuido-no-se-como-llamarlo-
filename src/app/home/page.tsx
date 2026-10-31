"use client";

import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Activity, AlertTriangle, LogIn, LogOut, Radio, ShieldHalf } from "lucide-react";
import { getStatus, listEvents, type Event, type ServiceStatus } from "@/lib/api";
import { clearSession, readSession } from "@/lib/session";
import { EventPanel } from "@/components/EventPanel";
import { EventsTable } from "@/components/EventsTable";
import { StatCard } from "@/components/StatCard";
import { Toasts, type ToastItem } from "@/components/Toasts";

// cada cuanto le pregunta al backend si hay algo nuevo
const POLL_MS = 3000;
// cuanto tiempo se queda iluminada una fila nueva
const FRESH_MS = 2500;
// maximo de avisos al mismo tiempo
const MAX_TOASTS = 3;

const subscribeNoop = () => () => {};

export default function HomePage() {
  const router = useRouter();

  const [events, setEvents] = useState<Event[]>([]);
  const [status, setStatus] = useState<ServiceStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [live, setLive] = useState(true);
  // la hora va en el estado pq llamar Date.now en pleno render no es puro
  const [now, setNow] = useState(0);

  const [freshIds, setFreshIds] = useState<Set<number>>(new Set());
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const [selected, setSelected] = useState<{ event: Event; trigger: HTMLElement } | null>(null);

  // el nombre sale del sessionStorage
  // useSyncExternalStore es la forma de leer algo de afuera de React
  // en el servidor devuelve null y en el navegador lee el storage sin chocar con el html que ya viene
  // subscribe no hace nada pq el storage no avisa de cambios en la misma pestana
  const user = useSyncExternalStore(subscribeNoop, readSession, () => null);

  // el id mas alto que ya vimos y con eso se sabe cual evento es nuevo
  // null = todavia no carga nada y por eso la primera carga no avisa de nada
  const lastMaxId = useRef<number | null>(null);
  const toastSeq = useRef(0);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const freshTimers: ReturnType<typeof setTimeout>[] = [];

    async function tick() {
      // con la pestana escondida no tiene caso estar preguntando
      if (!document.hidden) {
        try {
          const [eventsData, statusData] = await Promise.all([listEvents(), getStatus()]);
          if (cancelled) return;

          const prevMax = lastMaxId.current;
          const newest = eventsData.reduce((m, e) => Math.max(m, e.id), 0);
          lastMaxId.current = newest;

          if (prevMax !== null) {
            const arrived = eventsData.filter((e) => e.id > prevMax);
            if (arrived.length > 0) {
              const ids = new Set(arrived.map((e) => e.id));
              setFreshIds(ids);
              freshTimers.push(setTimeout(() => setFreshIds(new Set()), FRESH_MS));

              // solo los altos y criticos merecen aviso
              const alerts = arrived
                .filter((e) => e.severity === "critical" || e.severity === "high")
                .slice(0, MAX_TOASTS)
                .map((e) => ({ id: ++toastSeq.current, eventId: e.id, type: e.type, severity: e.severity }));
              if (alerts.length > 0) setToasts((cur) => [...cur, ...alerts].slice(-MAX_TOASTS));
            }
          }

          setEvents(eventsData);
          setStatus(statusData);
          setNow(Date.now());
          setLive(true);
          setError("");
        } catch {
          if (cancelled) return;
          setLive(false);
          setError("No se pudo conectar con el middleware. Sigo intentando cada 3 segundos.");
        } finally {
          if (!cancelled) setLoading(false);
        }
      }
      if (!cancelled) timer = setTimeout(tick, POLL_MS);
    }

    tick();
    return () => {
      cancelled = true;
      clearTimeout(timer);
      freshTimers.forEach(clearTimeout);
    };
  }, []);

  // estos dos tienen que ser estables
  // si cambian en cada render los timers de los toasts y el foco del panel se reinician con cada poll
  const dismissToast = useCallback((id: number) => {
    setToasts((cur) => cur.filter((t) => t.id !== id));
  }, []);
  const closePanel = useCallback(() => setSelected(null), []);

  function logout() {
    clearSession();
    router.push("/");
  }

  const criticas = events.filter((e) => e.severity === "critical").length;

  return (
    <div className="min-h-screen">
      <header
        className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b px-4 py-3 sm:px-6"
        style={{ background: "var(--c-surface)", borderColor: "var(--c-border)" }}
      >
        <div className="flex items-center gap-2.5">
          <ShieldHalf size={22} style={{ color: "var(--c-accent)" }} />
          <span className="text-sm font-semibold tracking-tight" style={{ color: "var(--c-text)" }}>
            XSIAM-sito
          </span>
        </div>

        <div className="flex items-center gap-3 sm:gap-4">
          <LiveIndicator live={live} />
          {user ? (
            <>
              <span className="hidden font-mono text-xs sm:inline" style={{ color: "var(--c-text-2)" }}>
                {user}
              </span>
              <button
                type="button"
                onClick={logout}
                className="btn-press inline-flex h-10 items-center gap-2 rounded-lg border px-3 text-xs font-medium"
                style={{ borderColor: "var(--c-border-strong)", color: "var(--c-text)" }}
              >
                <LogOut size={14} aria-hidden="true" />
                Salir
              </button>
            </>
          ) : (
            <Link
              href="/"
              className="btn-press inline-flex h-10 items-center gap-2 rounded-lg border px-3 text-xs font-medium"
              style={{ borderColor: "var(--c-border-strong)", color: "var(--c-text)" }}
            >
              <LogIn size={14} aria-hidden="true" />
              Iniciar sesión
            </Link>
          )}
        </div>
      </header>

      <main className="mx-auto flex max-w-5xl flex-col gap-8 px-4 py-8 sm:px-6">
        <div>
          <h1 className="text-xl font-semibold tracking-tight" style={{ color: "var(--c-text)" }}>
            Resumen
          </h1>
          <p className="mt-1 text-sm" style={{ color: "var(--c-text-2)" }}>
            Se actualiza solo cada {POLL_MS / 1000} segundos. Datos en vivo desde el backend, vía middleware.
          </p>
        </div>

        {error && (
          <p
            role="alert"
            className="rounded-lg px-3 py-2 text-sm"
            style={{
              background: "color-mix(in srgb, var(--c-error) 12%, transparent)",
              border: "1px solid color-mix(in srgb, var(--c-error) 55%, transparent)",
              color: "var(--c-error)",
            }}
          >
            {error}
          </p>
        )}

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <StatCard
            icon={<Activity size={18} />}
            label="Eventos registrados"
            value={loading ? null : events.length}
          />
          <StatCard
            icon={<AlertTriangle size={18} />}
            label="Eventos críticos"
            value={loading ? null : criticas}
            accentColor="var(--sev-critical)"
          />
          <StatCard
            icon={<Radio size={18} />}
            label={status ? `Nodo ${status.service_id}` : "Nodo status"}
            value={status ? status.uptime : null}
            hint="tiempo encendido"
          />
        </div>

        <EventsTable
          events={events}
          loading={loading}
          freshIds={freshIds}
          now={now}
          onSelect={(event, trigger) => setSelected({ event, trigger })}
        />
      </main>

      {selected && <EventPanel event={selected.event} trigger={selected.trigger} onClose={closePanel} />}
      <Toasts items={toasts} onDismiss={dismissToast} />
    </div>
  );
}

// puntito verde que respira cuando llegan datos y rojo y quieto cuando no hay conexion
// lleva texto tambien pa no depender solo del color
function LiveIndicator({ live }: { live: boolean }) {
  const color = live ? "var(--c-success)" : "var(--c-error)";
  return (
    <span
      className="inline-flex items-center gap-2 font-mono text-xs tracking-wider"
      style={{ color }}
      role="status"
    >
      <span
        className={`h-2 w-2 rounded-full ${live ? "live-dot" : ""}`}
        style={{ background: color }}
        aria-hidden="true"
      />
      {live ? "EN VIVO" : "SIN CONEXIÓN"}
    </span>
  );
}
