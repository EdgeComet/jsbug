import { useState, useEffect, useRef } from 'react'
import { Icon } from './components/common/Icon'
import { ConfigProvider, useConfig } from './context/ConfigContext'
import type { AppConfig } from './types/config'
import { Header } from './components/Header/Header'
import { Panel } from './components/Panel/Panel'
import { ConfigModal } from './components/ConfigModal/ConfigModal'
import { useRenderPanel } from './hooks/useRenderPanel'
import { useRobots } from './hooks/useRobots'
import styles from './App.module.css'

function AppContent() {
  const { config, updateLeftConfig, updateRightConfig } = useConfig()
  const [url, setUrl] = useState('https://example.com/')
  const [isUrlValid, setIsUrlValid] = useState(true)
  const [isConfigOpen, setIsConfigOpen] = useState(false)
  const [hasAnalyzed, setHasAnalyzed] = useState(false)
  const urlInputRef = useRef<HTMLInputElement>(null)

  const leftPanel = useRenderPanel()
  const rightPanel = useRenderPanel()
  const robots = useRobots()

  const isAnalyzing = leftPanel.isLoading || rightPanel.isLoading

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.metaKey && event.key === ',') {
        event.preventDefault()
        setIsConfigOpen(true)
      }
      if (event.metaKey && event.key === 'l') {
        event.preventDefault()
        urlInputRef.current?.focus()
        urlInputRef.current?.select()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])
  const bothPanelsSuccess =
    leftPanel.data?.technical.statusCode === 200 &&
    rightPanel.data?.technical.statusCode === 200

  const handleCompare = (overrideConfig?: AppConfig) => {
    const effectiveConfig = overrideConfig ?? config
    setHasAnalyzed(true)
    leftPanel.reset()
    rightPanel.reset()
    robots.reset()

    // Fire all requests simultaneously (2 renders + 1 robots)
    leftPanel.render(url, effectiveConfig.left)
    rightPanel.render(url, effectiveConfig.right)
    robots.check(url)
  }

  const handleRetryWithBrowserUA = (side: 'left' | 'right') => {
    const updateFn = side === 'left' ? updateLeftConfig : updateRightConfig
    updateFn({ userAgent: 'chrome-mobile' })
    const newConfig: AppConfig = {
      ...config,
      [side]: { ...config[side], userAgent: 'chrome-mobile' },
    }
    handleCompare(newConfig)
  }

  return (
    <div className={styles.app}>
      <Header
        url={url}
        onUrlChange={setUrl}
        onOpenConfig={() => setIsConfigOpen(true)}
        onCompare={() => handleCompare()}
        onUrlValidChange={setIsUrlValid}
        isUrlValid={isUrlValid}
        isAnalyzing={isAnalyzing}
        urlInputRef={urlInputRef}
      />

      {!hasAnalyzed && (
        <div className={styles.welcomeSection}>
          <div className={styles.welcomeContent}>
            <div className={styles.welcomeIcon}>
              <Icon name="bug" size={48} />
            </div>
            <h1 className={styles.welcomeName}>JSBug</h1>
            <p className={styles.welcomeTagline}>Page Render Comparison</p>
            <p className={styles.welcomeDescription}>
              Compare how search engines see your pages with and without JavaScript.
              Analyze rendering differences, content visibility, and SEO implications.
            </p>
          </div>
        </div>
      )}

      {hasAnalyzed && (
        <main className={styles.mainContent}>
          <div className={styles.panelsWrapper}>
            <Panel
              side="left"
              isLoading={leftPanel.isLoading}
              error={leftPanel.error}
              data={leftPanel.data ?? undefined}
              compareData={bothPanelsSuccess ? rightPanel.data ?? undefined : undefined}
              jsEnabled={config.left.jsEnabled}
              robotsAllowed={robots.data?.isAllowed}
              robotsLoading={robots.isLoading}
              onRetryWithBrowserUA={() => handleRetryWithBrowserUA('left')}
            />

            <div className={styles.panelDivider}>
              <span className={styles.dividerLabel}>VS</span>
            </div>

            <Panel
              side="right"
              isLoading={rightPanel.isLoading}
              error={rightPanel.error}
              data={rightPanel.data ?? undefined}
              compareData={bothPanelsSuccess ? leftPanel.data ?? undefined : undefined}
              jsEnabled={config.right.jsEnabled}
              robotsAllowed={robots.data?.isAllowed}
              robotsLoading={robots.isLoading}
              onRetryWithBrowserUA={() => handleRetryWithBrowserUA('right')}
            />
          </div>
        </main>
      )}

      <ConfigModal
        isOpen={isConfigOpen}
        onClose={() => setIsConfigOpen(false)}
        onApply={(newConfig) => {
          if (hasAnalyzed) {
            handleCompare(newConfig)
          }
        }}
      />
    </div>
  )
}

function App() {
  return (
    <ConfigProvider>
      <AppContent />
    </ConfigProvider>
  )
}

export default App
