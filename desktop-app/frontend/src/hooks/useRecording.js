import { useState, useEffect, useCallback } from 'react'
import { 
  StartRecording, 
  StopRecording, 
  GetActiveRecording, 
  GetRecording,
  GetAudioDevices
} from '../../wailsjs/go/main/App'

export const useRecording = () => {
  const [activeRecording, setActiveRecording] = useState(null)
  const [isRecording, setIsRecording] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const checkActiveRecording = useCallback(async () => {
    try {
      const recording = await GetActiveRecording()
      setActiveRecording(recording)
      setIsRecording(!!recording)
    } catch (err) {
      console.error('Failed to check active recording:', err)
      setActiveRecording(null)
      setIsRecording(false)
    }
  }, [])

  const startRecording = useCallback(async (activityId) => {
    try {
      setLoading(true)
      setError(null)
      const recording = await StartRecording(activityId)
      setActiveRecording(recording)
      setIsRecording(true)
      return recording
    } catch (err) {
      console.error('Failed to start recording:', err)
      setError(err.message || 'Failed to start recording')
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  const stopRecording = useCallback(async () => {
    if (!activeRecording) return

    try {
      setLoading(true)
      setError(null)
      await StopRecording(activeRecording.id)
      setActiveRecording(null)
      setIsRecording(false)
    } catch (err) {
      console.error('Failed to stop recording:', err)
      setError(err.message || 'Failed to stop recording')
      throw err
    } finally {
      setLoading(false)
    }
  }, [activeRecording])

  useEffect(() => {
    checkActiveRecording()
  }, [checkActiveRecording])

  return {
    activeRecording,
    isRecording,
    loading,
    error,
    startRecording,
    stopRecording,
    refreshRecording: checkActiveRecording
  }
}

export const useAudioDevices = () => {
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadDevices = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const deviceList = await GetAudioDevices()
      setDevices(deviceList || [])
    } catch (err) {
      console.error('Failed to load audio devices:', err)
      setError(err.message || 'Failed to load audio devices')
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshDevices = useCallback(() => {
    loadDevices()
  }, [loadDevices])

  useEffect(() => {
    loadDevices()
  }, [loadDevices])

  return {
    devices,
    loading,
    error,
    refreshDevices
  }
}

export const useRecordingDetails = (recordingId) => {
  const [recording, setRecording] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadRecording = useCallback(async (id) => {
    if (!id) {
      setRecording(null)
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await GetRecording(id)
      setRecording(result)
    } catch (err) {
      console.error('Failed to load recording:', err)
      setError(err.message || 'Failed to load recording')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadRecording(recordingId)
  }, [recordingId, loadRecording])

  return {
    recording,
    loading,
    error,
    refreshRecording: () => loadRecording(recordingId)
  }
}