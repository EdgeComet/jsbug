import { useState, useEffect, useRef } from 'react'
import { Icon } from './components/common/Icon'
import { EscapingBug } from './components/Header/EscapingBug'
import { ConfigProvider, useConfig, defaultConfig, defaultLeftConfig, defaultRightConfig } from './context/ConfigContext'
import { serializeToUrl, parseUrlState } from './utils/urlState'
import { ThemeProvider } from './context/ThemeContext'
import type { AppConfig } from './types/config'
import { Header } from './components/Header/Header'
import { Panel } from './components/Panel/Panel'
import { ConfigModal } from './components/ConfigModal/ConfigModal'
import { TurnstileModal } from './components/TurnstileModal'
import { useRenderPanel } from './hooks/useRenderPanel'
import { useRobots } from './hooks/useRobots'
import { useTurnstile } from './hooks/useTurnstile'
import { useSession } from './hooks/useSession'
import { isCaptchaEnabled } from './config/captcha'
import styles from './App.module.css'

const MAX_SESSION_RETRIES = 3

function AppContent() {
  const { config, setConfig, updateLeftConfig, updateRightConfig } = useConfig()
  const [url, setUrl] = useState('')
  const [isUrlValid, setIsUrlValid] = useState(true)
  const [isConfigOpen, setIsConfigOpen] = useState(false)
  const [hasAnalyzed, setHasAnalyzed] = useState(false)
  const urlInputRef = useRef<HTMLInputElement>(null)
  const initializedRef = useRef(false)

  const leftPanel = useRenderPanel()
  const rightPanel = useRenderPanel()
  const robots = useRobots()
  const turnstile = useTurnstile()
  const session = useSession()

  const isAnalyzing = leftPanel.isLoading || rightPanel.isLoading || leftPanel.isRetrying || rightPanel.isRetrying || turnstile.isLoading

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

  // Parse URL on mount and auto-start if target present
  useEffect(() => {
    if (initializedRef.current) return
    initializedRef.current = true

    const { targetUrl, leftConfig, rightConfig } = parseUrlState(
      window.location.pathname,
      window.location.hash
    )

    if (targetUrl) {
      // Merge parsed config with defaults
      const mergedConfig: AppConfig = {
        left: { ...defaultLeftConfig, ...leftConfig },
        right: { ...defaultRightConfig, ...rightConfig },
      }

      setUrl(targetUrl)
      setConfig(mergedConfig)

      // Auto-start analysis (defer to allow state to settle)
      setTimeout(() => {
        handleCompare(mergedConfig, targetUrl)
      }, 0)
    } else {
      // No target URL, set default
      setUrl('https://example.com/')
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Handle browser back/forward
  useEffect(() => {
    const handlePopState = async () => {
      const { targetUrl, leftConfig, rightConfig } = parseUrlState(
        window.location.pathname,
        window.location.hash
      )

      if (targetUrl) {
        const mergedConfig: AppConfig = {
          left: { ...defaultLeftConfig, ...leftConfig },
          right: { ...defaultRightConfig, ...rightConfig },
        }

        setUrl(targetUrl)
        setConfig(mergedConfig)

        // Trigger re-analysis (don't use handleCompare to avoid pushing new history entry)
        setHasAnalyzed(true)
        leftPanel.reset()
        rightPanel.reset()
        robots.reset()

        // Get session token if captcha is enabled
        let sessionToken: string | undefined
        if (isCaptchaEnabled()) {
          // Try to use existing valid session token
          sessionToken = session.getValidToken() ?? undefined
          if (!sessionToken) {
            // No valid token - get a Turnstile token and create new session
            const turnstileToken = await turnstile.getToken()
            if (turnstileToken === null) return
            const newToken = await session.createSession(turnstileToken)
            if (newToken === null) return
            sessionToken = newToken
          }
        }

        // Use same session token for both panels
        leftPanel.render(targetUrl, mergedConfig.left, sessionToken)
        rightPanel.render(targetUrl, mergedConfig.right, sessionToken)
        robots.check(targetUrl)
      } else {
        // Navigated to root, show welcome
        setUrl('https://example.com/')
        setHasAnalyzed(false)
      }
    }

    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const bothPanelsSuccess =
    leftPanel.data?.technical.statusCode === 200 &&
    rightPanel.data?.technical.statusCode === 200

  const handleCompare = async (overrideConfig?: AppConfig, urlOverride?: string, retryCount = 0) => {
    const effectiveConfig = overrideConfig ?? config
    const effectiveUrl = urlOverride ?? url

    // Cancel any active pool exhaustion retries
    leftPanel.cancelRetry()
    rightPanel.cancelRetry()

    // Update browser URL (only on first attempt, not retries)
    if (retryCount === 0) {
      const newUrl = serializeToUrl(effectiveUrl, effectiveConfig, defaultConfig)
      window.history.pushState(null, '', newUrl)
    }

    // Show loading state immediately (reset sets isLoading: true)
    setHasAnalyzed(true)
    leftPanel.reset()
    rightPanel.reset()
    robots.reset()

    // Get session token if captcha is enabled
    let sessionToken: string | undefined
    if (isCaptchaEnabled()) {
      // Try to use existing valid session token
      sessionToken = session.getValidToken() ?? undefined
      if (!sessionToken) {
        // No valid token - get a Turnstile token and create new session
        const turnstileToken = await turnstile.getToken()
        if (turnstileToken === null) {
          return // User cancelled or timeout
        }
        const newToken = await session.createSession(turnstileToken)
        if (newToken === null) {
          return // Session creation failed
        }
        sessionToken = newToken
      }
    }

    // Fire all requests simultaneously using the same session token
    const leftPromise = leftPanel.render(effectiveUrl, effectiveConfig.left, sessionToken)
    const rightPromise = rightPanel.render(effectiveUrl, effectiveConfig.right, sessionToken)
    robots.check(effectiveUrl)

    // Wait for both panels to complete
    await Promise.all([leftPromise, rightPromise])

    // Check for session token errors and retry silently if needed
    const isSessionTokenError = (error: string | null) =>
      error?.includes('SESSION_TOKEN_REQUIRED') ||
      error?.includes('SESSION_TOKEN_INVALID') ||
      error?.includes('SESSION_TOKEN_EXPIRED')

    if ((isSessionTokenError(leftPanel.error) || isSessionTokenError(rightPanel.error))
        && retryCount < MAX_SESSION_RETRIES) {
      // Clear invalid session and retry with new one
      session.clearSession()
      return await handleCompare(overrideConfig, urlOverride, retryCount + 1)
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

  // Retry just one panel (for "Try Again" button after pool exhaustion)
  const retryPanel = async (side: 'left' | 'right') => {
    const panel = side === 'left' ? leftPanel : rightPanel
    const panelConfig = side === 'left' ? config.left : config.right

    // Get session token (reuse existing if valid)
    let sessionToken: string | undefined
    if (isCaptchaEnabled()) {
      sessionToken = session.getValidToken() ?? undefined
      if (!sessionToken) {
        const turnstileToken = await turnstile.getToken()
        if (turnstileToken === null) return
        const newToken = await session.createSession(turnstileToken)
        if (newToken === null) return
        sessionToken = newToken
      }
    }

    panel.render(url, panelConfig, sessionToken)
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
              isRetrying={leftPanel.isRetrying}
              retryCount={leftPanel.retryCount}
              onRetry={() => retryPanel('left')}
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
              isRetrying={rightPanel.isRetrying}
              retryCount={rightPanel.retryCount}
              onRetry={() => retryPanel('right')}
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
