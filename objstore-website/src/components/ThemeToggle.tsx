import { useState, useEffect, useRef } from 'react';
import { useTheme } from '../hooks/useTheme';

const SunIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
  </svg>
);

const MoonIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
  </svg>
);

const SystemIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
  </svg>
);

const options = [
  { value: 'light' as const, label: 'Light', Icon: SunIcon },
  { value: 'dark' as const, label: 'Dark', Icon: MoonIcon },
  { value: 'system' as const, label: 'System', Icon: SystemIcon },
];

export default function ThemeToggle() {
  const { theme, preference, setPreference } = useTheme();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const ActiveIcon = preference === 'light' ? SunIcon : preference === 'dark' ? MoonIcon : theme === 'dark' ? MoonIcon : SunIcon;

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        className="p-2 rounded-lg text-text-muted hover:text-text hover:bg-overlay-hover transition-colors"
        aria-label="Theme settings"
        title="Theme settings"
      >
        <ActiveIcon />
      </button>

      {open && (
        <div className="absolute top-full mt-2 right-0 w-36 bg-bg-card border border-border rounded-lg shadow-xl overflow-hidden z-50">
          {options.map(({ value, label, Icon }) => {
            const isActive = preference === value;
            return (
              <button
                key={value}
                onClick={() => {
                  setPreference(value);
                  setOpen(false);
                }}
                className={`w-full flex items-center gap-2.5 px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'text-primary bg-primary/10'
                    : 'text-text-muted hover:text-text hover:bg-overlay'
                }`}
              >
                <Icon />
                <span>{label}</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
