import { cn } from "@/lib/utils";

// Logo mark: a rounded tile with an upward trend glyph.
export function LogoMark({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm",
        className
      )}
    >
      <svg viewBox="0 0 24 24" fill="none" className="h-[60%] w-[60%]">
        <path d="M4 16.5 9 11l3.5 3L20 6.5" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M15 6.5h5v5" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </span>
  );
}

interface BrandProps {
  className?: string;
  mark?: string;
  text?: string;
}

export function Brand({ className, mark = "h-9 w-9", text = "text-lg" }: BrandProps) {
  return (
    <div className={cn("flex items-center gap-2.5", className)}>
      <LogoMark className={mark} />
      <span className={cn("font-semibold tracking-tight", text)}>
        Simpli<span className="text-primary">Trade</span>
      </span>
    </div>
  );
}
