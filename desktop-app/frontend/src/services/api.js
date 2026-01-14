// Centralized API service layer for Wails backend communication
import { 
  GetVersion,
  GetCurrentUser,
  GetAppStatus,
  GetAppStatusDetailed,
  GetSystemInfo,
  
  // Activity Management
  CreateActivity,
  GetActivities,
  GetActivity,
  StartActivity,
  StopActivity,
  
  // Recording Management
  StartRecording,
  StopRecording,
  GetActiveRecording,
  GetRecording,
  GetAudioDevices,
  
  // Transcript Management
  GetTranscript,
  SearchTranscripts
} from '../../wailsjs/go/main/App'

// Error handling wrapper
const handleApiCall = async (apiCall, errorMessage = 'API call failed') => {
  try {
    return await apiCall()
  } catch (error) {
    console.error(errorMessage + ':', error)
    throw new Error(error.message || errorMessage)
  }
}

// System API
export const systemApi = {
  getVersion: () => handleApiCall(GetVersion, 'Failed to get version'),
  getCurrentUser: () => handleApiCall(GetCurrentUser, 'Failed to get current user'),
  getAppStatus: () => handleApiCall(GetAppStatus, 'Failed to get app status'),
  getAppStatusDetailed: () => handleApiCall(GetAppStatusDetailed, 'Failed to get detailed app status'),
  getSystemInfo: () => handleApiCall(GetSystemInfo, 'Failed to get system info')
}

// Activity API
export const activityApi = {
  create: (type, title) => 
    handleApiCall(() => CreateActivity(type, title), 'Failed to create activity'),
  
  getAll: () => 
    handleApiCall(GetActivities, 'Failed to get activities'),
  
  getById: (activityId) => 
    handleApiCall(() => GetActivity(activityId), 'Failed to get activity'),
  
  start: (activityId) => 
    handleApiCall(() => StartActivity(activityId), 'Failed to start activity'),
  
  stop: (activityId) => 
    handleApiCall(() => StopActivity(activityId), 'Failed to stop activity')
}

// Recording API
export const recordingApi = {
  start: (activityId) => 
    handleApiCall(() => StartRecording(activityId), 'Failed to start recording'),
  
  stop: (recordingId) => 
    handleApiCall(() => StopRecording(recordingId), 'Failed to stop recording'),
  
  getActive: () => 
    handleApiCall(GetActiveRecording, 'Failed to get active recording'),
  
  getById: (recordingId) => 
    handleApiCall(() => GetRecording(recordingId), 'Failed to get recording'),
  
  getAudioDevices: () => 
    handleApiCall(GetAudioDevices, 'Failed to get audio devices')
}

// Transcript API
export const transcriptApi = {
  getForActivity: (activityId) => 
    handleApiCall(() => GetTranscript(activityId), 'Failed to get transcript'),
  
  search: (query) => 
    handleApiCall(() => SearchTranscripts(query), 'Failed to search transcripts')
}

// Combined API object
const api = {
  system: systemApi,
  activity: activityApi,
  recording: recordingApi,
  transcript: transcriptApi
}

export default api