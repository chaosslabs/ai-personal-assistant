import { useState, useEffect } from 'react';
import { Dashboard } from './components/Dashboard';
import { ActivityView } from './components/ActivityView';
import { ActivityDetails } from './components/ActivityDetails';
import { SettingsView } from './components/SettingsView';
import { TitleBar } from './components/TitleBar';
import { StatusBar } from './components/StatusBar';
import { Toaster } from './components/ui/sonner';
import { recordingService } from './services/RecordingService';

type View = 'dashboard' | 'activities' | 'activity-details' | 'settings';

export default function App() {
  const [currentView, setCurrentView] = useState<View>('dashboard');
  const [selectedActivityId, setSelectedActivityId] = useState<string | null>(null);
  const [isRecording, setIsRecording] = useState(false);
  const [recordingError, setRecordingError] = useState<string | null>(null);

  // Subscribe to recording service state changes
  useEffect(() => {
    const unsubscribe = recordingService.subscribe((state) => {
      setIsRecording(state.isRecording);
      setRecordingError(state.error);
    });

    // Initialize with current state
    const currentState = recordingService.getState();
    setIsRecording(currentState.isRecording);
    setRecordingError(currentState.error);

    return unsubscribe;
  }, []);

  const handleToggleRecording = async () => {
    try {
      if (isRecording) {
        await recordingService.stopRecording();
      } else {
        await recordingService.startRecording();
      }
    } catch (error) {
      console.error('Recording toggle failed:', error);
      // Error is already managed by the recording service state
    }
  };

  const handleViewActivity = (activityId: string) => {
    setSelectedActivityId(activityId);
    setCurrentView('activity-details');
  };

  const handleBackToActivities = () => {
    setSelectedActivityId(null);
    setCurrentView('activities');
  };

  const renderView = () => {
    switch (currentView) {
      case 'dashboard':
        return <Dashboard isRecording={isRecording} onToggleRecording={handleToggleRecording} />;
      case 'activities':
        return <ActivityView onViewActivity={handleViewActivity} />;
      case 'activity-details':
        return selectedActivityId ? (
          <ActivityDetails activityId={selectedActivityId} onBack={handleBackToActivities} />
        ) : (
          <ActivityView onViewActivity={handleViewActivity} />
        );
      case 'settings':
        return <SettingsView />;
      default:
        return <Dashboard isRecording={isRecording} onToggleRecording={handleToggleRecording} />;
    }
  };

  return (
    <div className="size-full bg-background text-foreground flex flex-col">
      {/* Custom Title Bar */}
      <TitleBar
        currentView={currentView}
        onViewChange={setCurrentView}
        isRecording={isRecording}
        onToggleRecording={handleToggleRecording}
      />

      {/* Recording Error Banner */}
      {recordingError && (
        <div className="bg-destructive/10 border-b border-destructive/20 px-4 py-2 text-sm text-destructive">
          <div className="flex items-center justify-between">
            <span>Recording error: {recordingError}</span>
            <button
              onClick={() => recordingService.clearError()}
              className="text-destructive hover:text-destructive/80 underline"
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {renderView()}
      </main>

      {/* Status Bar */}
      <StatusBar isRecording={isRecording} />

      {/* Toast Notifications */}
      <Toaster />
    </div>
  );
}