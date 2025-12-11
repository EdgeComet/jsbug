import { useState, useEffect, useRef } from 'react'
import { Icon } from './components/common/Icon'
import { EscapingBug } from './components/Header/EscapingBug'
import { ConfigProvider, useConfig } from './context/ConfigContext'
import { ThemeProvider } from './context/ThemeContext'
import type { AppConfig } from './types/config'
import { Header } from './components/Header/Header'
import { Panel } from './components/Panel/Panel'
import { ConfigModal } from './components/ConfigModal/ConfigModal'
import { TurnstileModal } from './components/TurnstileModal'
import { useRenderPanel } from './hooks/useRenderPanel'
import { useRobots } from './hooks/useRobots'
import { useTurnstile } from './hooks/useTurnstile'
import { isCaptchaEnabled } from './config/captcha'
import styles from './App.module.css'

const MAX_CAPTCHA_RETRIES = 3

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
  const turnstile = useTurnstile()

  const isAnalyzing = leftPanel.isLoading || rightPanel.isLoading || turnstile.isLoading

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

  const handleCompare = async (overrideConfig?: AppConfig, retryCount = 0) => {
    const effectiveConfig = overrideConfig ?? config

    // Show loading state immediately (reset sets isLoading: true)
    setHasAnalyzed(true)
    leftPanel.reset()
    rightPanel.reset()
    robots.reset()

    // Get captcha token if enabled
    let captchaToken: string | undefined
    if (isCaptchaEnabled()) {
      const token = await turnstile.getToken()
      if (token === null) {
        // User cancelled or timeout - don't proceed
        return
      }
      captchaToken = token
    }

    // Fire all requests simultaneously (both use same captcha token)
    const leftPromise = leftPanel.render(url, effectiveConfig.left, captchaToken)
    const rightPromise = rightPanel.render(url, effectiveConfig.right, captchaToken)
    robots.check(url)

    // Wait for both panels to complete
    await Promise.all([leftPromise, rightPromise])

    // Check for captcha token errors and retry silently if needed
    // Note: CAPTCHA_SERVICE_UNAVAILABLE (503) is NOT retried - that's a server error
    const isCaptchaTokenError = (error: string | null) =>
      error?.includes('CAPTCHA_REQUIRED') || error?.includes('CAPTCHA_INVALID')

    if ((isCaptchaTokenError(leftPanel.error) || isCaptchaTokenError(rightPanel.error))
        && retryCount < MAX_CAPTCHA_RETRIES) {
      // Silent retry - get new token and try again
      return await handleCompare(overrideConfig, retryCount + 1)
    }
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
              <EscapingBug />
            </div>
            <h1 className={styles.welcomeName}>JSBug</h1>
            <p className={styles.welcomeHeadline}>See What Search Engines &amp; AI Bots See</p>
            <p className={styles.welcomeSubheadline}>
              Debug JavaScript rendering issues before they hurt your SEO
            </p>

            <div className={styles.featureCards}>
              <div className={styles.featureCard}>
                <div className={styles.featureCardIcon}>
                  <Icon name="columns" size={28} />
                </div>
                <h3 className={styles.featureCardTitle}>Compare Side-by-Side</h3>
                <p className={styles.featureCardDesc}>
                  View raw HTML vs JavaScript-rendered content instantly
                </p>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureCardIcon}>
                  <Icon name="search" size={28} />
                </div>
                <h3 className={styles.featureCardTitle}>Catch SEO Issues</h3>
                <p className={styles.featureCardDesc}>
                  Find missing titles, meta tags, and content hidden from crawlers
                </p>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureCardIcon}>
                  <Icon name="git-compare" size={28} />
                </div>
                <h3 className={styles.featureCardTitle}>Track Every Change</h3>
                <p className={styles.featureCardDesc}>
                  See exactly which links and elements JavaScript modifies
                </p>
              </div>
            </div>

            <p className={styles.openSourceTagline}>
              Free & open source. Built by the community.
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

      {/* Turnstile Captcha Modal - only shown when challenge needed */}
      <TurnstileModal
        isOpen={turnstile.showModal}
        containerRef={turnstile.modalContainerRef}
        onClose={turnstile.closeModal}
      />
    </div>
  )
}

function App() {
  return (
    <ThemeProvider>
      <ConfigProvider>
        <AppContent />
      </ConfigProvider>
    </ThemeProvider>
  )
}

export default App
