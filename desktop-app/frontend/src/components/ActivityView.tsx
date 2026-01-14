import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { ActivityCard } from './ActivityCard';
import { ActivityService } from '../services/ActivityService';
import { Activity as ActivityType } from '../types/activity';
import {
  Search,
  // Filter, // TODO: Implement filters
  // Plus, // TODO: Implement scheduled/planned activities
  Clock,
  Calendar,
  FileText,
  // Download, // TODO: Implement export
  // BarChart3 // TODO: Implement analytics
} from 'lucide-react';

// Transform backend activity data for UI display
const transformActivityForUI = (activity: ActivityType) => {
  return {
    id: activity.id,
    title: activity.title,
    type: activity.type,
    status: mapActivityStatus(activity.status),
    duration: calculateDuration(activity),
    startTime: activity.start_time,
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

interface ActivityViewProps {
  onViewActivity?: (activityId: string) => void;
}

export function ActivityView({ onViewActivity }: ActivityViewProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState('all');
  const [activities, setActivities] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const activityService = new ActivityService();

  // Load activities from backend
  const loadActivities = async () => {
    try {
      setIsLoading(true);
      const backendActivities = await activityService.getActivities();

      // Transform activities for UI
      const uiActivities = backendActivities.map(transformActivityForUI);
      setActivities(uiActivities);
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

  useEffect(() => {
    loadActivities();
    // Refresh every 30 seconds
    const interval = setInterval(loadActivities, 30000);
    return () => clearInterval(interval);
  }, []);

  const filterActivitiesByTab = (tab: string) => {
    if (tab === 'all') return activities;
    if (tab === 'active') return activities.filter(a => a.status === 'active');
    if (tab === 'completed') return activities.filter(a => a.status === 'completed');
    if (tab === 'scheduled') return activities.filter(a => a.status === 'scheduled');
    return activities;
  };

  const filteredActivities = filterActivitiesByTab(activeTab).filter(activity =>
    activity.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    activity.summary.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const stats = {
    total: activities.length,
    active: activities.filter(a => a.status === 'active').length,
    completed: activities.filter(a => a.status === 'completed').length,
    transcripts: activities.filter(a => a.transcriptAvailable).length
  };

  return (
    <div className="p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-semibold mb-2">Activities</h1>
          <p className="text-muted-foreground">
            Manage and review all your recorded activities
          </p>
        </div>
        {/* TODO: Implement scheduled/planned activities
        <Button className="rounded-full" onClick={() => handleCreateActivity('meeting')}>
          <Plus className="w-4 h-4 mr-2" />
          New Activity
        </Button>
        */}
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <Card className="rounded-2xl">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-semibold">{stats.total}</p>
                <p className="text-sm text-muted-foreground">Total Activities</p>
              </div>
              <Calendar className="w-8 h-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-semibold text-red-500">{stats.active}</p>
                <p className="text-sm text-muted-foreground">Active</p>
              </div>
              <div className="w-8 h-8 bg-red-500 rounded-full animate-pulse" />
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-semibold text-green-500">{stats.completed}</p>
                <p className="text-sm text-muted-foreground">Completed</p>
              </div>
              <Clock className="w-8 h-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-2xl">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-semibold text-purple-500">{stats.transcripts}</p>
                <p className="text-sm text-muted-foreground">Transcripts</p>
              </div>
              <FileText className="w-8 h-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="rounded-2xl mb-8">
        <CardContent className="p-6">
          <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
            <div className="flex-1 max-w-md">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                  placeholder="Search activities, transcripts, or notes..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 rounded-2xl"
                />
              </div>
            </div>

            {/* TODO: Implement filters, export, and analytics features
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" className="rounded-full">
                <Filter className="w-4 h-4 mr-2" />
                Filters
              </Button>
              <Button variant="outline" size="sm" className="rounded-full">
                <Download className="w-4 h-4 mr-2" />
                Export
              </Button>
              <Button variant="outline" size="sm" className="rounded-full">
                <BarChart3 className="w-4 h-4 mr-2" />
                Analytics
              </Button>
            </div>
            */}
          </div>
        </CardContent>
      </Card>

      {/* Activity Tabs and List */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-3 rounded-2xl p-1 bg-muted">
          <TabsTrigger value="all" className="rounded-2xl">
            All ({stats.total})
          </TabsTrigger>
          <TabsTrigger value="active" className="rounded-2xl">
            Active ({stats.active})
          </TabsTrigger>
          <TabsTrigger value="completed" className="rounded-2xl">
            Completed ({stats.completed})
          </TabsTrigger>
          {/* TODO: Implement scheduled activities backend status
          <TabsTrigger value="scheduled" className="rounded-2xl">
            Scheduled ({activities.filter(a => a.status === 'scheduled').length})
          </TabsTrigger>
          */}
        </TabsList>

        <TabsContent value={activeTab} className="space-y-6">
          {isLoading ? (
            <Card className="rounded-2xl">
              <CardContent className="p-12 text-center">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                  <div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                </div>
                <h3 className="font-semibold mb-2">Loading activities...</h3>
                <p className="text-muted-foreground">
                  Fetching your latest activities
                </p>
              </CardContent>
            </Card>
          ) : filteredActivities.length === 0 ? (
            <Card className="rounded-2xl">
              <CardContent className="p-12 text-center">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                  <Search className="w-8 h-8 text-muted-foreground" />
                </div>
                <h3 className="font-semibold mb-2">No activities found</h3>
                <p className="text-muted-foreground">
                  {searchQuery
                    ? 'Try adjusting your search terms'
                    : 'No activities in this category yet. Use "Start" from the Dashboard to start recording.'}
                </p>
                {/* TODO: Implement scheduled/planned activities
                <Button className="rounded-full" onClick={() => handleCreateActivity('meeting')}>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Activity
                </Button>
                */}
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {filteredActivities.map((activity) => (
                <ActivityCard
                  key={activity.id}
                  activity={activity}
                  onActivityUpdate={loadActivities}
                  onViewDetails={onViewActivity}
                />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}