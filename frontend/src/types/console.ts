export type ConsoleLevel = 'log' | 'warn' | 'error';

export interface ConsoleEntry {
  id: string;
  level: ConsoleLevel;
  message: string;
  time: number; // seconds
}
