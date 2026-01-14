import { useState, useEffect, useRef, useCallback } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Input } from './ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { ActivityCard } from './ActivityCard';
import {
  // Plus, // TODO: Implement scheduled/planned activities
  Search,
  Clock,
  Mic,
  Video,
  Calendar,
  Activity
} from 'lucide-react';
import { ActivityService } from '../services/ActivityService';
import { recordingService } from '../services/RecordingService';
import { Activity as ActivityType, ActivityType as ActivityTypeEnum } from '../types/activity';

// Hook for debounced auto-save
function useDebouncedCallback<T extends (...args: any[]) => void>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const debouncedCallback = useCallback(
    (...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    },
    [callback, delay]
  ) as T;

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return debouncedCallback;
}

// Transform backend activity data for UI display
const transformActivityForUI = (activity: ActivityType) => {
  return {
    id: activity.id,
    title: activity.title,
    type: activity.type,
    status: mapActivityStatus(activity.status),
    duration: calculateDuration(activity),
    startTime: activity.start_time,
    endTime: activity.end_time, // Preserve end_time for duration calculations
    transcriptAvailable: false, // Will be determined by checking for transcript data
    summary: getActivitySummary(activity)
  };
};

const mapActivityStatus = (status: string): 'active' | 'completed' | 'scheduled' => {
  switch (status) {
    case 'active':
    case 'recording':
      return 'active';
    case 'completed':
      return 'completed';
    default:
      return 'scheduled';
  }
};

const calculateDuration = (activity: ActivityType): string => {
  if (!activity.start_time) return 'Scheduled';

  const start = new Date(activity.start_time);
  let end: Date;

  // If activity is still active/recording, use current time
  if (activity.status === 'active' || activity.status === 'recording') {
    end = new Date();
  } else if (activity.end_time) {
    // For completed activities, use end_time
    end = new Date(activity.end_time);
  } else {
    // Fallback for completed activities without end_time
    return '< 1 min';
  }

  const diffMs = end.getTime() - start.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);

  // Handle very short durations (less than 1 minute)
  if (diffMinutes === 0 && diffSeconds > 0) {
    return `${diffSeconds}s`;
  }

  // Handle regular durations
  if (diffMinutes < 60) {
    return `${diffMinutes} min`;
  } else {
    const hours = Math.floor(diffMinutes / 60);
    const minutes = diffMinutes % 60;
    return `${hours}h ${minutes}min`;
  }
};

const getActivitySummary = (activity: ActivityType): string => {
  if (activity.metadata?.summary) {
    return activity.metadata.summary;
  }

  const typeLabel = activity.type.replace('_', ' ');
  return `${typeLabel} session`;
};

interface DashboardProps {
  isRecording: boolean;
  onToggleRecording: () => void;
}

export function Dashboard({ isRecording, onToggleRecording }: DashboardProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<'all' | 'meeting' | 'work_session' | 'call'>('all');
  const [activities, setActivities] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [recordingDuration, setRecordingDuration] = useState(0);
  const [currentActivityTitle, setCurrentActivityTitle] = useState('');
  const [currentActivityType, setCurrentActivityType] = useState<string>('other');
  const [currentActivityId, setCurrentActivityId] = useState<string | null>(null);
  const [stats, setStats] = useState({
    todayActivities: 0,
    totalDuration: '0min',
    transcriptHours: '0min',
    activeRecordings: 0
  });

  const activityService = new ActivityService();

  // Auto-save activity title after user stops typing
  const saveActivityTitle = useCallback(async (activityId: string, newTitle: string) => {
    if (!activityId || !newTitle.trim()) return;
    try {
      await activityService.updateActivityTitle(activityId, newTitle.trim());
      recordingService.updateCurrentActivityTitle(newTitle.trim());
    } catch (error) {
      console.error('Failed to save activity title:', error);
    }
  }, []);

  const debouncedSaveTitle = useDebouncedCallback(saveActivityTitle, 500);

  // Save activity type immediately when changed
  const saveActivityType = useCallback(async (activityId: string, newType: string) => {
    if (!activityId || !newType) return;
    try {
      await activityService.updateActivityType(activityId, newType as ActivityTypeEnum);
      recordingService.updateCurrentActivityType(newType);
    } catch (error) {
      console.error('Failed to save activity type:', error);
    }
  }, []);

  // Handle title change - update local state immediately and debounce save
  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTitle = e.target.value;
    setCurrentActivityTitle(newTitle);
    if (currentActivityId && newTitle.trim()) {
      debouncedSaveTitle(currentActivityId, newTitle);
    }
  };

  // Handle type change - update local state and save immediately
  const handleTypeChange = (newType: string) => {
    setCurrentActivityType(newType);
    if (currentActivityId) {
      saveActivityType(currentActivityId, newType);
    }
  };

  // Load activities from backend
  const loadActivities = async () => {
    try {
      setIsLoading(true);
      const backendActivities = await activityService.getActivities();

      // Transform activities for UI
      const uiActivities = backendActivities.map(transformActivityForUI);
      setActivities(uiActivities);

      // Calculate stats
      const today = new Date().toDateString();
      const todayActivities = uiActivities.filter(activity => {
        if (!activity.startTime) return false;
        return new Date(activity.startTime).toDateString() === today;
      });

      // Calculate total duration from completed activities
      const totalMinutes = todayActivities.reduce((acc, activity) => {
        if (activity.status === 'completed' && activity.startTime) {
          const start = new Date(activity.startTime);
          // Use end_time if available, otherwise use current time for active recordings
          const end = activity.endTime ? new Date(activity.endTime) : new Date();
          const diffMs = end.getTime() - start.getTime();
          return acc + Math.floor(diffMs / (1000 * 60));
        }
        return acc;
      }, 0);

      const formatDuration = (minutes: number) => {
        if (minutes < 60) return `${minutes}min`;
        const hours = Math.floor(minutes / 60);
        const mins = minutes % 60;
        return `${hours}h ${mins}min`;
      };

      // TODO: Calculate actual transcribed time from transcript chunks
      // This requires checking each activity for transcripts and summing their durations
      // const transcribedMinutes = todayActivities.reduce((acc, activity) => {
      //   if (activity.transcriptAvailable && activity.status === 'completed') {
      //     return acc + getTranscriptDuration(activity.id);
      //   }
      //   return acc;
      // }, 0);
      const transcribedMinutes = 0; // Placeholder until transcript duration calculation is implemented

      setStats({
        todayActivities: todayActivities.length,
        totalDuration: formatDuration(totalMinutes),
        transcriptHours: formatDuration(transcribedMinutes),
        activeRecordings: uiActivities.filter(a => a.status === 'active').length
      });

    } catch (error) {
      console.error('Failed to load activities:', error);
      setActivities([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateActivity = async (type: 'meeting' | 'work_session' | 'call' = 'meeting') => {
    try {
      const title = `New ${type === 'work_session' ? 'Work Session' : type === 'meeting' ? 'Meeting' : 'Call'}`;
      await activityService.createActivity(type, title);
      // Refresh activities after creating
      await loadActivities();
    } catch (error) {
      console.error('Failed to create activity:', error);
    }
  };

  // Update recording duration and activity info when recording
  useEffect(() => {
    if (isRecording) {
      // Get initial activity info
      const currentActivity = recordingService.getCurrentActivity();
      if (currentActivity) {
        setCurrentActivityId(currentActivity.id);
        setCurrentActivityTitle(currentActivity.title);
        setCurrentActivityType(currentActivity.type || 'other');
      }

      const interval = setInterval(() => {
        setRecordingDuration(recordingService.getCurrentDuration());
        // Update activity info in case it changed
        const activity = recordingService.getCurrentActivity();
        if (activity && activity.id !== currentActivityId) {
          setCurrentActivityId(activity.id);
          setCurrentActivityTitle(activity.title);
          setCurrentActivityType(activity.type || 'other');
        }
      }, 1000);
      return () => clearInterval(interval);
    } else {
      setRecordingDuration(0);
      setCurrentActivityId(null);
      setCurrentActivityTitle('');
      setCurrentActivityType('other');
    }
  }, [isRecording]);

  useEffect(() => {
    loadActivities();
    // Refresh every 30 seconds
    const interval = setInterval(loadActivities, 30000);
    return () => clearInterval(interval);
  }, []);

  // Refresh activities when recording stops to get updated end_time
  useEffect(() => {
    if (!isRecording) {
      // Small delay to ensure backend has updated the activity
      const timeout = setTimeout(() => {
        loadActivities();
      }, 1000);
      return () => clearTimeout(timeout);
    }
  }, [isRecording]);

  const filteredActivities = activities.filter(activity => {
    const matchesSearch = activity.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      activity.summary.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = selectedType === 'all' || activity.type === selectedType;
    return matchesSearch && matchesType;
  });

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Compact Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold mb-1">Today's Activities</h1>
            <p className="text-sm text-muted-foreground">
              {new Date().toLocaleDateString('en-US', {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}
            </p>
          </div>

          {/* Quick Actions */}
          <div className="flex items-center space-x-2">
            <Button
              variant={isRecording ? "destructive" : "default"}
              onClick={onToggleRecording}
              className="h-8 px-4 text-sm"
            >
              {isRecording ? (
                <>
                  <div className="w-2 h-2 bg-white rounded-full mr-2 animate-pulse" />
                  Stop ({recordingService.formatDuration(recordingDuration)})
                </>
              ) : (
                <>
                  <Mic className="w-4 h-4 mr-2" />
                  Start
                </>
              )}
            </Button>

            {/* TODO: Implement scheduled/planned activities
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-3 text-sm"
              onClick={() => handleCreateActivity('meeting')}
            >
              <Plus className="w-4 h-4 mr-2" />
              New Activity
            </Button>
            */}
          </div>
        </div>
      </div>

      {/* Recording Activity - Editable title and type when recording */}
      {isRecording && currentActivityId && (
        <Card className="mb-6 border-destructive/30 bg-destructive/5">
          <CardContent className="p-4">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-destructive rounded-full animate-pulse" />
                <span className="text-sm font-medium text-destructive">Recording</span>
              </div>
              <div className="flex-1">
                <Input
                  value={currentActivityTitle}
                  onChange={handleTitleChange}
                  placeholder="Enter activity name..."
                  className="bg-transparent border-transparent hover:bg-background/30 focus:bg-background/50 focus:border-border/50 text-lg font-medium h-10 transition-colors"
                />
              </div>
              <Select value={currentActivityType} onValueChange={handleTypeChange}>
                <SelectTrigger className="w-[140px] h-10 bg-transparent border-border/50">
                  <SelectValue placeholder="Type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="meeting">Meeting</SelectItem>
                  <SelectItem value="work_session">Work Session</SelectItem>
                  <SelectItem value="call">Call</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
              <span className="text-sm text-muted-foreground tabular-nums">
                {recordingService.formatDuration(recordingDuration)}
              </span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Compact Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card className="rounded-xl">
          <CardContent className="p-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Activity className="w-4 h-4 text-primary" />
              </div>
              <div>
                <p className="text-xl font-semibold">{stats.todayActivities}</p>
                <p className="text-xs text-muted-foreground">Activities Today</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-xl">
          <CardContent className="p-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-blue-500/10 rounded-lg">
                <Clock className="w-4 h-4 text-blue-500" />
              </div>
              <div>
                <p className="text-xl font-semibold">{stats.totalDuration}</p>
                <p className="text-xs text-muted-foreground">Total Duration</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* TODO: Implement transcribed duration calculation from transcript chunks
        <Card className="rounded-xl">
          <CardContent className="p-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-green-500/10 rounded-lg">
                <Mic className="w-4 h-4 text-green-500" />
              </div>
              <div>
                <p className="text-xl font-semibold">{stats.transcriptHours}</p>
                <p className="text-xs text-muted-foreground">Transcribed</p>
              </div>
            </div>
          </CardContent>
        </Card>
        */}

        <Card className="rounded-xl">
          <CardContent className="p-4">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-red-500/10 rounded-lg">
                <div className={`w-4 h-4 rounded-full ${isRecording ? 'bg-red-500 animate-pulse' : 'bg-gray-400'
                  }`} />
              </div>
              <div>
                <p className="text-xl font-semibold">{isRecording ? '1' : '0'}</p>
                <p className="text-xs text-muted-foreground">
                  {isRecording ? 'Recording' : 'No Recording'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}