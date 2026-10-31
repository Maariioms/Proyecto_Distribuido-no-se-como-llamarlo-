"use client";

import { useEffect, useRef, useState } from "react";

// sube el numero poco a poco hasta el valor nuevo en vez de cambiarlo de golpe
// easeOutCubic y ~700ms, y si el usuario pidio menos movimiento salta directo
function useCountUp(target: number, duration = 700) {
  const [value, setValue] = useState(0);
  const fromRef = useRef(0);

  useEffect(() => {
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const from = fromRef.current;
    const start = performance.now();
    let raf = 0;

    const step = (now: number) => {
      const t = reduce ? 1 : Math.min((now - start) / duration, 1);
      const eased = 1 - Math.pow(1 - t, 3);
      const v = Math.round(from + (target - from) * eased);
      fromRef.current = v;
      setValue(v);
      if (t < 1) raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);

    return () => cancelAnimationFrame(raf);
  }, [target, duration]);

  return value;
}

// value null = todavia cargando
// numero = cuenta hacia arriba y da un saltito cuando sube
// texto = se muestra tal cual, ej el uptime
export function StatCard({
  icon,
  label,
  value,
  accentColor,
  hint,
}: {
  icon: React.ReactNode;
  label: string;
  value: number | string | null;
  accentColor?: string;
  hint?: string;
}) {
  const numeric = typeof value === "number" ? value : 0;
  const counted = useCountUp(numeric);

  // el bump compara con el valor anterior durante el render
  // es el patron de React pa no usar un efecto, y solo salta cuando SUBE
  const [prev, setPrev] = useState<number | null>(null);
  const [bumpKey, setBumpKey] = useState(0);
  if (typeof value === "number" && value !== prev) {
    setPrev(value);
    if (prev !== null && value > prev) setBumpKey((k) => k + 1);
  }

  const color = accentColor ?? "var(--c-accent)";

  return (
    <div
      className="hover-lift flex items-center gap-3 rounded-lg border p-4"
      style={{ borderColor: "var(--c-border)", background: "var(--c-surface)" }}
    >
      <div
        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md"
        style={{
          background: `color-mix(in srgb, ${color} 16%, var(--c-surface))`,
          color,
        }}
      >
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-xs" style={{ color: "var(--c-text-2)" }}>
          {label}
        </p>
        {value === null ? (
          <div className="skeleton mt-1 h-6 w-16 rounded" aria-label="Cargando" />
        ) : (
          <p
            className="text-xl font-semibold tabular-nums"
            style={{ color: "var(--c-text)" }}
          >
            <span key={bumpKey} className={bumpKey > 0 ? "bump" : undefined}>
              {typeof value === "number" ? counted : value}
            </span>
          </p>
        )}
        {hint && (
          <p className="text-xs" style={{ color: "var(--c-text-3)" }}>
            {hint}
          </p>
        )}
      </div>
    </div>
  );
}
