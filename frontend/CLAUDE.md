# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
npm run dev          # Start development server (Vite)
npm run build        # TypeScript check + Vite production build
npm run lint         # ESLint - run before committing
npm run test         # Vitest in watch mode
npm run test:coverage # Vitest with coverage
```

Run a single test file:
```bash
npx vitest run src/components/Header/Header.test.tsx
```

## Tech Stack

- React 19 with TypeScript
- Vite 6 for build/dev
- ESLint 9 with typescript-eslint for linting
- Vitest with jsdom for testing, @testing-library/react
- Recharts for data visualization
- Lucide React for icons

## Architecture

This is a page comparison tool that displays two side-by-side panels comparing URL renders (JS-enabled vs JS-disabled).

### State Management
- `ConfigContext` (`src/context/ConfigContext.tsx`) - React Context for app configuration
- `AppConfig` contains `left` and `right` `PanelConfig` objects with settings like jsEnabled, userAgent, timeout, waitFor event, and blocking options

### Component Structure
- `App.tsx` - Root component with ConfigProvider, manages loading state
- `Panel` - Main display component for each side (left/right), receives data props for technical, indexation, links, content, network, timeline, console, and HTML
- `ResultTabs` - Tabbed interface for Network, Timeline, Console, and HTML views (only shown when JS is enabled)
- `ConfigModal` - Settings modal for configuring each panel

### Types
- `src/types/config.ts` - Configuration types (UserAgent, WaitEvent, PanelConfig, AppConfig)
- `src/types/content.ts` - Content types (TechnicalData, IndexationData, LinksData, ContentData, TimelineData)
- `src/types/network.ts` - Network request types
- `src/types/console.ts` - Console entry types

### CSS
- Uses CSS Modules (*.module.css files alongside components)

### Principles
- DRY, do not duplicate the functionality that already exists. Move it into utilities layer, etc
- think before adding new parameters to components
- Single Responsibility: One component, one purpose 
- Composition over Inheritance: Prefer composition patterns 
- Props Interface Design: Clear, typed prop interfaces
- Custom Hooks: Extract reusable logic 
- Error Boundaries: Graceful error handling 
- Accessibility: ARIA labels, semantic HTML

### Performance Optimization
- React.memo: Prevent unnecessary re-renders
- useMemo/useCallback: Memoize expensive operations
- Code Splitting: Lazy load components
