import styles from './TextValue.module.css';

interface TextValueProps {
  value: string | string[] | null | undefined;
  listClassName?: string;
}

export function TextValue({ value, listClassName }: TextValueProps) {
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return <span className={styles.empty}>empty</span>;
    }
    return (
      <div className={listClassName}>
        {value.map((item, index) => (
          <span key={index}>{item}</span>
        ))}
      </div>
    );
  }

  if (!value || value.trim() === '') {
    return <span className={styles.empty}>empty</span>;
  }
  return <>{value}</>;
}
