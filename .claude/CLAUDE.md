# Personal Assist (Memoria) - Claude.md

## Project Overview

**Personal Assist** is a local-only AI assistant for macOS that captures, transcribes, and organizes user activities like meetings, work sessions, and calls. The marketing name is **Memoria** with the tagline "Your private, adaptive, AI assistant."

**Core Value Proposition**: Privacy-first AI assistant that runs completely locally, never sending data to external servers, while providing intelligent transcription and activity management.

## Architecture & Technology Stack

### Core Technologies
- **Framework**: Wails v2 (Go backend + React frontend)
- **Backend**: Go 1.23+
- **Frontend**: React with modern JavaScript
- **Database**: SQLite with WAL mode
- **Transcription**: Whisper.cpp (local processing only)
- **Platform**: macOS (initial target)
- **API**: None - purely local desktop application

### Data Model (Activity-Centric)
```
User (UUID, settings)
├── Activities (meetings, work_sessions, calls)
│   ├── AudioRecordings (linked to activities)
│   └── TranscriptChunks (processed from audio)
```

### Directory Structure
```
personal-assist/
├── CLAUDE.md               # Development documentation  
├── README.md               # Run instructions
└── desktop-app/            # Main Wails application
    ├── go.mod              # github.com/biancarosa/personal-assist
    ├── main.go             # Application entry point
    ├── app.go              # Wails app context
    ├── models/             # Data models (User, Activity, AudioRecording, TranscriptChunk)
    ├── services/           # Business logic layer
    ├── storage/            # SQLite operations & file management
    ├── database/           # Schema, migrations, connections
    ├── frontend/           # React UI
    └── build/              # Built binaries
```

## Data Storage

### Local Storage Structure
```
~/Library/Application Support/personal-assist/
├── personal-assist.db      # SQLite database
├── activities/             # Activity-organized files
│   └── {activity_id}/
│       └── audio/
│           ├── recording_001.m4a
│           └── recording_002.m4a
└── models/                 # Whisper models
    ├── whisper-tiny.bin
    └── whisper-small.bin
```

### Key Principles
- **Local-only**: No external APIs required for any functionality
- **Privacy-first**: All data stays on user's device permanently
- **Activity-centric**: Everything organized around user activities
- **Offline-capable**: Works without internet connection
- **No cloud dependency**: Never sends data to external servers

## Business Model

### Freemium Desktop Software (No API Required)
- **Free Tier**: 5 hours/month transcription, basic features
- **Pro Tier**: $15/month - unlimited transcription, AI summaries, advanced features
- **License-based**: Local validation without requiring server infrastructure

### Revenue Streams
1. Pro license subscriptions (license key system)
2. One-time lifetime licenses  
3. Future: Enterprise features for teams

### Competitive Advantage
- **No ongoing server costs** - purely local processing
- **Complete privacy** - data never leaves device
- **Works offline** - no internet dependency
- **One-time purchase options** - not just subscriptions

## Technical Implementation Notes

### Audio Recording
- Use Go audio libraries (portaudio or similar)
- Support multiple audio devices
- macOS-specific permissions handling
- Real-time recording with proper cleanup

### Transcription Pipeline
- Whisper.cpp Go bindings for local processing
- Model management (download, cache, select)
- Progress tracking for long transcriptions
- Activity-relative timestamp correlation
- **No external API calls** - completely local

### Database Design
- SQLite with WAL mode for better concurrency
- Full-text search with FTS5 for transcript search
- Foreign key constraints for data integrity  
- Optimized indexes for common queries
- Complete CRUD operations implemented

### Frontend Architecture
- React components for activity management
- Wails bindings for Go backend communication
- Responsive design for macOS
- Local search and filtering
- Real-time updates via Wails events

## Privacy & Security

### Data Protection
- All processing happens locally on device
- No network requests for core functionality
- No telemetry or analytics collection
- Encryption for sensitive data at rest
- User controls data retention and deletion

### User Trust Building
- Transparent about data handling
- Open source components where possible
- Clear privacy policy
- No tracking or analytics by default
- **Truly air-gapped operation**

## API Design (Internal Wails Methods Only)

### Core Wails Methods
```go
// Activity Management
CreateActivity(actType, title string) (*Activity, error)
StartActivity(activityID string) error
StopActivity(activityID string) error
GetActivities(filter ActivityFilter) ([]*Activity, error)

// Recording
StartRecording(activityID string) (*AudioRecording, error)
StopRecording(recordingID string) error
GetRecording(recordingID string) (*AudioRecording, error)

// Transcription (Local Only)
GetTranscript(activityID string) ([]*TranscriptChunk, error)
SearchTranscripts(query string) (*SearchResults, error)
ProcessRecording(recordingID string) error

// System
GetAudioDevices() ([]AudioDevice, error)
GetSettings() (*UserSettings, error)
UpdateSettings(settings *UserSettings) error
```

## Development Guidelines

### Code Quality
- Use meaningful variable names and clear function signatures
- Write comprehensive error handling
- Include unit tests for critical functionality
- Document complex algorithms and business logic

### Performance Considerations
- Optimize for local processing efficiency
- Minimize memory usage during long recordings
- Implement progress tracking for user feedback
- Cache frequently accessed data appropriately

### macOS Integration
- Proper permissions handling (microphone, screen recording)
- System tray integration with native feel
- Respect macOS design guidelines
- Handle system sleep/wake events gracefully

## Testing Strategy

### Core Functionality Testing
- Audio recording reliability across different devices
- Transcription accuracy with various audio qualities
- Database operations and data integrity
- UI responsiveness and error handling

### Integration Testing
- End-to-end activity workflow
- System permission handling
- Background processing behavior
- App lifecycle management

## Deployment & Distribution

### Initial Distribution
- Unsigned macOS app bundles
- Direct download from website (memoria.4every1.ai)
- GitHub releases for updates
- Clear installation instructions

### Future Distribution
- Mac App Store (if beneficial)
- Homebrew cask for developers
- Notarized builds for smoother installation

## Success Metrics

### Technical Goals
- Reliable audio capture (>99% success rate)
- High transcription accuracy (>95% for clear audio)
- Low resource usage (<200MB RAM, <10% CPU idle)
- Fast search response times (<500ms)

### Business Goals
- Reach $100 MRR within 3 months of launch
- Maintain <5% monthly churn for paying users
- Achieve 5-10% freemium conversion rate
- Build sustainable, profitable business

## Competitive Positioning

### Key Differentiators
- **Complete privacy**: No cloud dependency whatsoever
- **Local processing**: Works offline permanently
- **Activity-centric**: Organized around user workflows
- **Transparent**: Clear about data handling
- **No subscriptions required**: Lifetime license options

### Target Users
- Privacy-conscious professionals
- Knowledge workers with sensitive information
- Remote workers in frequent meetings
- Researchers and consultants
- Anyone wanting AI assistance without cloud risks
- Users concerned about data sovereignty

### vs. Competitors
- **vs. Otter.ai**: "Your meetings stay on YOUR device"
- **vs. Notion AI**: "Understands YOUR work, not everyone's"
- **vs. ChatGPT Plus**: "No copying transcripts to external services"

## Future Roadmap

### Potential Features
- Multi-language support
- Custom Whisper model training
- Integration with popular productivity tools
- Team collaboration features (while maintaining privacy)
- Mobile companion apps (view-only)
- Browser extension for web meeting capture

### Platform Expansion
- Windows support
- Linux support
- iOS/Android companion apps (view-only)
- Web interface for viewing (local network only)

## Marketing Positioning

### Core Messages
- **"Your private, adaptive, AI assistant"**
- **"Finally, an AI that's actually yours"**
- **"AI assistance without surveillance"**
- **"Your data stays on your device, period"**

### Unique Value Props
- Works completely offline
- No monthly fees for core features
- Data ownership guaranteed
- Privacy by design, not as afterthought

## Contact & Support

- **Developer**: Bianca Rosa (@biancarosa)
- **Repository**: github.com/platformlabs/ai-personal-assistant
- **Marketing Site**: memoria.4every1.ai
- **Support**: Email-based for MVP

# Memoria UI/UX Design System & Style Guide

## Design Philosophy

**Memoria (Personal Assist)** follows a **macOS-native design language** that emphasizes:
- **Privacy-first visual identity**: Clean, trustworthy, professional aesthetics that reinforce data security
- **Activity-centric clarity**: Clear information hierarchy focused on user workflows and sessions
- **Desktop-native sophistication**: Polish and refinement that feels at home on macOS
- **Accessibility**: WCAG-compliant contrast ratios, keyboard navigation, and inclusive design
- **Minimal distraction**: Subtle interactions that support focus and productivity

## Brand Identity

### Core Values Expressed Through Design
- **Privacy**: Secure, local-only messaging with green accent indicators
- **Intelligence**: Sophisticated blue color palette suggesting AI capabilities
- **Simplicity**: Clean layouts with generous whitespace and clear hierarchy
- **Trust**: Professional typography and consistent interactions
- **Efficiency**: Quick access to features through compact, well-organized interfaces

## Color System

### Primary Brand Colors
```css
/* Memoria Blue - Primary Brand Color */
--primary-50: #eef2ff;   /* Lightest tint - backgrounds */
--primary-100: #e0e7ff;  /* Light backgrounds */
--primary-200: #c7d2fe;  /* Subtle accents */
--primary-300: #a5b4fc;  /* Muted interactive elements */
--primary-400: #818cf8;  /* Secondary buttons */
--primary-500: #6366f1;  /* Base brand color - PRIMARY */
--primary-600: #4f46e5;  /* Primary buttons, links */
--primary-700: #4338ca;  /* Hover states, active elements */
--primary-800: #3730a3;  /* Strong accents */
--primary-900: #312e81;  /* Darkest accents */
```

### Light Mode Color Palette
```css
/* Backgrounds */
--background: #ffffff;           /* Main app background */
--card: #ffffff;                 /* Card backgrounds */
--secondary: #f5f5f7;           /* Secondary backgrounds */
--muted: #ececf0;               /* Muted backgrounds */
--accent: #e9ebef;              /* Accent backgrounds */

/* Text Colors */
--foreground: #030213;          /* Primary text */
--muted-foreground: #717182;    /* Secondary text */
--secondary-foreground: #030213; /* Text on secondary backgrounds */

/* Interactive Elements */
--input-background: #f3f3f5;    /* Input field backgrounds */
--border: rgba(0, 0, 0, 0.1);   /* Light borders */
--switch-background: #cbced4;    /* Toggle switches */
```

### Dark Mode Color Palette (Refined for Comfort)
```css
/* Backgrounds - Soft blue-tinted grays */
--background: oklch(0.09 0.005 264);      /* Main app background */
--card: oklch(0.12 0.005 264);            /* Card backgrounds */
--secondary: oklch(0.18 0.005 264);       /* Secondary backgrounds */
--muted: oklch(0.16 0.005 264);           /* Muted backgrounds */
--accent: oklch(0.18 0.005 264);          /* Accent backgrounds */

/* Text Colors - Reduced contrast for comfort */
--foreground: oklch(0.95 0.005 264);      /* Primary text */
--muted-foreground: oklch(0.65 0.005 264); /* Secondary text */
--secondary-foreground: oklch(0.85 0.005 264); /* Text on secondary */

/* Interactive Elements */
--input-background: oklch(0.14 0.005 264); /* Input backgrounds */
--border: oklch(0.2 0.005 264);           /* Subtle borders */
--switch-background: oklch(0.3 0.005 264); /* Toggle switches */
```

### Status Colors
```css
/* Privacy/Security - Green */
--success-50: #ecfdf5;    --success-500: #10b981;    --success-700: #047857;
--success-foreground: #ffffff;

/* Recording/Active - Red */
--destructive: oklch(0.55 0.15 25);   /* Recording indicators */
--destructive-foreground: oklch(0.95 0.005 264);

/* Warning/Caution - Amber */
--warning-50: #fffbeb;    --warning-500: #f59e0b;    --warning-700: #b45309;

/* Information - Blue (Brand) */
--info: var(--primary-500);
--info-foreground: #ffffff;
```

## Typography Scale

### Font Stack
```css
/* Primary: macOS System Font */
--font-family-system: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;

/* Monospace: Code & technical content */
--font-family-mono: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
```

### Font Sizes (14px base for desktop)
```css
--font-size: 14px;                    /* Base font size */

/* Scale based on 14px base */
--text-xs: 0.75rem;     /* 10.5px - Small labels, badges */
--text-sm: 0.875rem;    /* 12.25px - Secondary text, captions */
--text-base: 1rem;      /* 14px - Primary body text */
--text-lg: 1.125rem;    /* 15.75px - Emphasized text */
--text-xl: 1.25rem;     /* 17.5px - Section headings */
--text-2xl: 1.5rem;     /* 21px - Page headings */
--text-3xl: 1.875rem;   /* 26.25px - Hero text */
```

### Font Weights
```css
--font-weight-normal: 400;    /* Regular body text */
--font-weight-medium: 500;    /* UI elements, emphasized text */
--font-weight-semibold: 600;  /* Headings (not currently used) */
--font-weight-bold: 700;      /* Strong emphasis (not currently used) */
```

### Typography Hierarchy
```css
/* Auto-applied to elements without Tailwind text classes */
h1 { font-size: var(--text-2xl); font-weight: var(--font-weight-medium); }
h2 { font-size: var(--text-xl); font-weight: var(--font-weight-medium); }
h3 { font-size: var(--text-lg); font-weight: var(--font-weight-medium); }
h4 { font-size: var(--text-base); font-weight: var(--font-weight-medium); }
p  { font-size: var(--text-base); font-weight: var(--font-weight-normal); }
```

## Spacing System

### 8px Grid System
```css
/* Consistent spacing based on 8px increments */
--spacing-0: 0;
--spacing-1: 0.25rem;   /* 4px - Tight spacing */
--spacing-2: 0.5rem;    /* 8px - Small gaps */
--spacing-3: 0.75rem;   /* 12px - Medium-small gaps */
--spacing-4: 1rem;      /* 16px - Standard spacing */
--spacing-5: 1.25rem;   /* 20px - Medium spacing */
--spacing-6: 1.5rem;    /* 24px - Large spacing */
--spacing-8: 2rem;      /* 32px - Section spacing */
--spacing-10: 2.5rem;   /* 40px - Large section spacing */
--spacing-12: 3rem;     /* 48px - Major section spacing */
--spacing-16: 4rem;     /* 64px - Page-level spacing */
--spacing-20: 5rem;     /* 80px - Hero spacing */
```

## Border Radius System

### Rounded Corner Scale
```css
--radius: 0.625rem;                /* 10px - Base radius */

/* Calculated variations */
--radius-sm: calc(var(--radius) - 4px);  /* 6px - Small elements */
--radius-md: calc(var(--radius) - 2px);  /* 8px - Medium elements */
--radius-lg: var(--radius);              /* 10px - Standard cards/buttons */
--radius-xl: calc(var(--radius) + 4px);  /* 14px - Large cards */
--radius-2xl: 1rem;                      /* 16px - Hero elements */
--radius-3xl: 1.5rem;                    /* 24px - Large containers */
--radius-full: 9999px;                   /* Fully rounded - pills */
```

### Border Radius Usage
- **Small elements**: `--radius-sm` (badges, small buttons)
- **Standard elements**: `--radius-lg` (cards, inputs, standard buttons)
- **Large elements**: `--radius-xl` (activity cards, containers)
- **Pills**: `--radius-full` (nav buttons, status indicators)

## Component Design System

### Title Bar
```css
.title-bar {
  background: var(--background);
  border-bottom: 1px solid var(--border);
  padding: var(--spacing-3) var(--spacing-6);
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px; /* Fixed height for consistency */
}

.title-bar__nav-button {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  padding: var(--spacing-2) var(--spacing-4);
  min-width: 100px;
  transition: all 150ms ease;
}

.title-bar__nav-button.active {
  background: var(--primary-600);
  color: white;
  border-color: var(--primary-600);
}
```

### Activity Cards
```css
.activity-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);  /* 14px for larger cards */
  padding: var(--spacing-6);
  transition: all 200ms ease;
  position: relative;
}

.activity-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

/* Recording state indicator */
.activity-card--recording::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: var(--destructive);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
}
```

### Buttons

#### Primary Action Buttons
```css
.btn-primary {
  background: var(--primary-600);
  color: white;
  border: none;
  border-radius: var(--radius-lg);
  padding: var(--spacing-3) var(--spacing-5);
  font-weight: var(--font-weight-medium);
  transition: all 150ms ease;
}

.btn-primary:hover {
  background: var(--primary-700);
  transform: translateY(-1px);
}

.btn-primary:active {
  transform: translateY(0);
}
```

#### Secondary Buttons
```css
.btn-secondary {
  background: var(--card);
  color: var(--foreground);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3) var(--spacing-5);
  font-weight: var(--font-weight-medium);
  transition: all 150ms ease;
}

.btn-secondary:hover {
  background: var(--secondary);
  transform: translateY(-1px);
}
```

#### Quick Action Buttons (Record, etc.)
```css
.btn-record {
  background: var(--primary-600);
  color: white;
  border-radius: var(--radius-full);
  padding: var(--spacing-2) var(--spacing-4);
  font-size: var(--text-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.btn-record--recording {
  background: var(--destructive);
}

.btn-record:hover {
  transform: translateY(-1px);
}
```

### Form Elements

#### Input Fields
```css
.form-input {
  background: var(--input-background);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);  /* More rounded for friendly feel */
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--text-base);
  transition: all 150ms ease;
}

.form-input:focus {
  border-color: var(--primary-500);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  outline: none;
}
```

#### Search Fields
```css
.search-input {
  background: var(--secondary);
  border: 1px solid transparent;
  border-radius: var(--radius-2xl);  /* Extra rounded */
  padding: var(--spacing-4) var(--spacing-5);
  font-size: var(--text-lg);
  width: 100%;
}
```

### Privacy Indicators
```css
.privacy-indicator {
  background: color-mix(in srgb, var(--success-500) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--success-500) 20%, transparent);
  border-radius: var(--radius-xl);
  padding: var(--spacing-4);
}

.privacy-indicator__title {
  color: var(--success-700);
  font-weight: var(--font-weight-medium);
}
```

## Layout Patterns

### Desktop App Layout
```css
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;  /* Prevent scrolling on main container */
}

.app-main {
  flex: 1;
  overflow: auto;  /* Allow scrolling in main content area */
}
```

### Dashboard Layout
```css
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-8) var(--spacing-6);
}

.dashboard__hero {
  background: var(--card);
  border-radius: var(--radius-2xl);
  padding: var(--spacing-10);
  margin-bottom: var(--spacing-12);
  text-align: center;
}

.dashboard__quick-actions {
  display: flex;
  gap: var(--spacing-4);
  justify-content: center;
  margin-top: var(--spacing-8);
}
```

### Activity Grid
```css
.activity-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: var(--spacing-8);
  margin-top: var(--spacing-8);
}

/* Responsive: Single column on smaller screens */
@media (max-width: 768px) {
  .activity-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-6);
  }
}
```

### Settings Layout
```css
.settings {
  max-width: 1000px;
  margin: 0 auto;
  padding: var(--spacing-8) var(--spacing-6);
}

.settings__card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-2xl);
  padding: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}
```

## Status Bar Design
```css
.status-bar {
  background: var(--card);
  border-top: 1px solid var(--border);
  padding: var(--spacing-2) var(--spacing-6);
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  font-size: var(--text-sm);
  color: var(--muted-foreground);
}

.status-bar__recording {
  color: var(--destructive);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}
```

## Animation & Micro-interactions

### Timing Functions
```css
--transition-fast: 150ms ease;
--transition-normal: 250ms ease;
--transition-slow: 350ms ease;

/* Enhanced easing for smooth interactions */
--ease-out: cubic-bezier(0, 0, 0.2, 1);
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
```

### Hover Effects
```css
/* Lift effect for interactive cards */
.hover-lift:hover {
  transform: translateY(-2px);
  transition: transform var(--transition-fast);
}

/* Subtle lift for buttons */
.hover-lift-subtle:hover {
  transform: translateY(-1px);
  transition: transform var(--transition-fast);
}
```

### Focus States
```css
.focus-ring:focus-visible {
  outline: 2px solid var(--primary-500);
  outline-offset: 2px;
}

/* Hide focus for mouse interactions */
.focus-ring:focus:not(:focus-visible) {
  outline: none;
}
```

## Scrollbar Styling (Desktop Native)
```css
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--muted);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: var(--muted-foreground);
  border-radius: 4px;
  opacity: 0.5;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--foreground);
  opacity: 0.7;
}
```

## Dark Mode Implementation

### Theme Switching
```tsx
// ThemeProvider Context
const ThemeContext = createContext<{
  theme: 'light' | 'dark' | 'system';
  actualTheme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
}>();

// Automatic system theme detection
useEffect(() => {
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleChange = (e: MediaQueryListEvent) => {
    setSystemTheme(e.matches ? 'dark' : 'light');
  };
  mediaQuery.addEventListener('change', handleChange);
}, []);
```

### CSS Custom Properties Strategy
```css
/* Single source of truth for theme values */
:root { /* Light mode values */ }
.dark { /* Dark mode overrides */ }

/* Tailwind integration */
@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  /* ... */
}
```

## Accessibility Standards

### Color Contrast Compliance
- **Primary text**: 4.5:1 minimum contrast ratio
- **Secondary text**: 3:1 minimum contrast ratio
- **Interactive elements**: 3:1 minimum contrast ratio
- **Status indicators**: Color + icon/text for accessibility

### Keyboard Navigation
```css
/* Focus management */
:focus-visible {
  outline: 2px solid var(--primary-500);
  outline-offset: 2px;
}

/* Tab order optimization */
.tab-order {
  /* Ensure logical tab flow through interface */
}
```

### Screen Reader Support
```tsx
// Semantic HTML structure
<main role="main" aria-label="Dashboard">
  <section aria-labelledby="recent-activities">
    <h2 id="recent-activities">Recent Activities</h2>
    {/* ... */}
  </section>
</main>

// Status announcements
<div aria-live="polite" aria-atomic="true">
  {isRecording ? 'Recording in progress' : 'Recording stopped'}
</div>
```

## Implementation Guidelines

### CSS Architecture
```
styles/
├── globals.css          # Design tokens, base styles, typography
└── components/          # Component-specific styles (if needed)
```

### Component Organization
```
components/
├── ui/                  # Reusable ShadCN components
├── Dashboard.tsx        # Page-level components
├── ActivityCard.tsx     # Feature-specific components
└── TitleBar.tsx         # Layout components
```

### Tailwind Class Patterns
```css
/* Common utility combinations */
.card-base: bg-card border border-border rounded-xl p-6
.button-base: px-4 py-3 rounded-lg font-medium transition-all
.input-base: bg-input-background border border-border rounded-xl px-4 py-3
.text-muted: text-muted-foreground text-sm
```

## Design Tokens Quick Reference

### Spacing Classes
```css
/* Padding: p-2, p-4, p-6, p-8, p-10, p-12 */
/* Margin: m-0, mb-4, mt-8, mx-auto */
/* Gap: gap-2, gap-4, gap-6, gap-8 */
```

### Typography Classes
```css
/* Sizes: text-xs, text-sm, text-base, text-lg, text-xl, text-2xl */
/* Weights: font-normal, font-medium */
/* Colors: text-foreground, text-muted-foreground */
```

### Layout Classes
```css
/* Flexbox: flex, flex-col, items-center, justify-between */
/* Grid: grid, grid-cols-1, grid-cols-2, gap-4 */
/* Sizing: w-full, h-screen, max-w-4xl, min-h-0 */
```

### Color Classes
```css
/* Backgrounds: bg-background, bg-card, bg-secondary, bg-muted */
/* Text: text-foreground, text-muted-foreground, text-primary */
/* Borders: border-border, border-primary */
```

## Responsive Breakpoints

### Desktop-First Approach
```css
/* Base styles for desktop (1200px+) */
.desktop-layout { grid-template-columns: repeat(3, 1fr); }

/* Medium screens (768px - 1199px) */
@media (max-width: 1199px) {
  .desktop-layout { grid-template-columns: repeat(2, 1fr); }
}

/* Small screens (< 768px) */
@media (max-width: 767px) {
  .desktop-layout { grid-template-columns: 1fr; }
}
```

## Component State Variants

### Button States
```css
/* Default */
.btn { /* base styles */ }

/* Hover */
.btn:hover { transform: translateY(-1px); }

/* Active */
.btn:active { transform: translateY(0); }

/* Disabled */
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Loading */
.btn.loading { /* spinner or loading state */ }
```

### Recording States
```css
/* Default state */
.recording-indicator { opacity: 0; }

/* Recording active */
.recording-indicator.active {
  opacity: 1;
  color: var(--destructive);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```

---

*This design system ensures Memoria maintains a consistent, professional, and privacy-focused visual identity that feels native to macOS while supporting the app's core mission of local, secure activity management.*

---

## Development Status & Progress

For detailed implementation status, progress tracking, and development roadmap, see:
- **[DEVELOPMENT_STATUS.md](.claude/DEVELOPMENT_STATUS.md)** - Comprehensive status of UI implementation, backend integration progress, and 6-phase development roadmap

## Additional Documentation

- **[Go Desktop Architect Agent](agents/go-desktop-architect.md)** - Specialized agent for Go backend architecture
- **[Desktop UX Designer Agent](agents/desktop-ux-designer.md)** - Specialized agent for UI/UX design

---

# Memories

- I'm running the app with wails dev as you code, so if you need me to test something let me know and then afterwards you can inspect logs to see what happened in the action I took.
- Don't overdo with emojis. Not every button needs an emoji, mate.
- DO NOT ADD EMOJIS TO EVERY BUTTON.
- If initial task was retrieved from markdown roadmap please mark it as done in the same roadmap