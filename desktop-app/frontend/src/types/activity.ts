// TypeScript interfaces matching the Go backend models

export type ActivityType = 'meeting' | 'work_session' | 'call' | 'other';
export type ActivityStatus = 'active' | 'recording' | 'processing' | 'completed' | 'failed';

export interface Activity {
  id: string;
  user_id: string;
  type: ActivityType;
  title: string;
  start_time: string; // ISO string
  end_time?: string; // ISO string
  status: ActivityStatus;
  tags: string[];
  metadata: Record<string, any>;
  created_at: string; // ISO string
  updated_at: string; // ISO string
  deleted_at?: string; // ISO string
}

export interface AudioDeviceInfo {
  device_id: string;
  name: string;
  sample_rate: number;
  channels: number;
  device_type: string;
}

export interface AudioRecording {
  id: string;
  activity_id: string;
  user_id: string;
  device_info: AudioDeviceInfo;
  file_path: string;
  duration_seconds: number;
  sample_rate: number;
  channels: number;
  start_time: string; // ISO string
  end_time?: string; // ISO string
  file_size_bytes: number;
  status: string;
  metadata: Record<string, any>;
  created_at: string; // ISO string
  updated_at: string; // ISO string
}

export interface TranscriptChunk {
  id: string;
  audio_recording_id: string;
  activity_id: string;
  user_id: string;
  start_time: number;  // Seconds from activity start
  end_time: number;    // Seconds from activity start
  text: string;
  confidence?: number; // Optional 0-1 confidence score
  speaker?: string;    // Optional speaker identification
  language?: string;   // Optional language code
  created_at: string;  // ISO string
}

export interface RecordingSession {
  activity?: Activity;
  audio_recording?: AudioRecording;
  file_path: string;
}

// Helper functions to convert Go data to UI data
export function calculateDuration(activity: Activity): string {
  if (!activity.start_time) return '0 min';

  const start = new Date(activity.start_time);
  const end = activity.end_time ? new Date(activity.end_time) : new Date();
  const diffMs = end.getTime() - start.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));

  if (diffMinutes < 60) {
    return `${diffMinutes} min`;
  } else {
    const hours = Math.floor(diffMinutes / 60);
    const minutes = diffMinutes % 60;
    return `${hours}h ${minutes}min`;
  }
}

export function hasTranscript(activityId: string, transcripts: TranscriptChunk[]): boolean {
  return transcripts.some(chunk => chunk.activity_id === activityId);
}

export function getActivitySummary(activity: Activity, transcripts: TranscriptChunk[]): string {
  // For now, use metadata summary or generate from title
  if (activity.metadata?.summary) {
    return activity.metadata.summary;
  }

  // Generate default summary based on type and transcripts
  const transcriptAvailable = hasTranscript(activity.id, transcripts);
  const typeLabel = activity.type.replace('_', ' ');

  if (transcriptAvailable) {
    return `${typeLabel} session with transcript available`;
  } else {
    return `${typeLabel} session`;
  }
}