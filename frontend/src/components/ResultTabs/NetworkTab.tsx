import { useState, useMemo } from 'react';
import type { NetworkData, ResourceType } from '../../types/network';
import { calculateNetworkSummary, calculateResourceTypeSummary, formatBytes } from '../../utils/networkUtils';
import { NetworkModal, type NetworkStatusFilter, type NetworkInternalFilter, type SortColumn, type SortDirection } from './NetworkModal';
import styles from './ResultTabs.module.css';

interface NetworkTabProps {
  data: NetworkData;
}

interface OpenModalOptions {
  statusFilter?: NetworkStatusFilter;
  internalFilter?: NetworkInternalFilter;
  typeFilter?: ResourceType | 'all';
  sort?: { column: SortColumn; direction: SortDirection };
}

const RESOURCE_TYPE_LABELS: Record<string, string> = {
  document: 'Document',
  script: 'Script',
  stylesheet: 'Stylesheet',
  xhr: 'XHR',
  fetch: 'Fetch',
  image: 'Image',
  font: 'Font',
  media: 'Media',
  websocket: 'WebSocket',
  other: 'Other',
};

function getResourceTypeLabel(type: string): string {
  return RESOURCE_TYPE_LABELS[type] || type.charAt(0).toUpperCase() + type.slice(1);
}

export function NetworkTab({ data }: NetworkTabProps) {
  const [modalOpen, setModalOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState<NetworkStatusFilter>('all');
  const [internalFilter, setInternalFilter] = useState<NetworkInternalFilter>('all');
  const [typeFilter, setTypeFilter] = useState<ResourceType | 'all'>('all');
  const [sortColumn, setSortColumn] = useState<SortColumn | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  const summary = useMemo(() => calculateNetworkSummary(data.requests), [data.requests]);
  const resourceTypeSummaries = useMemo(() => calculateResourceTypeSummary(data.requests), [data.requests]);

  const handleOpenModal = (options: OpenModalOptions = {}) => {
    setStatusFilter(options.statusFilter ?? 'all');
    setInternalFilter(options.internalFilter ?? 'all');
    setTypeFilter(options.typeFilter ?? 'all');
    setSortColumn(options.sort?.column ?? null);
    setSortDirection(options.sort?.direction ?? 'asc');
    setModalOpen(true);
  };

  return (
    <div className={styles.networkTab}>
      {/* Summary Stats */}
      <div className={styles.networkSummaryRow}>
        <button
          type="button"
          className={styles.networkStatClickable}
          onClick={() => handleOpenModal()}
        >
          <span className={styles.statValue}>{summary.requests}</span>
          <span className={styles.statLabel}>requests</span>
        </button>
        <button
          type="button"
          className={styles.networkStatClickable}
          onClick={() => handleOpenModal({ internalFilter: 'external' })}
        >
          <span className={styles.statValue}>{summary.external}</span>
          <span className={styles.statLabel}>external</span>
        </button>
        <button
          type="button"
          className={styles.networkStatClickable}
          onClick={() => handleOpenModal({ sort: { column: 'size', direction: 'desc' } })}
        >
          <span className={styles.statValue}>{formatBytes(summary.transferred)}</span>
          <span className={styles.statLabel}>Transferred</span>
        </button>
      </div>

      {/* Resource Type Table */}
      <table className={styles.resourceTypeTable}>
        <thead>
          <tr>
            <th>Resource type</th>
            <th>Requests</th>
            <th>Data loaded</th>
            <th>Failed</th>
          </tr>
        </thead>
        <tbody>
          {resourceTypeSummaries.map((typeSummary) => (
            <tr
              key={typeSummary.type}
              className={styles.resourceTypeRow}
              onClick={() => handleOpenModal({ typeFilter: typeSummary.type })}
            >
              <td>{getResourceTypeLabel(typeSummary.type)}</td>
              <td>{typeSummary.requests}</td>
              <td>{formatBytes(typeSummary.dataLoaded)}</td>
              <td className={typeSummary.failed > 0 ? styles.statError : ''}>
                {typeSummary.failed}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <NetworkModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        requests={data.requests}
        totalRequests={summary.requests}
        initialStatusFilter={statusFilter}
        initialInternalFilter={internalFilter}
        initialTypeFilter={typeFilter}
        initialSortColumn={sortColumn}
        initialSortDirection={sortDirection}
      />
    </div>
  );
}
