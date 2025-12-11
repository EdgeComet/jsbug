import { describe, it, expect } from 'vitest';
import { validateHttpUrl, isValidHttpUrl } from './urlValidation';

describe('validateHttpUrl', () => {
  describe('empty and whitespace', () => {
    it('should reject empty string', () => {
      const result = validateHttpUrl('');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL is required');
    });

    it('should reject whitespace-only string', () => {
      const result = validateHttpUrl('   ');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL is required');
    });
  });

  describe('valid URLs', () => {
    it('should accept https URLs', () => {
      expect(validateHttpUrl('https://example.com').valid).toBe(true);
    });

    it('should accept http URLs', () => {
      expect(validateHttpUrl('http://example.com').valid).toBe(true);
    });

    it('should accept URLs with paths', () => {
      expect(validateHttpUrl('https://example.com/path/to/page').valid).toBe(true);
    });

    it('should accept URLs with query strings', () => {
      expect(validateHttpUrl('https://example.com?foo=bar&baz=qux').valid).toBe(true);
    });

    it('should accept URLs with ports', () => {
      expect(validateHttpUrl('https://example.com:8080').valid).toBe(true);
    });
  });

  describe('invalid protocols', () => {
    it('should reject javascript: URLs', () => {
      const result = validateHttpUrl('javascript:alert(1)');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL must use http or https');
    });

    it('should reject file: URLs', () => {
      const result = validateHttpUrl('file:///etc/passwd');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL must use http or https');
    });

    it('should reject data: URLs', () => {
      const result = validateHttpUrl('data:text/html,<script>alert(1)</script>');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL must use http or https');
    });

    it('should reject ftp: URLs', () => {
      const result = validateHttpUrl('ftp://example.com');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('URL must use http or https');
    });
  });

  describe('local URLs', () => {
    it('should reject localhost', () => {
      const result = validateHttpUrl('http://localhost');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Local URLs are not allowed');
    });

    it('should reject localhost with port', () => {
      const result = validateHttpUrl('http://localhost:3000');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Local URLs are not allowed');
    });

    it('should reject 127.0.0.1', () => {
      const result = validateHttpUrl('http://127.0.0.1');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Local URLs are not allowed');
    });

    it('should reject 192.168.x.x addresses', () => {
      const result = validateHttpUrl('http://192.168.1.1');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Local URLs are not allowed');
    });

    it('should reject ::1 (IPv6 localhost)', () => {
      const result = validateHttpUrl('http://[::1]');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Local URLs are not allowed');
    });
  });

  describe('malformed URLs', () => {
    it('should reject relative paths', () => {
      const result = validateHttpUrl('foo/bar');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please enter a valid URL');
    });

    it('should reject URLs without protocol', () => {
      const result = validateHttpUrl('example.com');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please enter a valid URL');
    });

    it('should reject malformed URLs', () => {
      const result = validateHttpUrl('http://');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please enter a valid URL');
    });
  });
});

describe('isValidHttpUrl', () => {
  it('should return true for valid URLs', () => {
    expect(isValidHttpUrl('https://example.com')).toBe(true);
  });

  it('should return false for invalid URLs', () => {
    expect(isValidHttpUrl('')).toBe(false);
    expect(isValidHttpUrl('javascript:alert(1)')).toBe(false);
    expect(isValidHttpUrl('http://localhost')).toBe(false);
    expect(isValidHttpUrl('foo/bar')).toBe(false);
  });
});
