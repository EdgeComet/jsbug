import { UrlLink } from './UrlLink';
import styles from './HrefLangDiffValue.module.css';

interface HrefLangEntry {
  lang: string;
  url: string;
  source: string;
}

interface HrefLangDiffValueProps {
  value: HrefLangEntry[];
  compareValue?: HrefLangEntry[];
}

function getKey(entry: HrefLangEntry): string {
  return `${entry.lang}|${entry.url}`;
}

export function HrefLangDiffValue({ value, compareValue }: HrefLangDiffValueProps) {
  if (value.length === 0 && (!compareValue || compareValue.length === 0)) {
    return <span className={styles.empty}>empty</span>;
  }

  if (!compareValue) {
    if (value.length === 0) {
      return <span className={styles.empty}>empty</span>;
    }
    return (
      <div className={styles.list}>
        {value.map((entry, i) => (
          <div key={i}>{entry.lang} → <UrlLink url={entry.url} /></div>
        ))}
      </div>
    );
  }

  const compareSet = new Set(compareValue.map(getKey));
  const valueSet = new Set(value.map(getKey));

  // Items only in compareValue (missing from this panel)
  const missing = compareValue.filter(entry => !valueSet.has(getKey(entry)));

  return (
    <div className={styles.list}>
      {value.map((entry, i) => (
        <div key={i} className={!compareSet.has(getKey(entry)) ? styles.added : ''}>
          {entry.lang} → <UrlLink url={entry.url} />
        </div>
      ))}
      {missing.map((entry, i) => (
        <div key={`missing-${i}`} className={styles.missing}>
          {entry.lang} → {entry.url}
        </div>
      ))}
    </div>
  );
}

export function hrefLangsEqual(a: HrefLangEntry[], b: HrefLangEntry[]): boolean {
  if (a.length !== b.length) return false;
  const setB = new Set(b.map(getKey));
  return a.every(entry => setB.has(getKey(entry)));
}
