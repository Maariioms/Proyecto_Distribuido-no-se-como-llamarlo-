// severidad -> datos, no chrome
// los colores viven en globals.css como --sev-* y no cambian con el tema

export const SEVERITIES = ["critical", "high", "medium", "low"] as const;

export const SEVERITY_LABEL: Record<string, string> = {
  critical: "Crítica",
  high: "Alta",
  medium: "Media",
  low: "Baja",
};

export const SEVERITY_COLOR: Record<string, string> = {
  critical: "var(--sev-critical)",
  high: "var(--sev-high)",
  medium: "var(--sev-medium)",
  low: "var(--sev-low)",
};

// ya no se depende solo del color -> cada severidad tiene su propia forma de punto
// pa que se distinga tambien sin ver colores
export const SEVERITY_SHAPE: Record<string, string> = {
  critical: "rounded-none rotate-45",
  high: "rounded-sm",
  medium: "rounded-full",
  low: "rounded-full opacity-70",
};
