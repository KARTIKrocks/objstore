import { useState, useEffect, useSyncExternalStore, useCallback, type ReactNode } from 'react';
import { ThemeContext, type ResolvedTheme, type ThemePreference } from '../context/ThemeContext';

function getInitialPreference(): ThemePreference {
  const stored = localStorage.getItem('theme');
  if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
  return 'system';
}

// Subscribe to OS color scheme changes via useSyncExternalStore
const mq = window.matchMedia('(prefers-color-scheme: dark)');

function subscribeSystemTheme(cb: () => void) {
  mq.addEventListener('change', cb);
  return () => mq.removeEventListener('change', cb);
}

function getSystemThemeSnapshot(): ResolvedTheme {
  return mq.matches ? 'dark' : 'light';
}

function applyTheme(resolved: ResolvedTheme) {
  document.documentElement.setAttribute('data-theme', resolved);
}

export default function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>(getInitialPreference);
  const systemTheme = useSyncExternalStore(subscribeSystemTheme, getSystemThemeSnapshot);

  // Derive resolved theme — no setState needed
  const theme: ResolvedTheme = preference === 'system' ? systemTheme : preference;

  // Sync to DOM and localStorage (side effects only, no setState)
  useEffect(() => {
    applyTheme(theme);
    localStorage.setItem('theme', preference);
  }, [theme, preference]);

  const setPreference = useCallback((p: ThemePreference) => {
    setPreferenceState(p);
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, preference, setPreference }}>
      {children}
    </ThemeContext.Provider>
  );
}
