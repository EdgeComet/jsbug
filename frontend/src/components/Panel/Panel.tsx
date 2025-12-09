import { useState, useMemo } from 'react';
import { Icon } from '../common/Icon';
import type { PanelData } from '../../hooks/useRenderPanel';
import { PanelHeader } from './PanelHeader';
import { TechnicalCard } from './TechnicalCard';
import { IndexationCard } from './IndexationCard';
import { LinksCard } from './LinksCard';
import { ImagesCard } from './ImagesCard';
import { ContentCard } from './ContentCard';
import { NoJSInfo } from './NoJSInfo';
import { ResultTabs } from '../ResultTabs/ResultTabs';
import { LinksModal, type LinkFilterType } from '../LinksModal/LinksModal';
import { ImagesModal, type ImageFilterType } from '../ImagesModal/ImagesModal';
import { BodyTextModal } from '../BodyTextModal/BodyTextModal';
import { WordDiffModal } from '../WordDiffModal/WordDiffModal';
import { HTMLModal } from '../ResultTabs/HTMLModal';
import { computeWordDiff } from '../../utils/wordDiff';
import styles from './Panel.module.css';

interface PanelProps {
  side: 'left' | 'right';
  isLoading?: boolean;
  error?: string | null;
  data?: PanelData;
  compareData?: PanelData;
  jsEnabled: boolean;
  robotsAllowed?: boolean;
  robotsLoading?: boolean;
}

export function Panel({
  side,
  isLoading,
  error,
  data,
  compareData,
  jsEnabled,
  robotsAllowed,
  robotsLoading,
}: PanelProps) {
  const [linksModalOpen, setLinksModalOpen] = useState(false);
  const [linksModalFilter, setLinksModalFilter] = useState<LinkFilterType>('all');
  const [linksModalDiffType, setLinksModalDiffType] = useState<'all' | 'added' | 'removed'>('all');
  const [imagesModalOpen, setImagesModalOpen] = useState(false);
  const [imagesModalFilter, setImagesModalFilter] = useState<ImageFilterType>('all');
  const [imagesModalDiffType, setImagesModalDiffType] = useState<'all' | 'added' | 'removed'>('all');
  const [bodyTextModalOpen, setBodyTextModalOpen] = useState(false);
  const [wordDiffModalOpen, setWordDiffModalOpen] = useState(false);
  const [wordDiffScrollTo, setWordDiffScrollTo] = useState<'added' | 'removed'>('added');
  const [htmlModalOpen, setHtmlModalOpen] = useState(false);

  const wordDiff = useMemo(() => {
    if (!jsEnabled || !data?.content?.bodyText || !compareData?.content?.bodyText) {
      return { added: [], removed: [] };
    }
    return computeWordDiff(data.content.bodyText, compareData.content.bodyText);
  }, [jsEnabled, data?.content?.bodyText, compareData?.content?.bodyText]);

  const handleOpenLinksModal = (filter: LinkFilterType) => {
    setLinksModalFilter(filter);
    setLinksModalDiffType('all');
    setLinksModalOpen(true);
  };

  const handleOpenLinksDiffModal = (filter: LinkFilterType, diffType: 'added' | 'removed') => {
    setLinksModalFilter(filter);
    setLinksModalDiffType(diffType);
    setLinksModalOpen(true);
  };

  const handleOpenImagesModal = (filter: ImageFilterType) => {
    setImagesModalFilter(filter);
    setImagesModalDiffType('all');
    setImagesModalOpen(true);
  };

  const handleOpenImagesDiffModal = (filter: ImageFilterType, diffType: 'added' | 'removed') => {
    setImagesModalFilter(filter);
    setImagesModalDiffType(diffType);
    setImagesModalOpen(true);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className={`${styles.panel} ${side === 'left' ? styles.panelLeft : styles.panelRight}`} data-side={side}>
        <PanelHeader side={side} />
        <div className={styles.loadingContainer}>
          <Icon name="loader" size={24} className={styles.spinner} />
          <p className={styles.loadingText}>Analyzing page...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={`${styles.panel} ${side === 'left' ? styles.panelLeft : styles.panelRight}`} data-side={side}>
        <PanelHeader side={side} />
        <div className={styles.errorContainer}>
          <Icon name="alert-circle" size={24} className={styles.errorIcon} />
          <p className={styles.panelErrorMessage}>{error}</p>
        </div>
      </div>
    );
  }

  // No data yet
  if (!data) {
    return (
      <div className={`${styles.panel} ${side === 'left' ? styles.panelLeft : styles.panelRight}`} data-side={side}>
        <PanelHeader side={side} />
        <div className={styles.emptyContainer}>
          <p className={styles.emptyText}>Enter a URL and click Analyze to see results</p>
        </div>
      </div>
    );
  }

  const { technical, indexation, links, images, content, network, timeline, console: consoleData } = data;
  const isSuccess = technical.statusCode === 200;

  return (
    <div className={`${styles.panel} ${side === 'left' ? styles.panelLeft : styles.panelRight}`} data-side={side}>
      <PanelHeader side={side} />

      <div className={styles.resultsSection}>
        <div className={styles.resultCards}>
          <TechnicalCard data={technical} compareData={jsEnabled ? compareData?.technical : undefined} onOpenHTMLModal={content.html ? () => setHtmlModalOpen(true) : undefined} />
          {isSuccess && (
            <>
              <IndexationCard data={indexation} compareData={jsEnabled ? compareData?.indexation : undefined} robotsAllowed={robotsAllowed} robotsLoading={robotsLoading} currentUrl={technical.finalUrl} />
              <ContentCard data={content} compareData={jsEnabled ? compareData?.content : undefined} onOpenBodyTextModal={() => setBodyTextModalOpen(true)} onOpenWordDiffModal={(scrollTo) => { setWordDiffScrollTo(scrollTo); setWordDiffModalOpen(true); }} />
              <LinksCard data={links} compareData={jsEnabled ? compareData?.links : undefined} onOpenModal={handleOpenLinksModal} onOpenDiffModal={handleOpenLinksDiffModal} />
              <ImagesCard data={images} compareData={jsEnabled ? compareData?.images : undefined} onOpenModal={handleOpenImagesModal} onOpenDiffModal={handleOpenImagesDiffModal} />
            </>
          )}
        </div>

        {isSuccess && (jsEnabled && network && timeline && consoleData && content.html ? (
          <ResultTabs
            networkData={network}
            timelineData={timeline}
            consoleData={consoleData}
            htmlData={content.html}
            onOpenHTMLModal={() => setHtmlModalOpen(true)}
          />
        ) : (
          <NoJSInfo />
        ))}
      </div>

      <LinksModal
        isOpen={linksModalOpen}
        onClose={() => setLinksModalOpen(false)}
        links={links.links}
        compareLinks={compareData?.links.links}
        initialFilter={linksModalFilter}
        initialDiffType={linksModalDiffType}
      />

      <ImagesModal
        isOpen={imagesModalOpen}
        onClose={() => setImagesModalOpen(false)}
        images={images.images}
        compareImages={compareData?.images.images}
        initialFilter={imagesModalFilter}
        initialDiffType={imagesModalDiffType}
      />

      <BodyTextModal
        isOpen={bodyTextModalOpen}
        onClose={() => setBodyTextModalOpen(false)}
        bodyText={content.bodyText}
        wordCount={content.bodyWords}
      />

      <WordDiffModal
        isOpen={wordDiffModalOpen}
        onClose={() => setWordDiffModalOpen(false)}
        addedWords={wordDiff.added}
        removedWords={wordDiff.removed}
        scrollTo={wordDiffScrollTo}
      />

      {content.html && (
        <HTMLModal
          isOpen={htmlModalOpen}
          onClose={() => setHtmlModalOpen(false)}
          html={content.html}
        />
      )}
    </div>
  );
}
