import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from './ui/alert-dialog';
import {
  Calendar,
  Clock,
  Mic,
  Video,
  // Play, // TODO: Implement audio playback
  FileText,
  MoreHorizontal,
  Volume2,
  Loader2,
  Trash2,
  FileAudio
} from 'lucide-react';
import { ActivityService } from '../services/ActivityService';
import { RecordingService } from '../services/RecordingService';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { toast } from 'sonner';

interface Activity {
  id: string;
  title: string;
  type: 'meeting' | 'work_session' | 'call';
  status: 'active' | 'completed' | 'scheduled';
  duration: string;
  startTime: string;
  transcriptAvailable: boolean;
  summary: string;
}

interface ActivityCardProps {
  activity: Activity;
  onActivityUpdate?: () => void; // Callback to refresh parent component
  onViewDetails?: (activityId: string) => void; // Callback to view activity details
}

export function ActivityCard({ activity, onActivityUpdate, onViewDetails }: ActivityCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isTranscribing, setIsTranscribing] = useState(false);
  const [transcriptionProgress, setTranscriptionProgress] = useState(0);
  const [transcriptionError, setTranscriptionError] = useState<string | null>(null);
  const activityService = new ActivityService();
  const recordingService = new RecordingService();

  useEffect(() => {
    // Listen for transcription events
    const unsubscribeStarted = EventsOn('transcription:started', (activityId: string) => {
      if (activityId === activity.id) {
        setIsTranscribing(true);
        setTranscriptionProgress(0);
        setTranscriptionError(null);
      }
    });

    const unsubscribeProgress = EventsOn('transcription:progress', (data: any) => {
      if (data.activity_id === activity.id) {
        setTranscriptionProgress(data.progress * 100);
      }
    });

    const unsubscribeCompleted = EventsOn('transcription:completed', (activityId: string) => {
      if (activityId === activity.id) {
        setIsTranscribing(false);
        setTranscriptionProgress(100);
        console.log('Transcription completed for activity:', activityId);

        // Show success toast
        toast.success('Transcription Complete', {
          description: `"${activity.title}" has been transcribed successfully.`,
          duration: 5000,
        });

        if (onActivityUpdate) {
          onActivityUpdate(); // Refresh to show transcript
        }
      }
    });

    const unsubscribeError = EventsOn('transcription:error', (data: any) => {
      if (data.activity_id === activity.id) {
        setIsTranscribing(false);
        setTranscriptionError(data.error);
        console.error('Transcription error:', data.error);

        // Show error toast
        toast.error('Transcription Failed', {
          description: data.error,
          duration: 7000,
        });
      }
    });

    return () => {
      // Cleanup event listeners
      EventsOff('transcription:started');
      EventsOff('transcription:progress');
      EventsOff('transcription:completed');
      EventsOff('transcription:error');
    };
  }, [activity.id, onActivityUpdate]);
  const getActivityIcon = () => {
    switch (activity.type) {
      case 'meeting':
        return <Video className="w-4 h-4" />;
      case 'work_session':
        return <Clock className="w-4 h-4" />;
      case 'call':
        return <Mic className="w-4 h-4" />;
    }
  };

  const getStatusBadge = () => {
    switch (activity.status) {
      case 'active':
        return (
          <Badge variant="destructive" className="animate-pulse text-xs h-5">
            <div className="w-1.5 h-1.5 bg-white rounded-full mr-1" />
            Recording
          </Badge>
        );
      case 'completed':
        return <Badge variant="secondary" className="text-xs h-5">Completed</Badge>;
      case 'scheduled':
        return <Badge variant="outline" className="text-xs h-5">Scheduled</Badge>;
    }
  };

  const getTypeLabel = () => {
    switch (activity.type) {
      case 'meeting':
        return 'Meeting';
      case 'work_session':
        return 'Work Session';
      case 'call':
        return 'Call';
    }
  };

  const formatTime = (timeString: string) => {
    const date = new Date(timeString);
    return date.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    });
  };

  const handleStopActivity = async () => {
    try {
      setIsLoading(true);
      await activityService.stopActivity(activity.id);
      if (onActivityUpdate) {
        onActivityUpdate(); // Refresh the parent component
      }
    } catch (error) {
      console.error('Failed to stop activity:', error);
      // Could add toast notification here
    } finally {
      setIsLoading(false);
    }
  };

  const handleStartActivity = async () => {
    try {
      setIsLoading(true);
      await activityService.startActivity(activity.id);
      if (onActivityUpdate) {
        onActivityUpdate(); // Refresh the parent component
      }
    } catch (error) {
      console.error('Failed to start activity:', error);
      // Could add toast notification here
    } finally {
      setIsLoading(false);
    }
  };

  // TODO: Implement transcript viewer modal/view
  // const handleViewTranscript = async () => {
  //   try {
  //     setIsLoading(true);
  //     const transcript = await activityService.getActivityTranscript(activity.id);
  //     // Open a modal or navigate to a transcript view
  //     console.log('Activity transcript:', transcript);
  //   } catch (error) {
  //     console.error('Failed to get transcript:', error);
  //   } finally {
  //     setIsLoading(false);
  //   }
  // };

  // TODO: Implement audio playback functionality
  // const handlePlayAudio = () => {
  //   console.log('Playing audio for activity:', activity.id);
  // };

  const handleTranscribe = async () => {
    try {
      setIsLoading(true);
      setTranscriptionError(null);
      console.log('Starting transcription for activity:', activity.id);
      await activityService.processTranscription(activity.id);
      console.log('Transcription request sent successfully');

      // Show info toast
      toast.info('Transcription Started', {
        description: `Processing "${activity.title}"...`,
        duration: 3000,
      });

      // Events will handle the rest (started, progress, completed/error)
    } catch (error) {
      console.error('Failed to start transcription:', error);
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      setTranscriptionError(errorMessage);

      toast.error('Failed to Start Transcription', {
        description: errorMessage,
        duration: 5000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteActivity = async () => {
    console.log(`🗑️  UI: Starting delete for activity ${activity.id} (${activity.title})`);
    try {
      setIsLoading(true);
      console.log('🗑️  UI: Calling ActivityService.deleteActivity...');
      await activityService.deleteActivity(activity.id);
      console.log('🗑️  UI: ActivityService.deleteActivity completed successfully');
      setShowDeleteDialog(false);
      if (onActivityUpdate) {
        console.log('🗑️  UI: Calling onActivityUpdate to refresh parent...');
        onActivityUpdate(); // Refresh the parent component
        console.log('🗑️  UI: onActivityUpdate completed');
      }
      console.log('🗑️  UI: Delete operation completed successfully');
    } catch (error) {
      console.error('🗑️  UI: Failed to delete activity:', error);
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      console.error('🗑️  UI: Error details:', errorMessage);
      // Could add toast notification here
      alert(`Failed to delete activity: ${errorMessage}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't navigate if clicking on buttons or interactive elements
    const target = e.target as HTMLElement;
    if (
      target.closest('button') ||
      target.closest('[role="button"]') ||
      target.closest('[role="menuitem"]') ||
      target.closest('[role="menu"]') ||
      target.closest('[data-radix-menu-content]') ||
      target.closest('[data-radix-dropdown-menu-trigger]')
    ) {
      return;
    }
    if (onViewDetails) {
      onViewDetails(activity.id);
    }
  };

  return (
    <Card
      className={`rounded-xl transition-all duration-200 hover:shadow-md hover:-translate-y-0.5 ${
        activity.status === 'active' ? 'ring-1 ring-red-500/30' : ''
      } ${onViewDetails ? 'cursor-pointer' : ''}`}
      onClick={handleCardClick}
    >
      {activity.status === 'active' && (
        <div className="h-0.5 bg-gradient-to-r from-red-500 to-red-400 rounded-t-xl" />
      )}

      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-2">
            <div className={`p-1.5 rounded-lg ${
              activity.type === 'meeting' ? 'bg-blue-500/10 text-blue-500' :
              activity.type === 'work_session' ? 'bg-green-500/10 text-green-500' :
              'bg-purple-500/10 text-purple-500'
            }`}>
              {getActivityIcon()}
            </div>
            <div>
              <h3 className="font-medium text-sm">{activity.title}</h3>
              <p className="text-xs text-muted-foreground">{getTypeLabel()}</p>
            </div>
          </div>

          <div className="flex items-center space-x-1">
            {getStatusBadge()}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="rounded-lg p-1.5 h-6 w-6"
                  onClick={(e) => e.stopPropagation()}
                >
                  <MoreHorizontal className="w-3 h-3" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-40">
                <DropdownMenuItem
                  variant="destructive"
                  className="cursor-pointer"
                  onClick={(e) => {
                    e.stopPropagation();
                    setShowDeleteDialog(true);
                  }}
                >
                  <Trash2 className="w-3 h-3 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Activity</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete "{activity.title}"? This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDeleteActivity}
                    disabled={isLoading}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    {isLoading ? (
                      <>
                        <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                        Deleting...
                      </>
                    ) : (
                      'Delete'
                    )}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-3 pt-0">
        {/* Activity Details */}
        <div className="grid grid-cols-2 gap-3 text-xs">
          <div className="flex items-center space-x-1.5 text-muted-foreground">
            <Calendar className="w-3 h-3" />
            <span>{formatTime(activity.startTime)}</span>
          </div>
          <div className="flex items-center space-x-1.5 text-muted-foreground">
            <Clock className="w-3 h-3" />
            <span>{activity.duration}</span>
          </div>
        </div>

        {/* Summary */}
        <p className="text-xs text-muted-foreground leading-relaxed line-clamp-2">
          {activity.summary}
        </p>

        {/* Action Buttons */}
        <div className="flex items-center justify-between pt-1">
          <div className="flex items-center space-x-1">
            {activity.transcriptAvailable && (
              <div className="flex items-center space-x-1 text-xs text-green-600 bg-green-50 dark:bg-green-900/20 px-2 py-0.5 rounded-md">
                <FileText className="w-2.5 h-2.5" />
                <span>Transcript Ready</span>
              </div>
            )}
            {transcriptionProgress === 100 && !activity.transcriptAvailable && (
              <div className="flex items-center space-x-1 text-xs text-green-600 bg-green-50 dark:bg-green-900/20 px-2 py-0.5 rounded-md animate-pulse">
                <FileText className="w-2.5 h-2.5" />
                <span>Just completed</span>
              </div>
            )}
          </div>

          <div className="flex items-center space-x-1">
            {activity.status === 'active' ? (
              <Button
                size="sm"
                variant="destructive"
                className="rounded-lg h-6 px-2 text-xs"
                onClick={handleStopActivity}
                disabled={isLoading}
              >
                {isLoading ? (
                  <Loader2 className="w-2 h-2 animate-spin mr-1" />
                ) : (
                  <div className="w-2 h-2 bg-white rounded-full mr-1" />
                )}
                Stop
              </Button>
            ) : activity.status === 'completed' ? (
              <>
                {!activity.transcriptAvailable && !isTranscribing && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="rounded-lg h-6 px-2 text-xs"
                    onClick={handleTranscribe}
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <Loader2 className="w-3 h-3 animate-spin mr-1" />
                    ) : (
                      <FileAudio className="w-3 h-3 mr-1" />
                    )}
                    Transcribe
                  </Button>
                )}
                {isTranscribing && (
                  <div className="flex items-center space-x-1 text-xs">
                    <Loader2 className="w-3 h-3 animate-spin" />
                    <span>{Math.round(transcriptionProgress)}%</span>
                  </div>
                )}
                {/* TODO: Implement transcript viewer
                {activity.transcriptAvailable && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="rounded-lg h-6 px-2 text-xs"
                    onClick={handleViewTranscript}
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <Loader2 className="w-3 h-3 animate-spin mr-1" />
                    ) : (
                      <FileText className="w-3 h-3 mr-1" />
                    )}
                    View
                  </Button>
                )}
                */}
                {/* TODO: Implement audio playback
                <Button
                  size="sm"
                  variant="outline"
                  className="rounded-lg h-6 px-2 text-xs"
                  onClick={handlePlayAudio}
                  disabled={isLoading}
                >
                  <Play className="w-3 h-3 mr-1" />
                  Play
                </Button>
                */}
              </>
            ) : (
              <Button
                size="sm"
                variant="outline"
                className="rounded-lg h-6 px-2 text-xs"
                onClick={handleStartActivity}
                disabled={isLoading}
              >
                {isLoading ? (
                  <Loader2 className="w-3 h-3 animate-spin mr-1" />
                ) : (
                  <Volume2 className="w-3 h-3 mr-1" />
                )}
                Start
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}