import { useState, useEffect } from 'react';
import { useConfig } from '../../context/ConfigContext';
import type { AppConfig, PanelConfig } from '../../types/config';
import { ConfigColumn } from './ConfigColumn';
import { Button } from '../common/Button';
import { Modal } from '../common/Modal';
import styles from './ConfigModal.module.css';

interface ConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onApply?: (newConfig: AppConfig) => void;
}

const isCustomUaInvalid = (config: PanelConfig) =>
  config.userAgent === 'custom' && !config.customUserAgent?.trim();

export function ConfigModal({ isOpen, onClose, onApply }: ConfigModalProps) {
  const { config, setConfig } = useConfig();
  const [draftConfig, setDraftConfig] = useState<AppConfig>(config);

  useEffect(() => {
    if (isOpen) {
      setDraftConfig(config);
    }
  }, [isOpen, config]);

  const leftCustomUaError = isCustomUaInvalid(draftConfig.left);
  const rightCustomUaError = isCustomUaInvalid(draftConfig.right);
  const isValid = !leftCustomUaError && !rightCustomUaError;

  const handleApply = () => {
    setConfig(draftConfig);
    onClose();
    onApply?.(draftConfig);
  };

  const handleCancel = () => {
    onClose();
  };

  const footer = (
    <>
      <Button variant="ghost" onClick={handleCancel}>
        Cancel
      </Button>
      <Button variant="primary" onClick={handleApply} disabled={!isValid}>
        Apply
      </Button>
    </>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Render Configuration"
      size="wide"
      footer={footer}
    >
      <div className={styles.modalContent}>
        <div className={styles.configColumns}>
          <ConfigColumn
            side="left"
            config={draftConfig.left}
            onChange={(left) => setDraftConfig({ ...draftConfig, left })}
            customUaError={leftCustomUaError}
          />

          <div className={styles.configDivider}></div>

          <ConfigColumn
            side="right"
            config={draftConfig.right}
            onChange={(right) => setDraftConfig({ ...draftConfig, right })}
            customUaError={rightCustomUaError}
          />
        </div>
      </div>
    </Modal>
  );
}
