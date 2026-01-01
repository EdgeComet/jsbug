import React, { useMemo, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import type { Components } from 'react-markdown';
import highlightStyles from '../../styles/highlight.module.css';
import styles from './MarkdownContent.module.css';

interface MarkdownContentProps {
  content: string;
  searchTerm?: string;
  activeMatchIndex?: number;
}

export const MarkdownContent: React.FC<MarkdownContentProps> = ({
  content,
  searchTerm,
  activeMatchIndex
}) => {
  const matchRefs = useRef<HTMLElement[]>([]);
  const matchCounterRef = useRef(0);

  // Reset refs array before each render
  matchRefs.current = [];
  matchCounterRef.current = 0;

  // Scroll to active match when it changes
  useEffect(() => {
    if (activeMatchIndex !== undefined && matchRefs.current[activeMatchIndex]) {
      matchRefs.current[activeMatchIndex].scrollIntoView({
        behavior: 'instant',
        block: 'center',
      });
    }
  }, [activeMatchIndex]);

  // Helper to highlight text matches with ref tracking
  const highlightMatches = (text: string, term: string): React.ReactNode => {
    if (!term || !text) return text;

    const escapedTerm = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(${escapedTerm})`, 'gi');
    const parts = text.split(regex);

    return parts.map((part, i) => {
      if (regex.test(part)) {
        const currentIndex = matchCounterRef.current++;
        const isActive = currentIndex === activeMatchIndex;
        return (
          <mark
            key={i}
            ref={(el) => { if (el) matchRefs.current[currentIndex] = el; }}
            className={isActive ? highlightStyles.highlightActive : highlightStyles.highlight}
          >
            {part}
          </mark>
        );
      }
      return part;
    });
  };

  // Process React children to highlight text nodes
  const processChildren = (children: React.ReactNode, term: string): React.ReactNode => {
    return React.Children.map(children, (child) => {
      if (typeof child === 'string') {
        return highlightMatches(child, term);
      }
      if (React.isValidElement<{ children?: React.ReactNode }>(child) && child.props.children) {
        return React.cloneElement(child, {
          ...child.props,
          children: processChildren(child.props.children, term),
        });
      }
      return child;
    });
  };

  const components: Components = useMemo(() => ({
    // Override text rendering to add highlights
    p: ({ children }) => {
      if (!searchTerm) return <p>{children}</p>;
      return <p>{processChildren(children, searchTerm)}</p>;
    },
    li: ({ children }) => {
      if (!searchTerm) return <li>{children}</li>;
      return <li>{processChildren(children, searchTerm)}</li>;
    },
    h1: ({ children }) => {
      if (!searchTerm) return <h1>{children}</h1>;
      return <h1>{processChildren(children, searchTerm)}</h1>;
    },
    h2: ({ children }) => {
      if (!searchTerm) return <h2>{children}</h2>;
      return <h2>{processChildren(children, searchTerm)}</h2>;
    },
    h3: ({ children }) => {
      if (!searchTerm) return <h3>{children}</h3>;
      return <h3>{processChildren(children, searchTerm)}</h3>;
    },
    // Non-clickable links
    a: ({ href, children }) => (
      <span className={styles.link} title={href}>
        {searchTerm ? processChildren(children, searchTerm) : children}
      </span>
    ),
    hr: () => <hr className={styles.sectionDivider} />,
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }), [searchTerm, activeMatchIndex]);

  return (
    <div className={styles.markdownContent}>
      <ReactMarkdown components={components}>
        {content}
      </ReactMarkdown>
    </div>
  );
};
