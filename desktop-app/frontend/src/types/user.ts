// User and system types matching Go backend

export interface User {
  id: string;
  username: string;
  settings: UserSettings;
  created_at: string; // ISO string
}

export interface UserSettings {
  theme?: 'light' | 'dark' | 'system';
  audio_device_id?: string;
  recording_quality?: 'low' | 'medium' | 'high';
  auto_transcription?: boolean;
  privacy_mode?: boolean;
  [key: string]: any; // Allow additional settings
}

export interface AppStatusDetailed {
  app_name: string;
  version: string;
  database: {
    connected: boolean;
    path?: string;
  };
  user?: {
    id: string;
    username: string;
  };
}

export interface SystemInfo {
  data_directory: string;
  database_file: string;
  database_path?: string;
  database_connected: boolean;
  activities_dir?: string;
  models_dir?: string;
  disk_usage_bytes?: number;
  disk_usage_mb?: number;
}

export interface AudioDevice {
  id: string;
  name: string;
  type: 'input' | 'output';
  sample_rate: number;
  channels: number;
  device_type?: string;
}

export interface RecordingMode {
  id: string;
  name: string;
  description: string;
  icon: string;
  recommended?: boolean;
}