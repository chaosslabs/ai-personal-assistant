import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Switch } from './ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Separator } from './ui/separator';
import { Badge } from './ui/badge';
import { useTheme } from './ThemeProvider';
import { SystemService } from '../services/SystemService';
import { ActivityService } from '../services/ActivityService';
import {
  Shield,
  Mic,
  Bell,
  HardDrive,
  Download,
  Trash2,
  Zap,
  Settings as SettingsIcon,
  Lock,
  Database,
  Palette,
  Sun,
  Moon,
  Monitor,
  RefreshCw
} from 'lucide-react';

export function SettingsView() {
  const { theme, setTheme } = useTheme();
  const [settings, setSettings] = useState({
    // Recording Settings
    audioDevice: 'default',
    audioQuality: 'high',
    autoTranscribe: true,
    backgroundRecording: false,

    // Privacy Settings
    dataRetention: '90', // days
    autoDelete: true,
    encryptData: true,

    // Notification Settings
    recordingNotifications: true,
    transcriptComplete: true,
    dailySummary: false,

    // Storage Settings
    localStoragePath: '~/Library/Application Support/personal-assist/',
    maxStorageSize: '10', // GB
    autoCleanup: true
  });

  // Real system data
  const [audioDevices, setAudioDevices] = useState([]);
  const [systemInfo, setSystemInfo] = useState(null);
  const [storageStats, setStorageStats] = useState({
    used: 'Loading...',
    available: 'Loading...',
    activities: 0,
    transcripts: 0,
    audioFiles: 0
  });
  const [isLoadingDevices, setIsLoadingDevices] = useState(false);

  const systemService = new SystemService();
  const activityService = new ActivityService();

  // Load real system data on mount
  useEffect(() => {
    const loadSystemData = async () => {
      try {
        // Load audio devices
        setIsLoadingDevices(true);
        const devices = await systemService.getAudioDevices();
        setAudioDevices(devices);

        // Load system info for storage stats
        const sysInfo = await systemService.getSystemInfo();
        setSystemInfo(sysInfo);

        // Load activities for stats
        const activities = await activityService.getActivities();

        // Calculate storage stats
        const totalActivities = activities.length;
        const transcribedActivities = activities.filter(a => a.status === 'completed').length;

        // Calculate storage usage
        let usedStorage = 'Unknown';
        let availableStorage = 'Unknown';

        if (sysInfo.disk_usage_bytes) {
          usedStorage = systemService.formatStorageUsage(sysInfo.disk_usage_bytes);
          // Assume 10GB limit for now (should come from settings)
          const totalBytes = 10 * 1024 * 1024 * 1024; // 10GB
          const availableBytes = totalBytes - sysInfo.disk_usage_bytes;
          availableStorage = systemService.formatStorageUsage(availableBytes);
        } else if (sysInfo.disk_usage_mb) {
          const usedBytes = sysInfo.disk_usage_mb * 1024 * 1024;
          usedStorage = systemService.formatStorageUsage(usedBytes);
          const totalBytes = 10 * 1024 * 1024 * 1024; // 10GB
          const availableBytes = totalBytes - usedBytes;
          availableStorage = systemService.formatStorageUsage(availableBytes);
        }

        setStorageStats({
          used: usedStorage,
          available: availableStorage,
          activities: totalActivities,
          transcripts: transcribedActivities,
          audioFiles: totalActivities // Assume 1 audio file per activity for now
        });

        // Update storage path from system info
        if (sysInfo.data_directory) {
          setSettings(prev => ({ ...prev, localStoragePath: sysInfo.data_directory }));
        }

        // Load user settings from backend
        const userSettings = await systemService.getUserSettings();
        if (userSettings) {
          setSettings(prev => ({
            ...prev,
            audioDevice: userSettings.preferred_audio_device || 'default',
            audioQuality: userSettings.audio_quality || 'high',
            autoTranscribe: userSettings.auto_start_recording || false,
            // Map other settings as needed
          }));
        }

      } catch (error) {
        console.error('Failed to load system data:', error);
      } finally {
        setIsLoadingDevices(false);
      }
    };

    loadSystemData();
  }, []);

  const handleSettingChange = async (key: string, value: any) => {
    setSettings(prev => ({ ...prev, [key]: value }));

    try {
      // Map frontend settings to backend format
      const backendSettings: any = {};

      if (key === 'audioDevice') {
        backendSettings.preferred_audio_device = value;
      } else if (key === 'audioQuality') {
        backendSettings.audio_quality = value;
      } else if (key === 'autoTranscribe') {
        backendSettings.auto_start_recording = value;
      }
      // Add more mappings as needed

      // Save to backend
      await systemService.updateUserSettings(backendSettings);
      console.log('Settings saved successfully');

    } catch (error) {
      console.error('Failed to save settings:', error);
      // Optionally show user notification here
    }
  };

  const refreshAudioDevices = async () => {
    setIsLoadingDevices(true);
    try {
      const devices = await systemService.getAudioDevices();
      setAudioDevices(devices);
    } catch (error) {
      console.error('Failed to refresh audio devices:', error);
    } finally {
      setIsLoadingDevices(false);
    }
  };

  return (
    <div className="p-8 max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-semibold mb-2">Settings</h1>
        <p className="text-muted-foreground">
          Configure your privacy, recording, and storage preferences
        </p>
      </div>

      <div className="space-y-8">
        {/* Appearance */}
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Palette className="w-5 h-5 text-indigo-500" />
              <span>Appearance</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-3">
              <Label>Theme</Label>
              <div className="grid grid-cols-3 gap-3">
                <Button
                  variant={theme === 'light' ? 'default' : 'outline'}
                  onClick={() => setTheme('light')}
                  className="rounded-xl h-16 flex flex-col items-center justify-center space-y-1"
                >
                  <Sun className="w-5 h-5" />
                  <span className="text-xs">Light</span>
                </Button>
                <Button
                  variant={theme === 'dark' ? 'default' : 'outline'}
                  onClick={() => setTheme('dark')}
                  className="rounded-xl h-16 flex flex-col items-center justify-center space-y-1"
                >
                  <Moon className="w-5 h-5" />
                  <span className="text-xs">Dark</span>
                </Button>
                <Button
                  variant={theme === 'system' ? 'default' : 'outline'}
                  onClick={() => setTheme('system')}
                  className="rounded-xl h-16 flex flex-col items-center justify-center space-y-1"
                >
                  <Monitor className="w-5 h-5" />
                  <span className="text-xs">System</span>
                </Button>
              </div>
              <p className="text-sm text-muted-foreground">
                {theme === 'system' 
                  ? 'Automatically switch between light and dark based on your system preferences'
                  : `Using ${theme} theme`
                }
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Privacy & Security */}
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Shield className="w-5 h-5 text-green-500" />
              <span>Privacy & Security</span>
              <Badge variant="secondary" className="ml-auto">Local Only</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-xl border border-green-200 dark:border-green-800">
              <div className="flex items-start space-x-3">
                <Lock className="w-5 h-5 text-green-600 mt-0.5" />
                <div>
                  <h4 className="font-medium text-green-800 dark:text-green-200">
                    Complete Privacy Protection
                  </h4>
                  <p className="text-sm text-green-700 dark:text-green-300 mt-1">
                    All your data stays on your device. Nothing is ever sent to external servers or cloud services.
                  </p>
                </div>
              </div>
            </div>

            {/* TODO: Implement data encryption and retention
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Data Encryption</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Encrypt sensitive data at rest
                  </p>
                </div>
                <Switch
                  checked={settings.encryptData}
                  onCheckedChange={(checked) => handleSettingChange('encryptData', checked)}
                />
              </div>

              <Separator />

              <div className="space-y-3">
                <Label>Data Retention Period</Label>
                <Select
                  value={settings.dataRetention}
                  onValueChange={(value) => handleSettingChange('dataRetention', value)}
                >
                  <SelectTrigger className="rounded-xl">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="30">30 days</SelectItem>
                    <SelectItem value="90">90 days</SelectItem>
                    <SelectItem value="180">6 months</SelectItem>
                    <SelectItem value="365">1 year</SelectItem>
                    <SelectItem value="never">Never delete</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-sm text-muted-foreground">
                  Automatically delete activities older than this period
                </p>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>Auto-delete old activities</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Automatically remove activities based on retention period
                  </p>
                </div>
                <Switch
                  checked={settings.autoDelete}
                  onCheckedChange={(checked) => handleSettingChange('autoDelete', checked)}
                />
              </div>
            </div>
            */}
          </CardContent>
        </Card>

        {/* Recording Settings */}
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Mic className="w-5 h-5 text-blue-500" />
              <span>Recording Settings</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label>Audio Input Device</Label>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={refreshAudioDevices}
                  disabled={isLoadingDevices}
                  className="rounded-xl"
                >
                  <RefreshCw className={`w-3 h-3 mr-1 ${isLoadingDevices ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
              </div>
              <Select
                value={settings.audioDevice}
                onValueChange={(value) => handleSettingChange('audioDevice', value)}
                disabled={isLoadingDevices}
              >
                <SelectTrigger className="rounded-xl">
                  <SelectValue placeholder={isLoadingDevices ? "Loading devices..." : "Select audio device"} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">Default System Input</SelectItem>
                  {audioDevices
                    .filter(device => device.type === 'input')
                    .map(device => (
                      <SelectItem key={device.id} value={device.id}>
                        {device.name}
                        {device.device_type && ` (${device.device_type})`}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
              {audioDevices.length === 0 && !isLoadingDevices && (
                <p className="text-sm text-muted-foreground">
                  No audio input devices found. Check your system settings.
                </p>
              )}
            </div>

            <div className="space-y-3">
              <Label>Audio Quality</Label>
              <Select
                value={settings.audioQuality}
                onValueChange={(value) => handleSettingChange('audioQuality', value)}
              >
                <SelectTrigger className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low (32 kbps)</SelectItem>
                  <SelectItem value="medium">Medium (64 kbps)</SelectItem>
                  <SelectItem value="high">High (128 kbps)</SelectItem>
                  <SelectItem value="lossless">Lossless (1411 kbps)</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-muted-foreground">
                Higher quality uses more storage space
              </p>
            </div>

            {/* TODO: Implement auto-transcribe and background recording
            <Separator />

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Auto-transcribe recordings</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Automatically generate transcripts for new recordings using Whisper
                  </p>
                </div>
                <Switch
                  checked={settings.autoTranscribe}
                  onCheckedChange={(checked) => handleSettingChange('autoTranscribe', checked)}
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>Background recording</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Continue recording when app is minimized
                  </p>
                </div>
                <Switch
                  checked={settings.backgroundRecording}
                  onCheckedChange={(checked) => handleSettingChange('backgroundRecording', checked)}
                />
              </div>
            </div>
            */}
          </CardContent>
        </Card>

        {/* TODO: Implement notification system
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Bell className="w-5 h-5 text-purple-500" />
              <span>Notifications</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label>Recording notifications</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Show notifications when recording starts/stops
                </p>
              </div>
              <Switch
                checked={settings.recordingNotifications}
                onCheckedChange={(checked) => handleSettingChange('recordingNotifications', checked)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Transcript completion</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Notify when transcript processing is complete
                </p>
              </div>
              <Switch
                checked={settings.transcriptComplete}
                onCheckedChange={(checked) => handleSettingChange('transcriptComplete', checked)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Daily summary</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Receive daily activity summaries
                </p>
              </div>
              <Switch
                checked={settings.dailySummary}
                onCheckedChange={(checked) => handleSettingChange('dailySummary', checked)}
              />
            </div>
          </CardContent>
        </Card>
        */}

        {/* Storage Management */}
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <HardDrive className="w-5 h-5 text-orange-500" />
              <span>Storage Management</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Storage Stats */}
            <div className="p-4 bg-muted/50 rounded-xl">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium">Storage Usage</h4>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={async () => {
                    try {
                      // Refresh storage stats
                      const sysInfo = await systemService.getSystemInfo();
                      const activities = await activityService.getActivities();

                      let usedStorage = 'Unknown';
                      if (sysInfo.disk_usage_bytes) {
                        usedStorage = systemService.formatStorageUsage(sysInfo.disk_usage_bytes);
                      } else if (sysInfo.disk_usage_mb) {
                        usedStorage = systemService.formatStorageUsage(sysInfo.disk_usage_mb * 1024 * 1024);
                      }

                      setStorageStats(prev => ({
                        ...prev,
                        used: usedStorage,
                        activities: activities.length
                      }));
                    } catch (error) {
                      console.error('Failed to refresh storage stats:', error);
                    }
                  }}
                  className="rounded-xl"
                >
                  <RefreshCw className="w-3 h-3 mr-1" />
                  Refresh
                </Button>
              </div>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Used Space</p>
                  <p className="font-semibold">{storageStats.used}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Available</p>
                  <p className="font-semibold">{storageStats.available}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Activities</p>
                  <p className="font-semibold">{storageStats.activities}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Audio Files</p>
                  <p className="font-semibold">{storageStats.audioFiles}</p>
                </div>
              </div>

              {/* TODO: Implement storage usage bar with real limits
              <div className="mt-4">
                <div className="flex justify-between text-sm mb-2">
                  <span>Storage Usage</span>
                  <span>{systemInfo ? 'Real-time data' : 'Estimated'}</span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                    style={{
                      width: systemInfo?.disk_usage_mb
                        ? `${Math.min((systemInfo.disk_usage_mb / (10 * 1024)) * 100, 100)}%`
                        : '24%'
                    }}
                  ></div>
                </div>
              </div>
              */}
            </div>

            <div className="space-y-3">
              <Label>Storage Location</Label>
              <Input
                value={settings.localStoragePath}
                readOnly
                className="rounded-xl bg-muted"
              />
              {/* TODO: Implement storage location change
              <div className="flex space-x-2">
                <Input
                  value={settings.localStoragePath}
                  readOnly
                  className="rounded-xl"
                />
                <Button variant="outline" className="rounded-xl">
                  Change
                </Button>
              </div>
              */}
            </div>

            {/* TODO: Implement storage limits and auto-cleanup
            <div className="space-y-3">
              <Label>Maximum Storage Size</Label>
              <Select
                value={settings.maxStorageSize}
                onValueChange={(value) => handleSettingChange('maxStorageSize', value)}
              >
                <SelectTrigger className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="5">5 GB</SelectItem>
                  <SelectItem value="10">10 GB</SelectItem>
                  <SelectItem value="25">25 GB</SelectItem>
                  <SelectItem value="50">50 GB</SelectItem>
                  <SelectItem value="unlimited">Unlimited</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Auto-cleanup storage</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Automatically manage storage when approaching limit
                </p>
              </div>
              <Switch
                checked={settings.autoCleanup}
                onCheckedChange={(checked) => handleSettingChange('autoCleanup', checked)}
              />
            </div>

            <Separator />

            <div className="flex space-x-3">
              <Button variant="outline" className="rounded-xl">
                <Download className="w-4 h-4 mr-2" />
                Export Data
              </Button>
              <Button
                variant="outline"
                className="rounded-xl"
                onClick={async () => {
                  try {
                    const activities = await activityService.getActivities();
                    const thirtyDaysAgo = new Date();
                    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

                    const oldActivities = activities.filter(activity =>
                      activity.start_time && new Date(activity.start_time) < thirtyDaysAgo
                    );

                    console.log(`Found ${oldActivities.length} activities older than 30 days`);
                  } catch (error) {
                    console.error('Failed to analyze activities for cleanup:', error);
                  }
                }}
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Clean Old Data
              </Button>
              <Button variant="destructive" className="rounded-xl">
                <Trash2 className="w-4 h-4 mr-2" />
                Clear All Data
              </Button>
            </div>
            */}
          </CardContent>
        </Card>

        {/* TODO: Implement advanced settings
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Zap className="w-5 h-5 text-yellow-500" />
              <span>Advanced</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label>Debug Mode</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Enable detailed logging for troubleshooting
                </p>
              </div>
              <Switch />
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Hardware Acceleration</Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Use GPU for faster transcription processing
                </p>
              </div>
              <Switch defaultChecked />
            </div>

            <Separator />

            <div className="flex space-x-3">
              <Button variant="outline" className="rounded-xl">
                <Database className="w-4 h-4 mr-2" />
                Rebuild Database
              </Button>
              <Button variant="outline" className="rounded-xl">
                <SettingsIcon className="w-4 h-4 mr-2" />
                Reset to Defaults
              </Button>
            </div>
          </CardContent>
        </Card>
        */}
      </div>
    </div>
  );
}