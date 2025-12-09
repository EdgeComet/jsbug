import { useState, useEffect, useMemo, useRef } from 'react';
import { Icon } from '../common/Icon';
import type { NetworkRequest, ResourceType } from '../../types/network';
import { StatusBadge, TypeBadge } from '../common/Badge';
import { PageSize } from '../common/PageSize';
import { LoadTime } from '../common/LoadTime';
import { Modal } from '../common/Modal';
import { FilterInput } from '../common/FilterInput';
import filterStyles from '../common/Modal/ModalFilters.module.css';
import styles from './NetworkModal.module.css';

export type NetworkStatusFilter = 'all' | 'blocked' | 'failed';
export type NetworkInternalFilter = 'all' | 'internal' | 'external';
export type SortColumn = 'url' | 'status' | 'type' | 'size' | 'time';
export type SortDirection = 'asc' | 'desc';

interface NetworkModalProps {
  isOpen: boolean;
  onClose: () => void;
  requests: NetworkRequest[];
  totalRequests: number;
  initialStatusFilter?: NetworkStatusFilter;
  initialInternalFilter?: NetworkInternalFilter;
  initialTypeFilter?: ResourceType | 'all';
  initialSortColumn?: SortColumn | null;
  initialSortDirection?: SortDirection;
}

export function NetworkModal({
  isOpen,
  onClose,
  requests,
  totalRequests,
  initialStatusFilter = 'all',
  initialInternalFilter = 'all',
  initialTypeFilter = 'all',
  initialSortColumn = null,
  initialSortDirection = 'asc'
}: NetworkModalProps) {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [filter, setFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState<ResourceType | 'all'>(initialTypeFilter);
  const [statusFilter, setStatusFilter] = useState<NetworkStatusFilter>(initialStatusFilter);
  const [internalFilter, setInternalFilter] = useState<NetworkInternalFilter>(initialInternalFilter);
  const [sortColumn, setSortColumn] = useState<SortColumn | null>(initialSortColumn);
  const [sortDirection, setSortDirection] = useState<SortDirection>(initialSortDirection);

  useEffect(() => {
    if (isOpen) {
      setFilter('');
      setTypeFilter(initialTypeFilter);
      setStatusFilter(initialStatusFilter);
      setInternalFilter(initialInternalFilter);
      setSortColumn(initialSortColumn);
      setSortDirection(initialSortDirection);
    }
  }, [isOpen, initialStatusFilter, initialInternalFilter, initialTypeFilter, initialSortColumn, initialSortDirection]);

  const handleSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const uniqueTypes = useMemo(() => {
    const types = new Set(requests.map(r => r.type));
    return Array.from(types).sort();
  }, [requests]);

  const filteredRequests = requests.filter((req) => {
    const matchesText = filter === '' || req.url.toLowerCase().includes(filter.toLowerCase());
    const matchesType = typeFilter === 'all' || req.type === typeFilter;
    const matchesStatus = statusFilter === 'all' ||
      (statusFilter === 'blocked' && req.blocked) ||
      (statusFilter === 'failed' && req.failed);
    const matchesInternal = internalFilter === 'all' ||
      (internalFilter === 'internal' && req.isInternal) ||
      (internalFilter === 'external' && !req.isInternal);
    return matchesText && matchesType && matchesStatus && matchesInternal;
  });

  const sortedRequests = sortColumn ? [...filteredRequests].sort((a, b) => {
    const dir = sortDirection === 'asc' ? 1 : -1;
    switch (sortColumn) {
      case 'url':
        return dir * a.url.localeCompare(b.url);
      case 'status': {
        const aStatus = a.blocked ? -2 : a.failed ? -1 : a.status;
        const bStatus = b.blocked ? -2 : b.failed ? -1 : b.status;
        return dir * (aStatus - bStatus);
      }
      case 'type':
        return dir * a.type.localeCompare(b.type);
      case 'size': {
        const aSize = a.size ?? -1;
        const bSize = b.size ?? -1;
        return dir * (aSize - bSize);
      }
      case 'time': {
        const aTime = a.time ?? -1;
        const bTime = b.time ?? -1;
        return dir * (aTime - bTime);
      }
      default:
        return 0;
    }
  }) : filteredRequests;

  const renderSortIcon = (column: SortColumn) => {
    const isActive = sortColumn === column;
    const iconName = isActive && sortDirection === 'desc' ? 'chevron-down' : 'chevron-up';
    return (
      <span className={isActive ? styles.sortIcon : styles.sortIconHidden}>
        <Icon name={iconName} size={16} />
      </span>
    );
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Network Requests (${filteredRequests.length} of ${totalRequests})`}
      size="xl"
      searchInputRef={searchInputRef}
    >
      <div className={filterStyles.modalFilters}>
        <FilterInput
          ref={searchInputRef}
          placeholder="Filter requests..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <select
          className={filterStyles.filterSelect}
          value={internalFilter}
          onChange={(e) => setInternalFilter(e.target.value as NetworkInternalFilter)}
        >
          <option value="all">All Sources</option>
          <option value="internal">Internal</option>
          <option value="external">External</option>
        </select>
        <select
          className={filterStyles.filterSelect}
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as NetworkStatusFilter)}
        >
          <option value="all">All Status</option>
          <option value="blocked">Blocked</option>
          <option value="failed">Failed</option>
        </select>
        <select
          className={filterStyles.filterSelect}
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as ResourceType | 'all')}
        >
          <option value="all">All Types</option>
          {uniqueTypes.map(type => (
            <option key={type} value={type}>
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.modalBody}>
        <table className={styles.networkTable}>
          <thead>
            <tr>
              <th className={`${styles.sortableHeader} ${sortColumn === 'url' ? styles.sortableHeaderActive : ''}`} onClick={() => handleSort('url')}>
                <span className={styles.headerContent}>URL {renderSortIcon('url')}</span>
              </th>
              <th className={`${styles.sortableHeader} ${sortColumn === 'status' ? styles.sortableHeaderActive : ''}`} onClick={() => handleSort('status')}>
                <span className={styles.headerContent}>Status {renderSortIcon('status')}</span>
              </th>
              <th className={`${styles.sortableHeader} ${sortColumn === 'type' ? styles.sortableHeaderActive : ''}`} onClick={() => handleSort('type')}>
                <span className={styles.headerContent}>Type {renderSortIcon('type')}</span>
              </th>
              <th className={`${styles.sortableHeader} ${sortColumn === 'size' ? styles.sortableHeaderActive : ''}`} onClick={() => handleSort('size')}>
                <span className={styles.headerContent}>Size {renderSortIcon('size')}</span>
              </th>
              <th className={`${styles.sortableHeader} ${sortColumn === 'time' ? styles.sortableHeaderActive : ''}`} onClick={() => handleSort('time')}>
                <span className={styles.headerContent}>Time {renderSortIcon('time')}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {sortedRequests.map((req) => {
              const shouldTruncate = req.url.length > 200 || req.url.startsWith('data:');
              return (
              <tr
                key={req.id}
                className={`${styles.requestRow} ${req.blocked ? styles.requestBlocked : req.failed ? styles.requestFailed : ''}`}
              >
                <td
                  className={`${styles.requestUrl} ${shouldTruncate ? styles.requestUrlTruncated : ''}`}
                  title={shouldTruncate ? req.url : undefined}
                >
                  {req.url}
                </td>
                <td>
                  <StatusBadge status={req.blocked ? 'blocked' : req.failed ? 'failed' : String(req.status) as '200' | '301' | '302' | '404' | '500'}>
                    {req.blocked ? 'Blocked' : req.failed ? 'Failed' : req.status}
                  </StatusBadge>
                </td>
                <td>
                  <TypeBadge type={req.type}>{req.type}</TypeBadge>
                </td>
                <td><PageSize bytes={req.size} /></td>
                <td><LoadTime seconds={req.time} /></td>
              </tr>
              );
            })}
            {sortedRequests.length === 0 && (
              <tr>
                <td colSpan={5} className={styles.emptyMessage}>
                  No requests match your filters
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </Modal>
  );
}
