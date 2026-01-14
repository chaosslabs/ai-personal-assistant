import { RecordingSession, AudioRecording, Activity } from '../types/activity';
import { APIError } from '../types/api';
import {
  StartRecordingButtonAction,
  StopRecordingButtonAction,
  CreateRecordingWithMode
} from '../../wailsjs/go/main/App';

// Recording state management
interface RecordingState {
  isRecording: boolean;
  currentSession: RecordingSession | null;
  currentActivity: Activity | null;
  startTime: Date | null;
  error: string | null;
}

export class RecordingService {
  private state: RecordingState = {
    isRecording: false,
    currentSession: null,
    currentActivity: null,
    startTime: null,
    error: null
  };

  private listeners: ((state: RecordingState) => void)[] = [];

  // State management
  subscribe(listener: (state: RecordingState) => void): () => void {
    this.listeners.push(listener);
    return () => {
      const index = this.listeners.indexOf(listener);
      if (index > -1) {
        this.listeners.splice(index, 1);
      }
    };
  }

  private setState(updates: Partial<RecordingState>) {
    this.state = { ...this.state, ...updates };
    this.listeners.forEach(listener => listener(this.state));
  }

  getState(): RecordingState {
    return { ...this.state };
  }

  // Start a simple recording (creates activity automatically)
  async startRecording(): Promise<RecordingSession> {
    try {
      // Prevent multiple simultaneous recordings
      if (this.state.isRecording) {
        throw new Error('Recording already in progress');
      }

      this.setState({
        isRecording: true,
        startTime: new Date(),
        error: null
      });

      const session = await StartRecordingButtonAction();

      this.setState({
        currentSession: session,
        currentActivity: session.activity || null,
        isRecording: true,
        startTime: new Date(),
        error: null
      });

      return session;
    } catch (error) {
      console.error('Failed to start recording:', error);
      this.setState({
        isRecording: false,
        currentSession: null,
        currentActivity: null,
        startTime: null,
        error: error instanceof Error ? error.message : 'Failed to start recording'
      });
      throw new APIError('Failed to start recording', error);
    }
  }

  // Stop the current recording
  async stopRecording(): Promise<void> {
    try {
      if (!this.state.currentSession) {
        throw new Error('No active recording to stop');
      }

      await StopRecordingButtonAction(this.state.currentSession.audio_recording?.id!);

      this.setState({
        isRecording: false,
        currentSession: null,
        currentActivity: null,
        startTime: null,
        error: null
      });
    } catch (error) {
      console.error('Failed to stop recording:', error);
      this.setState({
        error: error instanceof Error ? error.message : 'Failed to stop recording'
      });
      throw new APIError('Failed to stop recording', error);
    }
  }

  // Stop a specific recording by ID (for backward compatibility)
  async stopRecordingById(recordingId: string): Promise<void> {
    try {
      await StopRecordingButtonAction(recordingId);

      // Update state if this was the current recording
      if (this.state.currentSession?.audio_recording?.id === recordingId) {
        this.setState({
          isRecording: false,
          currentSession: null,
          currentActivity: null,
          startTime: null,
          error: null
        });
      }
    } catch (error) {
      console.error('Failed to stop recording:', error);
      throw new APIError('Failed to stop recording', error);
    }
  }

  // Create recording with specific mode for an activity
  async createRecordingWithMode(activityId: string, recordingMode: string): Promise<RecordingSession> {
    try {
      const session = await CreateRecordingWithMode(activityId, recordingMode);
      return session;
    } catch (error) {
      console.error('Failed to create recording with mode:', error);
      throw new APIError('Failed to create recording', error);
    }
  }

  // Get current recording duration in seconds
  getCurrentDuration(): number {
    if (!this.state.isRecording || !this.state.startTime) {
      return 0;
    }
    return Math.floor((Date.now() - this.state.startTime.getTime()) / 1000);
  }

  // Clear any error state
  clearError(): void {
    this.setState({ error: null });
  }

  // Update the current activity title in state
  updateCurrentActivityTitle(newTitle: string): void {
    if (this.state.currentActivity) {
      this.setState({
        currentActivity: { ...this.state.currentActivity, title: newTitle }
      });
    }
  }

  // Update the current activity type in state
  updateCurrentActivityType(newType: string): void {
    if (this.state.currentActivity) {
      this.setState({
        currentActivity: { ...this.state.currentActivity, type: newType }
      });
    }
  }

  // Get current activity
  getCurrentActivity(): Activity | null {
    return this.state.currentActivity;
  }

  // Format recording duration for display
  formatDuration(durationSeconds: number): string {
    if (durationSeconds < 60) {
      return `${Math.floor(durationSeconds)}s`;
    }

    const minutes = Math.floor(durationSeconds / 60);
    const seconds = Math.floor(durationSeconds % 60);

    if (minutes < 60) {
      return `${minutes}m ${seconds}s`;
    }

    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    return `${hours}h ${remainingMinutes}m`;
  }

  // Get file size in human-readable format
  formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  }
}

// Export singleton instance for shared state across components
export const recordingService = new RecordingService();