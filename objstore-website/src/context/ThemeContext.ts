import { createContext } from 'react';

export type ResolvedTheme = 'light' | 'dark';
export type ThemePreference = 'light' | 'dark' | 'system';

export interface ThemeContextValue {
  theme: ResolvedTheme;
  preference: ThemePreference;
  setPreference: (p: ThemePreference) => void;
}

export const ThemeContext = createContext<ThemeContextValue | null>(null);
