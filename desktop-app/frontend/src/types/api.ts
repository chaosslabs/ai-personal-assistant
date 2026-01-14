// API response types and error handling

export interface APIResponse<T> {
  data?: T;
  error?: string;
  success: boolean;
}

export class APIError extends Error {
  constructor(message: string, public originalError?: any) {
    super(message);
    this.name = 'APIError';
  }
}

export interface LoadingState {
  isLoading: boolean;
  error?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  has_more: boolean;
}

// Wails event types
export interface WailsEvent<T = any> {
  name: string;
  data: T;
}

export type EventHandler<T = any> = (data: T) => void;

// Common events emitted by the backend
export interface WailsEvents {
  'activity:started': string; // activity ID
  'activity:stopped': string; // activity ID
  'activity:deleted': string; // activity ID
  'recording:started': any; // AudioRecording object
  'recording:stopped': string; // recording ID
  'transcription:started': string; // activity ID
  'transcription:recording_started': { activity_id: string; recording_id: string };
  'model:download_started': string; // model ID
  'model:activated': string; // model ID
}