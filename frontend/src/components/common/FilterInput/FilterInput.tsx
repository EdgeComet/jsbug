import { InputHTMLAttributes, forwardRef } from 'react';
import styles from './FilterInput.module.css';

type FilterInputProps = InputHTMLAttributes<HTMLInputElement>;

export const FilterInput = forwardRef<HTMLInputElement, FilterInputProps>(
  ({ className, ...props }, ref) => {
    return (
      <input
        ref={ref}
        type="text"
        className={`${styles.filterInput} ${className || ''}`}
        {...props}
      />
    );
  }
);

FilterInput.displayName = 'FilterInput';
