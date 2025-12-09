import { Logo } from './Logo';
import { URLInput } from './URLInput';
import { ConfigButton } from './ConfigButton';
import { CompareButton } from './CompareButton';
import styles from './Header.module.css';

interface HeaderProps {
  url: string;
  onUrlChange: (url: string) => void;
  onOpenConfig: () => void;
  onCompare: () => void;
  onUrlValidChange?: (isValid: boolean) => void;
  isUrlValid?: boolean;
  isAnalyzing?: boolean;
  urlInputRef?: React.RefObject<HTMLInputElement | null>;
}

export function Header({ url, onUrlChange, onOpenConfig, onCompare, onUrlValidChange, isUrlValid = true, isAnalyzing = false, urlInputRef }: HeaderProps) {
  return (
    <header className={styles.appHeader}>
      <div className={styles.headerContent}>
        <Logo />
        <URLInput ref={urlInputRef} value={url} onChange={onUrlChange} onValidChange={onUrlValidChange} onSubmit={isUrlValid && !isAnalyzing ? onCompare : undefined} />
        <div className={styles.headerActions}>
          <ConfigButton onClick={onOpenConfig} />
          <CompareButton onClick={onCompare} disabled={!isUrlValid || isAnalyzing} />
        </div>
      </div>
    </header>
  );
}
