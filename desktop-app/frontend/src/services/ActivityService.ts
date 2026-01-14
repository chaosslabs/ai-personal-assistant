import { Activity, ActivityType, RecordingSession, TranscriptChunk } from '../types/activity';
import { APIError } from '../types/api';
import {
  GetActivities,
  CreateActivity,
  GetActivity,
  StartActivity,
  StopActivity,
  DeleteActivity,
  GetActivityTranscript,
  ProcessActivityTranscription,
  UpdateActivityTitle,
  UpdateActivityType
} from '../../wailsjs/go/main/App';

export class ActivityService {
  // Get all activities for the current user
  async getActivities(): Promise<Activity[]> {
    try {
      const activities = await GetActivities();
      return activities || [];
    } catch (error) {
      console.error('Failed to fetch activities:', error);
      throw new APIError('Failed to fetch activities', error);
    }
  }

  // Create a new activity
  async createActivity(type: ActivityType, title: string): Promise<Activity> {
    try {
      const activity = await CreateActivity(type, title);
      return activity;
    } catch (error) {
      console.error('Failed to create activity:', error);
      throw new APIError('Failed to create activity', error);
    }
  }

  // Get a specific activity
  async getActivity(activityId: string): Promise<Activity> {
    try {
      const activity = await GetActivity(activityId);
      return activity;
    } catch (error) {
      console.error('Failed to get activity:', error);
      throw new APIError('Failed to get activity', error);
    }
  }

  // Start an activity
  async startActivity(activityId: string): Promise<void> {
    try {
      await StartActivity(activityId);
    } catch (error) {
      console.error('Failed to start activity:', error);
      throw new APIError('Failed to start activity', error);
    }
  }

  // Stop an activity
  async stopActivity(activityId: string): Promise<void> {
    try {
      await StopActivity(activityId);
    } catch (error) {
      console.error('Failed to stop activity:', error);
      throw new APIError('Failed to stop activity', error);
    }
  }

  // Delete an activity
  async deleteActivity(activityId: string): Promise<void> {
    console.log(`🔧 ActivityService: Starting deleteActivity for ID: ${activityId}`);
    try {
      console.log('🔧 ActivityService: Calling backend DeleteActivity...');
      await DeleteActivity(activityId);
      console.log('🔧 ActivityService: Backend DeleteActivity completed successfully');
    } catch (error) {
      console.error('🔧 ActivityService: Backend DeleteActivity failed:', error);
      console.error('🔧 ActivityService: Error details:', error.message, error.stack);
      throw new APIError('Failed to delete activity', error);
    }
  }

  // Update activity title
  async updateActivityTitle(activityId: string, newTitle: string): Promise<void> {
    try {
      await UpdateActivityTitle(activityId, newTitle);
    } catch (error) {
      console.error('Failed to update activity title:', error);
      throw new APIError('Failed to update activity title', error);
    }
  }

  // Update activity type
  async updateActivityType(activityId: string, newType: ActivityType): Promise<void> {
    try {
      await UpdateActivityType(activityId, newType);
    } catch (error) {
      console.error('Failed to update activity type:', error);
      throw new APIError('Failed to update activity type', error);
    }
  }

  // Get transcript for an activity
  async getActivityTranscript(activityId: string): Promise<TranscriptChunk[]> {
    try {
      const chunks = await GetActivityTranscript(activityId);
      return chunks || [];
    } catch (error) {
      console.error('Failed to get activity transcript:', error);
      throw new APIError('Failed to get activity transcript', error);
    }
  }

  // Process transcription for an activity
  async processTranscription(activityId: string): Promise<void> {
    try {
      await ProcessActivityTranscription(activityId);
    } catch (error) {
      console.error('Failed to process transcription:', error);
      throw new APIError('Failed to process transcription', error);
    }
  }

  // Transform Go activity data to UI format
  transformActivityForUI(activity: Activity): any {
    return {
      id: activity.id,
      title: activity.title,
      type: activity.type,
      status: this.mapActivityStatus(activity.status),
      duration: this.calculateDuration(activity),
      startTime: activity.start_time,
      transcriptAvailable: false, // Will be set separately
      summary: this.getActivitySummary(activity)
    };
  }

  private mapActivityStatus(status: string): 'active' | 'completed' | 'scheduled' {
    switch (status) {
      case 'active':
      case 'recording':
        return 'active';
      case 'completed':
        return 'completed';
      default:
        return 'scheduled';
    }
  }

  private calculateDuration(activity: Activity): string {
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

  private getActivitySummary(activity: Activity): string {
    if (activity.metadata?.summary) {
      return activity.metadata.summary;
    }

    const typeLabel = activity.type.replace('_', ' ');
    return `${typeLabel} session`;
  }
}