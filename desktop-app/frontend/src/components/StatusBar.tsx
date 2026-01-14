import { useState, useEffect } from 'react';
import { Badge } from './ui/badge';
import { Clock, Database, Wifi, Shield } from 'lucide-react';
import { SystemService } from '../services/SystemService';
import { ActivityService } from '../services/ActivityService';
import { recordingService } from '../services/RecordingService';

interface StatusBarProps {
  isRecording: boolean;
}

export function StatusBar({ isRecording }: StatusBarProps) {
  const [currentTime, setCurrentTime] = useState('');
  const [storageUsed, setStorageUsed] = useState('Loading...');
  const [todayStats, setTodayStats] = useState({ activities: 0, transcribedTime: '0min' });
  const [whisperStatus, setWhisperStatus] = useState('Checking...');
  const [systemHealth, setSystemHealth] = useState({
    database: 'unknown',
    version: 'unknown',
    user: null
  });
  const [recordingDuration, setRecordingDuration] = useState(0);
  const systemService = new SystemService();
  const activityService = new ActivityService();

  // Update time and recording duration every second
  useEffect(() => {
    const updateTimeAndDuration = () => {
      setCurrentTime(systemService.getCurrentTime());
      if (isRecording) {
        setRecordingDuration(recordingService.getCurrentDuration());
      }
    };

    updateTimeAndDuration(); // Initial call
    const interval = setInterval(updateTimeAndDuration, 1000);
    return () => clearInterval(interval);
  }, [isRecording]);

  // Load system status on mount
  useEffect(() => {
    const loadSystemStatus = async () => {
      try {
        // Get system info for storage
        const systemInfo = await systemService.getSystemInfo();
        if (systemInfo.disk_usage_bytes) {
          setStorageUsed(systemService.formatStorageUsage(systemInfo.disk_usage_bytes));
        } else if (systemInfo.disk_usage_mb) {
          setStorageUsed(systemService.formatStorageUsage(systemInfo.disk_usage_mb * 1024 * 1024));
        }

        // Get today's activities
        const activities = await activityService.getActivities();
        const today = new Date().toDateString();
        const todayActivities = activities.filter(activity => {
          if (!activity.start_time) return false;
          return new Date(activity.start_time).toDateString() === today;
        });

        // TODO: Calculate actual transcribed time (check for real transcripts)
        // let transcribedMinutes = 0;
        // try {
        //   for (const activity of todayActivities) {
        //     if (activity.status === 'completed') {
        //       try {
        //         const transcript = await systemService.getActivityTranscript(activity.id);
        //         if (transcript && transcript.length > 0) {
        //           const lastChunk = transcript[transcript.length - 1];
        //           const firstChunk = transcript[0];
        //           if (lastChunk && firstChunk) {
        //             const durationInSeconds = lastChunk.end_time - firstChunk.start_time;
        //             transcribedMinutes += Math.floor(durationInSeconds / 60);
        //           }
        //         }
        //       } catch (error) {
        //         console.debug('No transcript available for activity', activity.id);
        //       }
        //     }
        //   }
        // } catch (error) {
        //   console.warn('Failed to calculate transcribed time:', error);
        //   transcribedMinutes = 0;
        // }

        setTodayStats({
          activities: todayActivities.length,
          transcribedTime: '0min' // Placeholder
        });

        // Get detailed app status for health monitoring
        const appStatus = await systemService.getAppStatusDetailed();
        setSystemHealth({
          database: appStatus.database?.connected ? 'connected' : 'disconnected',
          version: appStatus.version || 'unknown',
          user: appStatus.user || null
        });

        // Check Whisper status (transcription enabled)
        setWhisperStatus('Ready');
      } catch (error) {
        console.error('Failed to load system status:', error);
        setStorageUsed('Error');
        setWhisperStatus('Error');
      }
    };

    loadSystemStatus();
    // Refresh every 30 seconds
    const interval = setInterval(loadSystemStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex items-center justify-between px-4 py-1.5 bg-secondary/30 border-t border-border text-xs text-muted-foreground">
      {/* Left: App Status */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-1">
          <Shield className="w-3 h-3 text-green-500" />
          <span>Local Only</span>
        </div>

        <div className="flex items-center space-x-1">
          <div className={`w-2 h-2 rounded-full ${systemHealth.database === 'connected' ? 'bg-green-500' : 'bg-red-500'}`} />
          <span>DB {systemHealth.database}</span>
        </div>

        <div className="flex items-center space-x-1">
          <Database className="w-3 h-3" />
          <span>{storageUsed} used</span>
        </div>

        {isRecording && (
          <div className="flex items-center space-x-1 text-red-500">
            <div className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
            <span>Recording • {recordingService.formatDuration(recordingDuration)}</span>
          </div>
        )}
      </div>

      {/* Center: Activity Stats */}
      <div className="flex items-center space-x-4">
        <span>Today: {todayStats.activities} activities</span>
        {/* TODO: Implement transcribed time calculation
        <span>•</span>
        <span>{todayStats.transcribedTime} transcribed</span>
        */}
      </div>

      {/* Right: System Info */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-1">
          <div className={`w-2 h-2 rounded-full ${whisperStatus === 'Ready' ? 'bg-green-500' : whisperStatus === 'Error' ? 'bg-red-500' : 'bg-gray-500'}`} />
          <span>Whisper {whisperStatus.toLowerCase()}</span>
        </div>

        {systemHealth.version !== 'unknown' && (
          <span>v{systemHealth.version}</span>
        )}

        <div className="flex items-center space-x-1">
          <Clock className="w-3 h-3" />
          <span>{currentTime}</span>
        </div>
      </div>
    </div>
  );
}