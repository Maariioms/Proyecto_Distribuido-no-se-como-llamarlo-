"use client";

import { useEffect, useRef, useState } from "react";
import { X } from "lucide-react";
import { ApiError, getEvent, type Event } from "@/lib/api";
import { SeverityPill } from "./SeverityPill";

// estos tipos los manda accounts por kafka y el resto los inventa el generador de events
const FROM_ACCOUNTS = new Set(["account_created", "login_success", "login_failed"]);

// panel lateral con el detalle de un evento
// pide GET /api/events/{id} aunque la fila ya traiga los datos
// asi se ve que el endpoint de detalle existe y funciona
export function EventPanel({
  event,
  trigger,
  onClose,
}: {
  event: Event;
  trigger: HTMLElement | null;
  onClose: () => void;
}) {
  const [detail, setDetail] = useState<Event | null>(null);
  const [error, setError] = useState("");
  const closeRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLElement>(null);

  // pide el detalle y si el panel se cierra antes de que llegue ignora la respuesta
  useEffect(() => {
    let cancelled = false;
    getEvent(event.id)
      .then((d) => !cancelled && setDetail(d))
      .catch((err) => !cancelled && setError(err instanceof ApiError ? err.message : "No se pudo cargar el detalle."));
    return () => {
      cancelled = true;
    };
  }, [event.id]);

  // foco al boton de cerrar al abrir y de vuelta al disparador al cerrar
  // Escape cierra y Tab se queda dando vueltas dentro del panel
  useEffect(() => {
    closeRef.current?.focus();

    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        onClose();
        return;
      }
      if (e.key !== "Tab" || !panelRef.current) return;

      const focusables = panelRef.current.querySelectorAll<HTMLElement>("button, [href], input, [tabindex]:not([tabindex='-1'])");
      if (focusables.length === 0) return;
      const first = focusables[0];
      const last = focusables[focusables.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
      if (trigger?.isConnected) trigger.focus();
    };
  }, [onClose, trigger]);

  const shown = detail ?? event;
  const created = new Date(shown.created_at);

  return (
    <>
      <div
        className="backdrop-in fixed inset-0 z-20"
        style={{ background: "var(--c-overlay)" }}
        onClick={onClose}
        aria-hidden="true"
      />
      <aside
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="panel-title"
        className="panel-in fixed inset-y-0 right-0 z-30 flex w-full max-w-md flex-col gap-6 overflow-y-auto border-l p-6"
        style={{
          background: "var(--c-surface)",
          borderColor: "var(--c-border-strong)",
          boxShadow: "var(--shadow-3)",
        }}
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-mono tracking-wider" style={{ color: "var(--c-text-2)" }}>
              EVENTO
            </p>
            <h2 id="panel-title" className="font-mono text-lg font-semibold tabular-nums" style={{ color: "var(--c-text)" }}>
              #{shown.id}
            </h2>
          </div>
          <button
            ref={closeRef}
            type="button"
            onClick={onClose}
            aria-label="Cerrar detalle"
            className="btn-press flex h-11 w-11 items-center justify-center rounded-lg border"
            style={{ borderColor: "var(--c-border-strong)", color: "var(--c-text)" }}
          >
            <X size={18} />
          </button>
        </div>

        {error && (
          <p
            role="alert"
            className="rounded-lg px-3 py-2 text-xs"
            style={{
              background: "color-mix(in srgb, var(--c-error) 12%, transparent)",
              border: "1px solid color-mix(in srgb, var(--c-error) 55%, transparent)",
              color: "var(--c-error)",
            }}
          >
            {error} Cierra el panel y vuelve a abrirlo.
          </p>
        )}

        <dl className="flex flex-col gap-5">
          <Field label="Tipo">
            <span className="font-mono text-sm" style={{ color: "var(--c-text)" }}>
              {shown.type}
            </span>
          </Field>
          <Field label="Severidad">
            <SeverityPill severidad={shown.severity} />
          </Field>
          <Field label="Registrado">
            <span className="text-sm" style={{ color: "var(--c-text)" }}>
              {created.toLocaleDateString("es-MX", { dateStyle: "long" })}
            </span>
            <span className="font-mono text-xs tabular-nums" style={{ color: "var(--c-text-2)" }}>
              {created.toLocaleTimeString("es-MX")}
            </span>
          </Field>
          <Field label="Origen">
            <span className="text-sm" style={{ color: "var(--c-text)" }}>
              {FROM_ACCOUNTS.has(shown.type) ? "Lo mandó accounts por Kafka" : "Lo inventó el generador de events"}
            </span>
            <span className="font-mono text-xs" style={{ color: "var(--c-text-2)" }}>
              {FROM_ACCOUNTS.has(shown.type) ? "accounts.activity" : "events"}
            </span>
          </Field>
        </dl>
      </aside>
    </>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <dt className="text-xs font-mono tracking-wider" style={{ color: "var(--c-text-2)" }}>
        {label.toUpperCase()}
      </dt>
      <dd className="flex flex-col items-start gap-1">{children}</dd>
    </div>
  );
}
