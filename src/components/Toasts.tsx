"use client";

import { useEffect } from "react";
import { ShieldAlert, X } from "lucide-react";
import { SEVERITY_COLOR, SEVERITY_LABEL } from "@/lib/severity";

export interface ToastItem {
  id: number;
  eventId: number;
  type: string;
  severity: string;
}

// avisos que salen abajo a la derecha cuando llega un evento alto o critico
// cada uno se quita solo a los 6s o con la X
export function Toasts({ items, onDismiss }: { items: ToastItem[]; onDismiss: (id: number) => void }) {
  return (
    <div
      className="fixed bottom-4 right-4 z-40 flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-2"
      aria-live="polite"
    >
      {items.map((t) => (
        <Toast key={t.id} item={t} onDismiss={onDismiss} />
      ))}
    </div>
  );
}

function Toast({ item, onDismiss }: { item: ToastItem; onDismiss: (id: number) => void }) {
  const color = SEVERITY_COLOR[item.severity] ?? "var(--c-accent)";

  useEffect(() => {
    const timer = setTimeout(() => onDismiss(item.id), 6000);
    return () => clearTimeout(timer);
  }, [item.id, onDismiss]);

  return (
    <div
      // los criticos son alert y interrumpen al lector de pantalla y los altos solo avisan
      role={item.severity === "critical" ? "alert" : "status"}
      className="toast-in flex items-start gap-3 rounded-lg border p-3"
      style={{
        background: `color-mix(in srgb, ${color} 10%, var(--c-surface-raised))`,
        borderColor: `color-mix(in srgb, ${color} 45%, var(--c-border))`,
        boxShadow: "var(--shadow-3)",
      }}
    >
      <span style={{ color }} className="mt-0.5">
        <ShieldAlert size={18} aria-hidden="true" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold" style={{ color }}>
          Evento {SEVERITY_LABEL[item.severity]?.toLowerCase() ?? item.severity}
        </p>
        <p className="truncate font-mono text-xs" style={{ color: "var(--c-text)" }}>
          #{item.eventId} {item.type}
        </p>
      </div>
      <button
        type="button"
        onClick={() => onDismiss(item.id)}
        aria-label="Cerrar aviso"
        className="btn-press -m-1 flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
        style={{ color: "var(--c-text-2)" }}
      >
        <X size={16} />
      </button>
    </div>
  );
}
