import styles from './Header.module.css';

interface AnimatedBugIconProps {
  size?: number;
  isRunning?: boolean;
}

export function AnimatedBugIcon({ size = 25, isRunning = false }: AnimatedBugIconProps) {
  const leftLegsClass = isRunning ? styles.leftLegsRunning : styles.leftLegs;
  const rightLegsClass = isRunning ? styles.rightLegsRunning : styles.rightLegs;

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* Antennae */}
      <path d="M8 2l1.88 1.88" />
      <path d="M14.12 3.88L16 2" />

      {/* Head */}
      <path d="M9 7.13v-1a3.003 3.003 0 116 0v1" />

      {/* Body */}
      <path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 014-4h4a4 4 0 014 4v3c0 3.3-2.7 6-6 6z" />
      <path d="M12 20v-9" />

      {/* Left side connectors */}
      <path d="M6.53 9C4.6 8.8 3 7.1 3 5" />

      {/* Right side connectors */}
      <path d="M21 5c0 2.1-1.6 3.8-3.53 4" />

      {/* Left legs - animated */}
      <g className={leftLegsClass}>
        <path d="M6 13H2" />
        <path d="M6 17H2" />
      </g>

      {/* Right legs - animated */}
      <g className={rightLegsClass}>
        <path d="M18 13h4" />
        <path d="M18 17h4" />
      </g>
    </svg>
  );
}
