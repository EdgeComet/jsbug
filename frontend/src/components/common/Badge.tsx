import styles from './Badge.module.css';

type BadgeVariant = 'status' | 'type' | 'index';
type StatusType = 'success' | 'warning' | 'error' | 'blocked' | 'failed' | '200' | '301' | '302' | '404' | '500';
type ResourceType = 'document' | 'script' | 'stylesheet' | 'image' | 'font' | 'xhr' | 'other';
type IndexType = 'yes' | 'no';

interface BadgeProps {
  variant: BadgeVariant;
  type?: StatusType | ResourceType | IndexType;
  children: React.ReactNode;
  className?: string;
}

export function Badge({ variant, type, children, className }: BadgeProps) {
  const classNames = [
    styles.badge,
    styles[variant],
    type && styles[type],
    className,
  ].filter(Boolean).join(' ');

  return <span className={classNames}>{children}</span>;
}

// Specialized badge components for convenience
export function StatusBadge({ status, children }: { status: StatusType; children: React.ReactNode }) {
  return <Badge variant="status" type={status}>{children}</Badge>;
}

export function TypeBadge({ type, children }: { type: ResourceType; children: React.ReactNode }) {
  return <Badge variant="type" type={type}>{children}</Badge>;
}

export function IndexBadge({ indexed }: { indexed: boolean }) {
  return (
    <Badge variant="index" type={indexed ? 'yes' : 'no'}>
      {indexed ? 'Yes' : 'No'}
    </Badge>
  );
}
