import { Sun, Moon, Monitor } from 'lucide-react';
import { useTheme } from './ThemeProvider';

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  const getThemeIcon = () => {
    switch (theme) {
      case 'light':
        return <Sun className="w-4 h-4" />;
      case 'dark':
        return <Moon className="w-4 h-4" />;
      case 'system':
        return <Monitor className="w-4 h-4" />;
    }
  };

  return (
    <div className="relative">
      <button
        onClick={() => {
          const themes: Array<'light' | 'dark' | 'system'> = ['light', 'dark', 'system'];
          const currentIndex = themes.indexOf(theme);
          const nextTheme = themes[(currentIndex + 1) % themes.length];
          setTheme(nextTheme);
        }}
        className="inline-flex items-center justify-center rounded-lg w-8 h-8 text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        title={`Current theme: ${theme}. Click to switch.`}
      >
        {getThemeIcon()}
        <span className="sr-only">Toggle theme</span>
      </button>
    </div>
  );
}