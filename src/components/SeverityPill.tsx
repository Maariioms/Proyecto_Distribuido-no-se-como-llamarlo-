import { SEVERITY_COLOR, SEVERITY_LABEL, SEVERITY_SHAPE } from "@/lib/severity";

// la severidad no se dice solo con color
// cada una tiene su forma de punto y su texto pa que se lea tambien sin ver colores
export function SeverityDot({ severidad }: { severidad: string }) {
  const color = SEVERITY_COLOR[severidad] ?? "var(--c-text-3)";
  return (
    <span
      className={`h-2 w-2 shrink-0 ${SEVERITY_SHAPE[severidad] ?? "rounded-full"}`}
      style={{ background: color }}
      aria-hidden="true"
    />
  );
}

export function SeverityPill({ severidad }: { severidad: string }) {
  const color = SEVERITY_COLOR[severidad] ?? "var(--c-text-3)";
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
      style={{
        background: `color-mix(in srgb, ${color} 14%, var(--c-surface))`,
        border: `1px solid color-mix(in srgb, ${color} 40%, var(--c-border))`,
        color,
      }}
    >
      <SeverityDot severidad={severidad} />
      {SEVERITY_LABEL[severidad] ?? severidad}
    </span>
  );
}
