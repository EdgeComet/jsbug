import type { TimelineData } from '../../types/content';
import styles from './ResultTabs.module.css';

interface TimelineTabProps {
  data: TimelineData;
}

export function TimelineTab({ data }: TimelineTabProps) {
  const maxTime = Math.max(...data.lifeTimeEvents.map(e => e.time));

  return (
    <div className={styles.timelineList}>
      {data.lifeTimeEvents.map((event, index) => (
        <div key={index} className={styles.timelineRow}>
          <span className={styles.timelineLabel}>{event.event}</span>
          <div className={styles.timelineBarContainer}>
            <div
              className={styles.timelineBar}
              style={{ width: `${(event.time / maxTime) * 100}%` }}
            />
          </div>
          <span className={styles.timelineValue}>{event.time.toFixed(2)}s</span>
        </div>
      ))}
    </div>
  );
}
