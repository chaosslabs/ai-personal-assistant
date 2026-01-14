import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';
import {
  ArrowLeft,
  Calendar,
  Clock,
  Mic,
  Video,
  FileText,
  Loader2,
  Download,
  Copy,
  Volume2,
  AlertCircle
} from 'lucide-react';
import { ActivityService } from '../services/ActivityService';
import { Activity, TranscriptChunk } from '../types/activity';
import { toast } from 'sonner';

interface ActivityDetailsProps {
  activityId: string;
  onBack: () => void;
}

export function ActivityDetails({ activityId, onBack }: ActivityDetailsProps) {
  const [activity, setActivity] = useState<Activity | null>(null);
  const [transcript, setTranscript] = useState<TranscriptChunk[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [transcriptLoading, setTranscriptLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const activityService = new ActivityService();

  useEffect(() => {
    loadActivityDetails();
  }, [activityId]);

  const loadActivityDetails = async () => {
    try {
      setIsLoading(true);
      setError(null);

      // Load activity details
      const activityData = await activityService.getActivity(activityId);
      setActivity(activityData);

      // Try to load transcript if activity is completed
      if (activityData.status === 'completed') {
        await loadTranscript();
      }
    } catch (error) {
      console.error('Failed to load activity details:', error);
      setError('Failed to load activity details');
    } finally {
      setIsLoading(false);
    }
  };

  const loadTranscript = async () => {
    try {
      setTranscriptLoading(true);
      const transcriptData = await activityService.getActivityTranscript(activityId);
      setTranscript(transcriptData || []);
    } catch (error) {
      console.error('Failed to load transcript:', error);
      // Don't set error - transcript might not exist yet
      setTranscript([]);
    } finally {
      setTranscriptLoading(false);
    }
  };

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'meeting':
        return <Video className="w-5 h-5" />;
      case 'work_session':
        return <Clock className="w-5 h-5" />;
      case 'call':
        return <Mic className="w-5 h-5" />;
      default:
        return <FileText className="w-5 h-5" />;
    }
  };

  const formatTime = (timeString: string) => {
    const date = new Date(timeString);
    return date.toLocaleString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    });
  };

  const formatDuration = (startTime: string, endTime?: string) => {
    const start = new Date(startTime);
    const end = endTime ? new Date(endTime) : new Date();
    const diffMs = end.getTime() - start.getTime();
    const diffMinutes = Math.floor(diffMs / (1000 * 60));

    if (diffMinutes < 60) {
      return `${diffMinutes} min`;
    } else {
      const hours = Math.floor(diffMinutes / 60);
      const minutes = diffMinutes % 60;
      return `${hours}h ${minutes}min`;
    }
  };

  const formatTranscriptTime = (seconds: number) => {
    if (typeof seconds !== 'number' || isNaN(seconds)) {
      return '0:00';
    }
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const copyTranscriptToClipboard = async () => {
    try {
      const fullText = transcript.map(chunk => chunk.text).join(' ');
      await navigator.clipboard.writeText(fullText);

      toast.success('Copied to Clipboard', {
        description: 'Transcript text has been copied to your clipboard.',
        duration: 3000,
      });
    } catch (error) {
      console.error('Failed to copy transcript:', error);
      toast.error('Copy Failed', {
        description: 'Could not copy transcript to clipboard.',
        duration: 3000,
      });
    }
  };

  const downloadTranscript = () => {
    try {
      const fullText = transcript.map(chunk =>
        `[${formatTranscriptTime(chunk.start_time)}] ${chunk.text}`
      ).join('\n\n');

      const blob = new Blob([fullText], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${activity?.title || 'transcript'}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success('Download Started', {
        description: `Transcript saved as "${activity?.title || 'transcript'}.txt"`,
        duration: 3000,
      });
    } catch (error) {
      console.error('Failed to download transcript:', error);
      toast.error('Download Failed', {
        description: 'Could not download transcript file.',
        duration: 3000,
      });
    }
  };

  if (isLoading) {
    return (
      <div className="p-8 max-w-7xl mx-auto">
        <div className="flex items-center justify-center h-96">
          <div className="text-center">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
            <p className="text-muted-foreground">Loading activity details...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !activity) {
    return (
      <div className="p-8 max-w-7xl mx-auto">
        <Button variant="ghost" onClick={onBack} className="mb-4">
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back
        </Button>
        <Card className="rounded-2xl">
          <CardContent className="p-12 text-center">
            <AlertCircle className="w-12 h-12 text-destructive mx-auto mb-4" />
            <h3 className="font-semibold mb-2">Failed to load activity</h3>
            <p className="text-muted-foreground mb-4">{error || 'Activity not found'}</p>
            <Button onClick={loadActivityDetails}>Try Again</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const hasTranscript = transcript.length > 0;

  return (
    <div className="p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <Button variant="ghost" onClick={onBack} className="mb-4">
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Activities
        </Button>

        <div className="flex items-start justify-between">
          <div className="flex items-start space-x-4">
            <div className={`p-3 rounded-xl ${
              activity.type === 'meeting' ? 'bg-blue-500/10 text-blue-500' :
              activity.type === 'work_session' ? 'bg-green-500/10 text-green-500' :
              'bg-purple-500/10 text-purple-500'
            }`}>
              {getActivityIcon(activity.type)}
            </div>
            <div>
              <h1 className="text-3xl font-semibold mb-2">{activity.title}</h1>
              <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                <div className="flex items-center space-x-1">
                  <Calendar className="w-4 h-4" />
                  <span>{formatTime(activity.start_time)}</span>
                </div>
                <div className="flex items-center space-x-1">
                  <Clock className="w-4 h-4" />
                  <span>{formatDuration(activity.start_time, activity.end_time)}</span>
                </div>
                <Badge variant={activity.status === 'completed' ? 'secondary' : 'default'}>
                  {activity.status}
                </Badge>
              </div>
            </div>
          </div>

          {hasTranscript && (
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" onClick={copyTranscriptToClipboard}>
                <Copy className="w-4 h-4 mr-2" />
                Copy
              </Button>
              <Button variant="outline" size="sm" onClick={downloadTranscript}>
                <Download className="w-4 h-4 mr-2" />
                Download
              </Button>
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content - Transcript */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="rounded-2xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center space-x-2">
                  <FileText className="w-5 h-5" />
                  <span>Transcript</span>
                </CardTitle>
                {hasTranscript && (
                  <Badge variant="secondary" className="text-xs">
                    {transcript.length} chunks
                  </Badge>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {transcriptLoading ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2 className="w-6 h-6 animate-spin text-primary" />
                  <span className="ml-2 text-muted-foreground">Loading transcript...</span>
                </div>
              ) : hasTranscript ? (
                <div className="space-y-4">
                  {transcript.map((chunk, index) => (
                    <div key={chunk.id || index} className="flex space-x-3 group">
                      <div className="flex-shrink-0 w-16 text-xs text-muted-foreground pt-1">
                        {formatTranscriptTime(chunk.start_time)}
                      </div>
                      <div className="flex-1">
                        <p className="text-base leading-relaxed">{chunk.text}</p>
                        {chunk.confidence && chunk.confidence > 0 && (
                          <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                            Confidence: {Math.round(chunk.confidence * 100)}%
                          </span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-12">
                  <FileText className="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-50" />
                  <h3 className="font-semibold mb-2">No transcript available</h3>
                  <p className="text-sm text-muted-foreground">
                    {activity.status === 'completed'
                      ? 'This activity has not been transcribed yet. Use the Transcribe button on the activity card to generate a transcript.'
                      : 'Transcription is only available for completed activities.'}
                  </p>
                  {/* TODO: Add transcription trigger from details page
                  {activity.status === 'completed' && (
                    <Button onClick={() => {
                      console.log('Trigger transcription for', activityId);
                    }}>
                      Start Transcription
                    </Button>
                  )}
                  */}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar - Metadata & Actions */}
        <div className="space-y-6">
          {/* Activity Info */}
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">Activity Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-xs font-medium text-muted-foreground">Type</label>
                <p className="text-sm mt-1 capitalize">{activity.type.replace('_', ' ')}</p>
              </div>
              <Separator />
              <div>
                <label className="text-xs font-medium text-muted-foreground">Status</label>
                <p className="text-sm mt-1 capitalize">{activity.status}</p>
              </div>
              <Separator />
              <div>
                <label className="text-xs font-medium text-muted-foreground">Start Time</label>
                <p className="text-sm mt-1">{formatTime(activity.start_time)}</p>
              </div>
              {activity.end_time && (
                <>
                  <Separator />
                  <div>
                    <label className="text-xs font-medium text-muted-foreground">End Time</label>
                    <p className="text-sm mt-1">{formatTime(activity.end_time)}</p>
                  </div>
                </>
              )}
              <Separator />
              <div>
                <label className="text-xs font-medium text-muted-foreground">Duration</label>
                <p className="text-sm mt-1">{formatDuration(activity.start_time, activity.end_time)}</p>
              </div>
              {activity.tags && activity.tags.length > 0 && (
                <>
                  <Separator />
                  <div>
                    <label className="text-xs font-medium text-muted-foreground">Tags</label>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {activity.tags.map((tag, index) => (
                        <Badge key={index} variant="outline" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          {/* TODO: Audio Playback
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">Audio Recording</CardTitle>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full" disabled>
                <Volume2 className="w-4 h-4 mr-2" />
                Play Recording
              </Button>
              <p className="text-xs text-muted-foreground mt-2 text-center">
                Audio playback coming soon
              </p>
            </CardContent>
          </Card>
          */}

          {/* Actions */}
          {hasTranscript && (
            <Card className="rounded-2xl">
              <CardHeader>
                <CardTitle className="text-base">Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => loadTranscript()}
                >
                  <FileText className="w-4 h-4 mr-2" />
                  Refresh Transcript
                </Button>
              </CardContent>
            </Card>
          )}
          {/* TODO: Add transcription trigger and other actions
          {!hasTranscript && activity.status === 'completed' && (
            <Card className="rounded-2xl">
              <CardHeader>
                <CardTitle className="text-base">Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant="default"
                  className="w-full"
                  onClick={() => console.log('Start transcription')}
                >
                  <FileText className="w-4 h-4 mr-2" />
                  Transcribe Activity
                </Button>
              </CardContent>
            </Card>
          )}
          */}
        </div>
      </div>
    </div>
  );
}
