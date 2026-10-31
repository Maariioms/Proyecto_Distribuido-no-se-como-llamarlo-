"use client";

import { useState } from "react";
import { ChevronLeft, ChevronRight, Inbox, Search, SearchX } from "lucide-react";
import type { Event } from "@/lib/api";
import { SEVERITIES, SEVERITY_LABEL } from "@/lib/severity";
import { SeverityPill } from "./SeverityPill";

const PAGE_SIZE = 10;

type Filter = "all" | (typeof SEVERITIES)[number];

// tabla con filtro por severidad, busqueda por tipo y paginacion
// todo se hace aqui en el navegador con la lista completa que manda events
// freshIds son los eventos que acaban de llegar y por eso se iluminan un rato
export function EventsTable({
  events,
  loading,
  freshIds,
  now,
  onSelect,
}: {
  events: Event[];
  loading: boolean;
  freshIds: Set<number>;
  now: number;
  onSelect: (event: Event, trigger: HTMLElement) => void;
}) {
  const [filter, setFilter] = useState<Filter>("all");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(0);

  const q = query.trim().toLowerCase();
  const filtered = events.filter(
    (e) => (filter === "all" || e.severity === filter) && (q === "" || e.type.toLowerCase().includes(q))
  );

  const pages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  // si la lista se encoge y la pagina ya no existe se queda en la ultima
  const safePage = Math.min(page, pages - 1);
  const visible = filtered.slice(safePage * PAGE_SIZE, (safePage + 1) * PAGE_SIZE);

  const counts: Record<string, number> = {};
  for (const e of events) counts[e.severity] = (counts[e.severity] ?? 0) + 1;

  function changeFilter(f: Filter) {
    setFilter(f);
    setPage(0);
  }

  function clearFilters() {
    setFilter("all");
    setQuery("");
    setPage(0);
  }

  const from = filtered.length === 0 ? 0 : safePage * PAGE_SIZE + 1;
  const to = Math.min((safePage + 1) * PAGE_SIZE, filtered.length);

  return (
    <section aria-labelledby="events-title" className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <h2 id="events-title" className="text-sm font-semibold" style={{ color: "var(--c-text)" }}>
          Eventos recientes
        </h2>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
          <div role="group" aria-label="Filtrar por severidad" className="flex flex-wrap gap-2">
            <FilterChip active={filter === "all"} onClick={() => changeFilter("all")}>
              Todas <Count n={events.length} />
            </FilterChip>
            {SEVERITIES.map((s) => (
              <FilterChip key={s} active={filter === s} onClick={() => changeFilter(s)}>
                {SEVERITY_LABEL[s]} <Count n={counts[s] ?? 0} />
              </FilterChip>
            ))}
          </div>

          <div className="flex flex-col gap-1">
            <label htmlFor="search-type" className="text-xs font-mono tracking-wider" style={{ color: "var(--c-text-2)" }}>
              BUSCAR TIPO
            </label>
            <div className="relative">
              <Search
                size={14}
                className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: "var(--c-text-3)" }}
                aria-hidden="true"
              />
              <input
                id="search-type"
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setPage(0);
                }}
                placeholder="login_failed"
                autoComplete="off"
                className="h-10 w-full rounded-lg border pl-9 pr-3 text-sm font-mono outline-none transition sm:w-52"
                style={{
                  background: "var(--c-surface-raised)",
                  borderColor: "var(--c-border)",
                  color: "var(--c-text)",
                }}
              />
            </div>
          </div>
        </div>
      </div>

      <div
        className="overflow-hidden rounded-lg border"
        style={{ borderColor: "var(--c-border)", background: "var(--c-surface)" }}
      >
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr
                className="text-left text-xs font-mono tracking-wider"
                style={{ color: "var(--c-text-2)", borderBottom: "1px solid var(--c-border)" }}
              >
                <th scope="col" className="px-4 py-3 font-medium">ID</th>
                <th scope="col" className="px-4 py-3 font-medium">TIPO</th>
                <th scope="col" className="px-4 py-3 font-medium">SEVERIDAD</th>
                <th scope="col" className="px-4 py-3 font-medium text-right">HACE</th>
              </tr>
            </thead>
            <tbody>
              {loading &&
                Array.from({ length: 6 }, (_, i) => (
                  <tr key={i} style={{ borderTop: i === 0 ? "none" : "1px solid var(--c-border)" }}>
                    <td colSpan={4} className="px-4 py-3">
                      <div className="skeleton h-5 w-full rounded" aria-hidden="true" />
                    </td>
                  </tr>
                ))}

              {!loading &&
                visible.map((ev, i) => (
                  <tr
                    key={ev.id}
                    onClick={(e) => onSelect(ev, e.currentTarget.querySelector("button") ?? e.currentTarget)}
                    className={`row-click ${freshIds.has(ev.id) ? "row-new" : "card-in"}`}
                    style={
                      {
                        borderTop: i === 0 ? "none" : "1px solid var(--c-border)",
                        "--i": i,
                      } as React.CSSProperties
                    }
                  >
                    <td className="px-4 py-3 font-mono text-xs tabular-nums" style={{ color: "var(--c-text-2)" }}>
                      {ev.id}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        onClick={(e) => {
                          // el click del tr ya abre el panel y esto evita que se abra dos veces
                          e.stopPropagation();
                          onSelect(ev, e.currentTarget);
                        }}
                        className="min-h-10 text-left font-mono text-xs"
                        style={{ color: "var(--c-text)" }}
                      >
                        {ev.type}
                      </button>
                    </td>
                    <td className="px-4 py-3">
                      <SeverityPill severidad={ev.severity} />
                    </td>
                    <td
                      className="px-4 py-3 text-right font-mono text-xs tabular-nums"
                      style={{ color: "var(--c-text-2)" }}
                      title={new Date(ev.created_at).toLocaleString("es-MX")}
                    >
                      {timeAgo(ev.created_at, now)}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>

        {!loading && events.length === 0 && (
          <EmptyState
            icon={<Inbox size={22} />}
            title="Sin eventos todavía"
            text="Cuando alguien haga login o el generador de events cree uno, aparece aquí solo."
          />
        )}

        {!loading && events.length > 0 && filtered.length === 0 && (
          <EmptyState
            icon={<SearchX size={22} />}
            title="Ningún evento coincide"
            text="Prueba con otra severidad o quita la búsqueda."
            action={
              <button
                type="button"
                onClick={clearFilters}
                className="btn-press mt-1 h-10 rounded-lg border px-4 text-sm font-medium"
                style={{ borderColor: "var(--c-border-strong)", color: "var(--c-text)" }}
              >
                Limpiar filtros
              </button>
            }
          />
        )}
      </div>

      {!loading && filtered.length > 0 && (
        <nav aria-label="Paginación" className="flex items-center justify-between gap-3">
          <p className="text-xs tabular-nums" style={{ color: "var(--c-text-2)" }}>
            <span className="font-mono">{from}–{to}</span> de <span className="font-mono">{filtered.length}</span>
          </p>
          <div className="flex items-center gap-2">
            <PageButton
              label="Anterior"
              disabled={safePage === 0}
              onClick={() => setPage(safePage - 1)}
              icon={<ChevronLeft size={16} />}
            />
            <span className="px-1 text-xs font-mono tabular-nums" style={{ color: "var(--c-text-2)" }}>
              {safePage + 1}/{pages}
            </span>
            <PageButton
              label="Siguiente"
              disabled={safePage >= pages - 1}
              onClick={() => setPage(safePage + 1)}
              icon={<ChevronRight size={16} />}
              iconAfter
            />
          </div>
        </nav>
      )}
    </section>
  );
}

// "hace 5 s" "hace 3 min" "hace 2 h" y despues de un dia ya la fecha
// now viene de afuera pq llamar Date.now en pleno render no es puro
function timeAgo(iso: string, now: number): string {
  const diff = Math.max(0, Math.floor((now - new Date(iso).getTime()) / 1000));
  if (diff < 60) return `${diff} s`;
  if (diff < 3600) return `${Math.floor(diff / 60)} min`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} h`;
  return new Date(iso).toLocaleDateString("es-MX", { day: "2-digit", month: "short" });
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className="btn-press inline-flex h-10 items-center gap-1.5 rounded-lg border px-3 text-xs font-medium transition"
      style={{
        // el activo se nota por fondo y por borde, no solo por color
        background: active ? "var(--c-accent-bg)" : "transparent",
        borderColor: active ? "var(--c-border-accent)" : "var(--c-border)",
        color: active ? "var(--c-text)" : "var(--c-text-2)",
        fontWeight: active ? 600 : 500,
      }}
    >
      {children}
    </button>
  );
}

function Count({ n }: { n: number }) {
  return <span className="font-mono tabular-nums opacity-80">{n}</span>;
}

function PageButton({
  label,
  disabled,
  onClick,
  icon,
  iconAfter,
}: {
  label: string;
  disabled: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  iconAfter?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="btn-press inline-flex h-10 items-center gap-1 rounded-lg border px-3 text-xs font-medium disabled:opacity-40"
      style={{ borderColor: "var(--c-border-strong)", color: "var(--c-text)" }}
    >
      {!iconAfter && icon}
      {label}
      {iconAfter && icon}
    </button>
  );
}

function EmptyState({
  icon,
  title,
  text,
  action,
}: {
  icon: React.ReactNode;
  title: string;
  text: string;
  action?: React.ReactNode;
}) {
  return (
    <div
      className="m-4 flex flex-col items-center gap-2 rounded-lg border border-dashed p-10 text-center"
      style={{ borderColor: "var(--c-border-strong)" }}
    >
      <span style={{ color: "var(--c-accent-2)" }}>{icon}</span>
      <p className="text-sm font-semibold" style={{ color: "var(--c-text)" }}>
        {title}
      </p>
      <p className="max-w-sm text-xs" style={{ color: "var(--c-text-2)" }}>
        {text}
      </p>
      {action}
    </div>
  );
}
