import type { TechnicalData, IndexationData, LinksData, ImagesData, ContentData, TimelineData } from '../types/content';
import type { NetworkData } from '../types/network';
import type { ConsoleEntry } from '../types/console';

export const mockLeftPanel = {
  technical: {
    statusCode: 200,
    pageSize: 376832,      // 368 KB in bytes
    loadTime: 2.13,
  } as TechnicalData,

  indexation: {
    metaIndexable: true,
    metaFollow: true,
    canonicalUrl: 'https://jsbug.org/',
    hrefLangs: [
      { lang: 'en', url: 'https://jsbug.org/', source: 'html' },
      { lang: 'es', url: 'https://jsbug.org/es/', source: 'html' },
      { lang: 'fr', url: 'https://jsbug.org/fr/', source: 'html' },
    ],
    isIndexable: true,
    indexabilityReason: 'Page is indexable',
  } as IndexationData,

  links: {
    links: [
      // Internal links (47)
      { href: '/', text: 'Home', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/features', text: 'Features', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/pricing', text: 'Pricing', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/about', text: 'About Us', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/contact', text: 'Contact', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/blog', text: 'Blog', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/docs', text: 'Documentation', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/api', text: 'API Reference', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/login', text: 'Login', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/signup', text: 'Sign Up', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/privacy', text: 'Privacy Policy', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/terms', text: 'Terms of Service', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/careers', text: 'Careers', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/partners', text: 'Partners', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/case-studies', text: 'Case Studies', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/integrations', text: 'Integrations', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/webinars', text: 'Webinars', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/resources', text: 'Resources', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/support', text: 'Support', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/faq', text: 'FAQ', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/changelog', text: 'Changelog', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/status', text: 'Status', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/security', text: 'Security', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/enterprise', text: 'Enterprise', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/demo', text: 'Request Demo', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/comparison', text: 'Comparison', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/tutorials', text: 'Tutorials', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/guides', text: 'Guides', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/videos', text: 'Videos', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/podcast', text: 'Podcast', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/newsletter', text: 'Newsletter', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/community', text: 'Community', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/events', text: 'Events', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/press', text: 'Press', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/awards', text: 'Awards', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/testimonials', text: 'Testimonials', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/team', text: 'Team', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/investors', text: 'Investors', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/affiliate', text: 'Affiliate Program', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/referral', text: 'Referral Program', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/sitemap', text: 'Sitemap', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/accessibility', text: 'Accessibility', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/cookies', text: 'Cookie Policy', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/gdpr', text: 'GDPR', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/images/logo.png', text: '', isExternal: false, isDofollow: true, isImageLink: true, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/images/hero.webp', text: '', isExternal: false, isDofollow: true, isImageLink: true, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/images/feature.svg', text: '', isExternal: false, isDofollow: true, isImageLink: true, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      // External links (12)
      { href: 'https://twitter.com/jsbug', text: 'Twitter', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://linkedin.com/company/jsbug', text: 'LinkedIn', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://facebook.com/jsbug', text: 'Facebook', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://github.com/jsbug', text: 'GitHub', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://youtube.com/jsbug', text: 'YouTube', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://instagram.com/jsbug', text: 'Instagram', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://google.com/analytics', text: 'Google Analytics', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://stripe.com', text: 'Stripe', isExternal: true, isDofollow: true, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://aws.amazon.com', text: 'AWS', isExternal: true, isDofollow: true, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://cloudflare.com', text: 'Cloudflare', isExternal: true, isDofollow: true, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://developer.mozilla.org', text: 'MDN Docs', isExternal: true, isDofollow: true, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://web.dev', text: 'Web.dev', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
    ],
  } as LinksData,

  images: {
    images: [
      { src: 'https://jsbug.org/images/logo.png', alt: 'Logo', isExternal: false, isAbsolute: false, isInLink: true, linkHref: '/', size: 12456 },
      { src: 'https://jsbug.org/images/hero.webp', alt: 'Hero image', isExternal: false, isAbsolute: false, isInLink: false, linkHref: '', size: 159744 },
      { src: 'https://jsbug.org/images/feature.svg', alt: 'Feature illustration', isExternal: false, isAbsolute: false, isInLink: false, linkHref: '', size: 8234 },
    ],
  } as ImagesData,

  content: {
    title: 'JSBug - JavaScript Rendering Comparison Tool',
    metaDescription: 'JSBug helps developers and SEO professionals compare how pages render with and without JavaScript. Identify rendering differences before they impact your users.',
    h1: ['JavaScript Rendering Comparison Tool'],
    h2: [
      'Side-by-Side Comparison',
      'Network Analysis',
      'Console Debugging',
      'Performance Timeline',
      'Start Comparing Today',
    ],
    h3: [],
    bodyWords: 2647,
    bodyTextTokensCount: 1985,
    textHtmlRatio: 0.42,
    bodyText: `JavaScript Rendering Comparison Tool

Compare how your pages render with JavaScript enabled versus disabled. JSBug provides a side-by-side view to help you identify content differences, performance impacts, and potential issues.

Side-by-Side Comparison
View your page rendered in two panels simultaneously. The left panel shows the fully rendered JavaScript version while the right panel displays the static HTML. Spot missing content, broken layouts, and accessibility concerns at a glance.

Network Analysis
Track every network request made during page rendering. See which scripts, stylesheets, and API calls are loaded. Identify blocked resources, failed requests, and opportunities to optimize your page load performance.

Console Debugging
Capture all console output including logs, warnings, and errors. Debug JavaScript issues by comparing what happens when scripts execute versus when they're blocked. Find runtime errors before your users do.

Performance Timeline
Monitor key performance metrics like First Paint, DOM Content Loaded, and Time to Interactive. Understand how JavaScript affects your page load times and core web vitals.

Start Comparing Today
Enter any URL and instantly see the rendering differences. Perfect for debugging client-side rendering issues, validating server-side rendering implementations, and ensuring your content is accessible to all users and search engines.

Key Features:
- Real-time rendering comparison
- Network request monitoring
- Console log capture
- Performance metrics tracking
- Custom user agent selection
- Configurable wait conditions
- Request blocking options`,
    bodyMarkdown: `# JavaScript Rendering Comparison Tool

Compare how your pages render with JavaScript enabled versus disabled. JSBug provides a side-by-side view to help you identify content differences, performance impacts, and potential issues.

## Side-by-Side Comparison

View your page rendered in two panels simultaneously. The left panel shows the fully rendered JavaScript version while the right panel displays the static HTML. Spot missing content, broken layouts, and accessibility concerns at a glance.

## Network Analysis

Track every network request made during page rendering. See which scripts, stylesheets, and API calls are loaded. Identify blocked resources, failed requests, and opportunities to optimize your page load performance.

## Console Debugging

Capture all console output including logs, warnings, and errors. Debug JavaScript issues by comparing what happens when scripts execute versus when they're blocked. Find runtime errors before your users do.

## Performance Timeline

Monitor key performance metrics like First Paint, DOM Content Loaded, and Time to Interactive. Understand how JavaScript affects your page load times and core web vitals.

## Start Comparing Today

Enter any URL and instantly see the rendering differences. Perfect for debugging client-side rendering issues, validating server-side rendering implementations, and ensuring your content is accessible to all users and search engines.

### Key Features:
- Real-time rendering comparison
- Network request monitoring
- Console log capture
- Performance metrics tracking
- Custom user agent selection
- Configurable wait conditions
- Request blocking options`,
  } as ContentData,

  network: {
    requests: [
      { id: '1', url: 'https://jsbug.org/', status: 200, type: 'document', size: 152576, time: 0.245, isInternal: true },
      { id: '2', url: 'https://jsbug.org/assets/app.js', status: 200, type: 'script', size: 319488, time: 0.180, isInternal: true },
      { id: '3', url: 'https://jsbug.org/assets/style.css', status: 200, type: 'stylesheet', size: 46080, time: 0.120, isInternal: true },
      { id: '4', url: 'https://www.googletagmanager.com/gtag/js', status: 'blocked', type: 'script', size: null, time: null, blocked: true, isInternal: false },
      { id: '5', url: 'https://jsbug.org/api/render', status: 200, type: 'xhr', size: 2150, time: 0.340, isInternal: true },
      { id: '6', url: 'https://jsbug.org/images/hero.webp', status: 200, type: 'image', size: 159744, time: 0.290, isInternal: true },
      { id: '7', url: 'https://fonts.gstatic.com/inter.woff2', status: 200, type: 'font', size: 24576, time: 0.095, isInternal: false },
      { id: '8', url: 'https://connect.facebook.net/signals/config', status: 'blocked', type: 'script', size: null, time: null, blocked: true, isInternal: false },
    ],
  } as NetworkData,

  timeline: {
    lifeTimeEvents: [
      { event: 'First Paint', time: 0.18 },
      { event: 'First Contentful Paint', time: 0.22 },
      { event: 'DOMContentLoaded', time: 0.245 },
      { event: 'Largest Contentful Paint', time: 0.65 },
      { event: 'Load', time: 0.75 },
      { event: 'Time to Interactive', time: 1.2 },
      { event: 'Network Idle', time: 2.13 },
    ],
  } as TimelineData,

  console: [
    { id: '1', level: 'warn', message: '[Deprecation] The window.webkitStorageInfo API has been removed. Please use navigator.storage.estimate() instead. This API was removed in Chrome 108 and is no longer available. For more information, see https://developer.chrome.com/blog/deprecating-and-removing-webkitstorageinfo/', time: 0.125 },
    { id: '2', level: 'warn', message: 'A cookie associated with a cross-site resource was set without the `SameSite` attribute.', time: 0.340 },
    { id: '3', level: 'log', message: 'JSBug initialized successfully', time: 0.450 },
    { id: '4', level: 'log', message: 'Rendering comparison started', time: 0.520 },
    { id: '5', level: 'error', message: 'Failed to load resource: net::ERR_BLOCKED_BY_CLIENT', time: 0.680 },
    { id: '6', level: 'error', message: 'Uncaught TypeError: Cannot read properties of undefined (reading \'innerHTML\')\n    at RenderPanel (webpack://app/./src/components/RenderPanel.tsx:45:23)\n    at renderWithHooks (webpack://app/./node_modules/react-dom/cjs/react-dom.development.js:14985:18)\n    at mountIndeterminateComponent (webpack://app/./node_modules/react-dom/cjs/react-dom.development.js:17811:13)\n    at beginWork (webpack://app/./node_modules/react-dom/cjs/react-dom.development.js:19049:16)\n    at HTMLUnknownElement.callCallback (webpack://app/./node_modules/react-dom/cjs/react-dom.development.js:3945:14)', time: 0.892 },
    { id: '7', level: 'log', message: 'API Response: {"status":"success","data":{"url":"https://jsbug.org","jsEnabled":true,"renderTime":1.24,"contentLength":152576}}', time: 1.2 },
    { id: '8', level: 'warn', message: 'React does not recognize the `customProp` prop on a DOM element. If you intentionally want it to appear in the DOM as a custom attribute, spell it as lowercase `customprop` instead. If you accidentally passed it from a parent component, remove it from the DOM element.', time: 1.5 },
    { id: '9', level: 'log', message: 'Performance metrics collected', time: 1.8 },
    { id: '10', level: 'error', message: 'CORS policy: No \'Access-Control-Allow-Origin\' header is present on the requested resource at https://api.third-party.com/v2/tracking. Origin \'https://jsbug.org\' is therefore not allowed access. If an opaque response serves your needs, set the request\'s mode to \'no-cors\' to fetch the resource with CORS disabled.', time: 2.1 },
  ] as ConsoleEntry[],

  html: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JSBug - JavaScript Rendering Comparison Tool</title>
    <meta name="description" content="Compare how pages render with and without JavaScript...">
    <link rel="canonical" href="https://jsbug.org/">
    <meta name="robots" content="index, follow">
</head>
<body>
    <header class="main-header">
        <nav class="navbar">...</nav>
    </header>
    <main>
        <section class="hero">
            <h1>JavaScript Rendering Comparison Tool</h1>
            <p>Compare JS-enabled and JS-disabled rendering side by side...</p>
        </section>
        <!-- Additional content rendered by JavaScript -->
    </main>
    <footer>...</footer>
</body>
</html>`,
};

export const mockRightPanel = {
  technical: {
    statusCode: 200,
    pageSize: 149,
    loadTime: 0.46,
  } as TechnicalData,

  indexation: {
    metaIndexable: true,
    metaFollow: true,
    canonicalUrl: 'https://jsbug.org/',
    hrefLangs: [],
    isIndexable: true,
    indexabilityReason: 'Page is indexable',
  } as IndexationData,

  links: {
    links: [
      // Internal links (18)
      { href: '/', text: 'Home', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/features', text: 'Features', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/pricing', text: 'Pricing', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/about', text: 'About Us', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/contact', text: 'Contact', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/blog', text: 'Blog', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/docs', text: 'Documentation', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/login', text: 'Login', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/signup', text: 'Sign Up', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/privacy', text: 'Privacy Policy', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/terms', text: 'Terms of Service', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/support', text: 'Support', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/faq', text: 'FAQ', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/sitemap', text: 'Sitemap', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/images/logo.png', text: '', isExternal: false, isDofollow: true, isImageLink: true, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/images/hero.webp', text: '', isExternal: false, isDofollow: true, isImageLink: true, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/api', text: 'API', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      { href: '/demo', text: 'Demo', isExternal: false, isDofollow: true, isImageLink: false, isAbsolute: false, isSocial: false, isUgc: false, isSponsored: false },
      // External links (5)
      { href: 'https://twitter.com/jsbug', text: 'Twitter', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://linkedin.com/company/jsbug', text: 'LinkedIn', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://github.com/jsbug', text: 'GitHub', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: true, isUgc: false, isSponsored: false },
      { href: 'https://google.com/analytics', text: 'Analytics', isExternal: true, isDofollow: false, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
      { href: 'https://stripe.com', text: 'Stripe', isExternal: true, isDofollow: true, isImageLink: false, isAbsolute: true, isSocial: false, isUgc: false, isSponsored: false },
    ],
  } as LinksData,

  images: {
    images: [
      { src: 'https://jsbug.org/images/logo.png', alt: 'Logo', isExternal: false, isAbsolute: false, isInLink: true, linkHref: '/', size: 12456 },
      { src: 'https://jsbug.org/images/hero.webp', alt: 'Hero image', isExternal: false, isAbsolute: false, isInLink: false, linkHref: '', size: 159744 },
    ],
  } as ImagesData,

  content: {
    title: 'JSBug - JavaScript Rendering Comparison Tool',
    metaDescription: 'JSBug helps developers and SEO professionals compare how pages render with and without JavaScript. Identify rendering differences before they impact your users.',
    h1: ['JavaScript Rendering Comparison Tool'],
    h2: [
      'Side-by-Side Comparison',
      'Network Analysis',
      'Console Debugging',
    ],
    h3: [],
    bodyWords: 847,
    bodyTextTokensCount: 635,
    textHtmlRatio: 0.38,
    bodyText: `JavaScript Rendering Comparison Tool

Compare how your pages render with JavaScript enabled versus disabled.

Side-by-Side Comparison
View your page rendered in two panels simultaneously.

Network Analysis
Track every network request made during page rendering.

Console Debugging
Capture all console output including logs, warnings, and errors.`,
    bodyMarkdown: `# JavaScript Rendering Comparison Tool

Compare how your pages render with JavaScript enabled versus disabled.

## Side-by-Side Comparison

View your page rendered in two panels simultaneously.

## Network Analysis

Track every network request made during page rendering.

## Console Debugging

Capture all console output including logs, warnings, and errors.`,
  } as ContentData,
};
