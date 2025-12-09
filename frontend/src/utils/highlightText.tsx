import { ReactNode } from 'react';

export function highlightText(
  text: string,
  searchTerm: string,
  highlightClass: string
): ReactNode {
  if (!searchTerm.trim()) return text;

  const regex = new RegExp(
    `(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`,
    'gi'
  );
  const parts = text.split(regex);

  return parts.map((part, index) => {
    if (part.toLowerCase() === searchTerm.toLowerCase()) {
      return (
        <mark key={index} className={highlightClass}>
          {part}
        </mark>
      );
    }
    return part;
  });
}
