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
import { isCaptchaEnabled } from './config/captcha'
import styles from './App.module.css'

const MAX_CAPTCHA_RETRIES = 3

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
    const handlePopState = () => {
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
        leftPanel.render(targetUrl, mergedConfig.left)
        rightPanel.render(targetUrl, mergedConfig.right)
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
    const leftPromise = leftPanel.render(effectiveUrl, effectiveConfig.left, captchaToken)
    const rightPromise = rightPanel.render(effectiveUrl, effectiveConfig.right, captchaToken)
    robots.check(effectiveUrl)

    // Wait for both panels to complete
    await Promise.all([leftPromise, rightPromise])

    // Check for captcha token errors and retry silently if needed
    // Note: CAPTCHA_SERVICE_UNAVAILABLE (503) is NOT retried - that's a server error
    const isCaptchaTokenError = (error: string | null) =>
      error?.includes('CAPTCHA_REQUIRED') || error?.includes('CAPTCHA_INVALID')

    if ((isCaptchaTokenError(leftPanel.error) || isCaptchaTokenError(rightPanel.error))
        && retryCount < MAX_CAPTCHA_RETRIES) {
      // Silent retry - get new token and try again
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
