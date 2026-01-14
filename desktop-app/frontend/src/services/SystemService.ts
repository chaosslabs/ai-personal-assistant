import { User, AppStatusDetailed, SystemInfo, AudioDevice, RecordingMode } from '../types/user';
import { APIError } from '../types/api';
import {
  GetVersion,
  GetCurrentUser,
  GetAppStatus,
  GetAppStatusDetailed,
  GetSystemInfo,
  GetAudioDevices,
  GetRecordingModes,
  GetUserSettings,
  UpdateUserSettings,
  GetActivityTranscript,
  SearchActivityTranscripts,
  GetTranscriptionStatus,
  ProcessActivityTranscription
} from '../../wailsjs/go/main/App';

export class SystemService {
  // Get app version
  async getVersion(): Promise<string> {
    try {
      return await GetVersion();
    } catch (error) {
      console.error('Failed to get version:', error);
      return 'Unknown';
    }
  }

  // Get current user
  async getCurrentUser(): Promise<User | null> {
    try {
      const userData = await GetCurrentUser();
      if (userData.error) {
        return null;
      }

      return {
        id: userData.id,
        username: userData.username,
        settings: userData.settings || {},
        created_at: userData.created_at
      };
    } catch (error) {
      console.error('Failed to get current user:', error);
      return null;
    }
  }

  // Get basic app status (string format)
  async getAppStatus(): Promise<string> {
    try {
      return await GetAppStatus();
    } catch (error) {
      console.error('Failed to get app status:', error);
      return 'Status unavailable';
    }
  }

  // Get detailed app status
  async getAppStatusDetailed(): Promise<AppStatusDetailed> {
    try {
      return await GetAppStatusDetailed();
    } catch (error) {
      console.error('Failed to get detailed app status:', error);
      throw new APIError('Failed to get app status', error);
    }
  }

  // Get system information
  async getSystemInfo(): Promise<SystemInfo> {
    try {
      return await GetSystemInfo();
    } catch (error) {
      console.error('Failed to get system info:', error);
      throw new APIError('Failed to get system info', error);
    }
  }

  // Get available audio devices
  async getAudioDevices(): Promise<AudioDevice[]> {
    try {
      const devices = await GetAudioDevices();
      return devices.map(device => ({
        id: device.id,
        name: device.name,
        type: device.type || 'input',
        sample_rate: device.sample_rate || 44100,
        channels: device.channels || 2,
        device_type: device.device_type
      }));
    } catch (error) {
      console.error('Failed to get audio devices:', error);
      // Return fallback device
      return [{
        id: '0',
        name: 'Default Microphone',
        type: 'input',
        sample_rate: 44100,
        channels: 2
      }];
    }
  }

  // Get available recording modes
  async getRecordingModes(): Promise<RecordingMode[]> {
    try {
      return await GetRecordingModes();
    } catch (error) {
      console.error('Failed to get recording modes:', error);
      // Return fallback modes
      return [
        {
          id: 'microphone',
          name: 'Microphone Only',
          description: 'Record only your microphone input',
          icon: 'microphone'
        }
      ];
    }
  }

  // Calculate storage usage in human-readable format
  formatStorageUsage(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  }

  // Get current time formatted for status bar
  getCurrentTime(): string {
    return new Date().toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  // Get user settings
  async getUserSettings(): Promise<any> {
    try {
      return await GetUserSettings();
    } catch (error) {
      console.error('Failed to get user settings:', error);
      throw new APIError('Failed to get user settings', error);
    }
  }

  // Update user settings
  async updateUserSettings(settings: any): Promise<void> {
    try {
      await UpdateUserSettings(settings);
    } catch (error) {
      console.error('Failed to update user settings:', error);
      throw new APIError('Failed to update user settings', error);
    }
  }

  // Get activity transcript
  async getActivityTranscript(activityId: string): Promise<any[]> {
    try {
      return await GetActivityTranscript(activityId);
    } catch (error) {
      console.error('Failed to get activity transcript:', error);
      throw new APIError('Failed to get activity transcript', error);
    }
  }

  // Search transcripts
  async searchTranscripts(query: string): Promise<any[]> {
    try {
      return await SearchActivityTranscripts(query);
    } catch (error) {
      console.error('Failed to search transcripts:', error);
      throw new APIError('Failed to search transcripts', error);
    }
  }

  // Get transcription status
  async getTranscriptionStatus(activityId: string): Promise<any> {
    try {
      return await GetTranscriptionStatus(activityId);
    } catch (error) {
      console.error('Failed to get transcription status:', error);
      throw new APIError('Failed to get transcription status', error);
    }
  }

  // Start transcription
  async startTranscription(activityId: string): Promise<void> {
    try {
      await ProcessActivityTranscription(activityId);
    } catch (error) {
      console.error('Failed to start transcription:', error);
      throw new APIError('Failed to start transcription', error);
    }
  }
}