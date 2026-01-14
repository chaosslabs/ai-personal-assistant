import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Tabs, TabsList, TabsTrigger } from './ui/tabs';
import { ThemeToggle } from './ThemeToggle';
import { SystemService } from '../services/SystemService';
import { RecordingService } from '../services/RecordingService';
import { User } from '../types/user';
import {
  Home,
  Activity,
  Settings,
  Mic,
  MicOff,
  Minimize2,
  Maximize2,
  X
} from 'lucide-react';

interface TitleBarProps {
  currentView: string;
  onViewChange: (view: 'dashboard' | 'activities' | 'settings') => void;
  isRecording: boolean;
  onToggleRecording: () => void;
}

export function TitleBar({ currentView, onViewChange, isRecording, onToggleRecording }: TitleBarProps) {
  const [appVersion, setAppVersion] = useState<string>('Loading...');
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [currentRecordingId, setCurrentRecordingId] = useState<string | null>(null);
  const systemService = new SystemService();
  const recordingService = new RecordingService();

  useEffect(() => {
    const loadSystemData = async () => {
      try {
        const [version, user] = await Promise.all([
          systemService.getVersion(),
          systemService.getCurrentUser()
        ]);

        setAppVersion(version);
        setCurrentUser(user);
      } catch (error) {
        console.error('Failed to load system data:', error);
        setAppVersion('Error loading version');
      } finally {
        setIsLoading(false);
      }
    };

    loadSystemData();
  }, []);

  const handleToggleRecording = async () => {
    try {
      if (isRecording && currentRecordingId) {
        // Stop the current recording
        await recordingService.stopRecording(currentRecordingId);
        setCurrentRecordingId(null);
        onToggleRecording(); // Update parent state
      } else {
        // Start a new recording
        const session = await recordingService.startRecording();
        setCurrentRecordingId(session.id);
        onToggleRecording(); // Update parent state
      }
    } catch (error) {
      console.error('Failed to toggle recording:', error);
      // Could add toast notification here for user feedback
    }
  };

  return (
    <div className="flex items-center justify-between px-4 py-2 bg-background border-b border-border">
      {/* Left: App Info & Recording Status */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2">
          <div className="w-4 h-4 bg-gradient-to-br from-blue-500 to-purple-600 rounded-sm flex items-center justify-center">
            <div className="w-2 h-2 bg-white rounded-sm" />
          </div>
          <div className="flex flex-col">
            <span className="font-medium text-sm">Memoria</span>
            {!isLoading && (
              <span className="text-xs text-muted-foreground">v{appVersion}</span>
            )}
          </div>
        </div>

        {isRecording && (
          <Badge variant="destructive" className="animate-pulse text-xs">
            <div className="w-2 h-2 bg-white rounded-full mr-1.5" />
            Recording
          </Badge>
        )}
      </div>

      {/* Center: Navigation Tabs */}
      <Tabs value={currentView} onValueChange={onViewChange} className="flex-1 max-w-md mx-8">
        <TabsList className="grid w-full grid-cols-3 h-8 p-1 bg-secondary/50">
          <TabsTrigger value="dashboard" className="text-xs px-3 py-1">
            <Home className="w-3 h-3 mr-1.5" />
            Dashboard
          </TabsTrigger>
          <TabsTrigger value="activities" className="text-xs px-3 py-1">
            <Activity className="w-3 h-3 mr-1.5" />
            Activities
          </TabsTrigger>
          <TabsTrigger value="settings" className="text-xs px-3 py-1">
            <Settings className="w-3 h-3 mr-1.5" />
            Settings
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Right: User Info, Theme Toggle, Recording Control & Window Controls */}
      <div className="flex items-center space-x-2">
        {/* User Info */}
        {currentUser && (
          <div className="text-xs text-muted-foreground mr-2">
            {currentUser.username}
          </div>
        )}

        {/* Theme Toggle */}
        <ThemeToggle />

      </div>
    </div>
  );
}