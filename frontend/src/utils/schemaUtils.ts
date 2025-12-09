/**
 * Utilities for parsing and displaying structured data (JSON-LD/schema.org)
 */

function extractTypesFromItem(item: unknown, types: Map<string, number>): boolean {
  if (!item || typeof item !== 'object') return false;

  // Handle arrays (nested schema items)
  if (Array.isArray(item)) {
    let foundAny = false;
    for (const arrayItem of item) {
      if (extractTypesFromItem(arrayItem, types)) {
        foundAny = true;
      }
    }
    return foundAny;
  }

  const obj = item as Record<string, unknown>;

  // Handle @graph array (nested JSON-LD)
  if ('@graph' in obj && Array.isArray(obj['@graph'])) {
    let foundAny = false;
    for (const graphItem of obj['@graph']) {
      if (extractTypesFromItem(graphItem, types)) {
        foundAny = true;
      }
    }
    return foundAny;
  }

  // Handle direct @type
  if ('@type' in obj) {
    const typeValue = obj['@type'];
    const typeNames = Array.isArray(typeValue) ? typeValue : [typeValue];
    let foundAny = false;
    for (const t of typeNames) {
      if (typeof t === 'string' && t.trim()) {
        types.set(t, (types.get(t) || 0) + 1);
        foundAny = true;
      }
    }
    return foundAny;
  }

  return false;
}

export interface SchemaTypeCounts {
  types: Map<string, number>;
  invalid: number;
}

export function getSchemaTypeCounts(data: unknown[]): SchemaTypeCounts {
  const types = new Map<string, number>();
  let invalid = 0;

  for (const item of data) {
    try {
      if (!extractTypesFromItem(item, types)) {
        invalid++;
      }
    } catch {
      invalid++;
    }
  }
  return { types, invalid };
}

export function formatSchemaTypes(result: SchemaTypeCounts): string {
  const parts: string[] = [];

  for (const [type, count] of result.types) {
    parts.push(count > 1 ? `${type} (${count})` : type);
  }

  if (result.invalid > 0) {
    parts.push(`invalid (${result.invalid})`);
  }

  return parts.length > 0 ? parts.join(', ') : 'No valid schemas';
}

export function schemaTypesEqual(a: SchemaTypeCounts, b: SchemaTypeCounts): boolean {
  if (a.invalid !== b.invalid) return false;
  if (a.types.size !== b.types.size) return false;
  for (const [type, count] of a.types) {
    if (b.types.get(type) !== count) return false;
  }
  return true;
}

export interface SchemaTypeDisplay {
  type: string;
  count: number;
  status: 'same' | 'added' | 'removed';
}

export function getSchemaTypesList(
  data: SchemaTypeCounts,
  compareData?: SchemaTypeCounts
): SchemaTypeDisplay[] {
  const result: SchemaTypeDisplay[] = [];
  const compareTypes = compareData?.types ?? new Map();

  // Current types
  for (const [type, count] of data.types) {
    const isAdded = compareData && !compareTypes.has(type);
    result.push({
      type,
      count,
      status: isAdded ? 'added' : 'same',
    });
  }

  // Removed types (in compare but not in current)
  if (compareData) {
    for (const [type, count] of compareData.types) {
      if (!data.types.has(type)) {
        result.push({
          type,
          count,
          status: 'removed',
        });
      }
    }
  }

  return result;
}
