import React, { useState, useEffect } from 'react'
import { StopRecordingButtonAction } from '../../../wailsjs/go/main/App'
import StatusIndicator from '../ui/StatusIndicator'

const RecordingStatus = ({ recording, onStop }) => {
  const [duration, setDuration] = useState(0)
  const [intervalId, setIntervalId] = useState(null)
  const [isStopping, setIsStopping] = useState(false)

  useEffect(() => {
    if (recording) {
      startTimer()
    } else {
      stopTimer()
    }

    return () => stopTimer()
  }, [recording])

  const startTimer = () => {
    const startTime = recording?.start_time ? new Date(recording.start_time) : new Date()
    
    const id = setInterval(() => {
      const now = new Date()
      const elapsed = Math.floor((now - startTime) / 1000)
      setDuration(elapsed)
    }, 1000)
    
    setIntervalId(id)
  }

  const stopTimer = () => {
    if (intervalId) {
      clearInterval(intervalId)
      setIntervalId(null)
    }
  }

  const formatDuration = (seconds) => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60

    if (hours > 0) {
      return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  const handleStopRecording = async () => {
    if (!recording) return

    try {
      setIsStopping(true)

      await StopRecordingButtonAction(recording.id)

      // Always call the callback to clear recording state
      if (onStop) {
        onStop()
      }
    } catch (err) {
      console.error('Failed to stop recording:', err)

      // Even if stopping fails, clear the recording state to prevent UI from staying stuck
      if (onStop) {
        console.warn('Clearing recording state despite stop failure to prevent UI lock')
        onStop()
      }
    } finally {
      setIsStopping(false)
    }
  }

  const getFileSize = () => {
    if (!recording || !recording.file_size) return 'Unknown'
    
    const bytes = recording.file_size
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  if (!recording) {
    return null
  }

  return (
    <div className="recording-status">
      <div className="recording-status-header">
        <div className="status-info">
          <StatusIndicator status="recording" color="red" />
          <span className="recording-label">Recording Active</span>
        </div>
        <button 
          onClick={handleStopRecording}
          disabled={isStopping}
          className="btn btn-danger stop-recording-btn"
        >
          {isStopping ? (
            <>
              <span className="btn-spinner"></span>
              Stopping...
            </>
          ) : (
            <>
              ⏹ Stop Recording
            </>
          )}
        </button>
      </div>

      <div className="recording-details">
        <div className="recording-timer">
          <span className="timer-label">Duration:</span>
          <span className="timer-value">{formatDuration(duration)}</span>
        </div>
        
        <div className="recording-info-grid">
          <div className="info-item">
            <span className="info-label">File:</span>
            <span className="info-value">{recording.filename || 'recording.m4a'}</span>
          </div>
          
          <div className="info-item">
            <span className="info-label">Quality:</span>
            <span className="info-value">
              {recording.sample_rate}Hz, {recording.bit_depth}bit
            </span>
          </div>
          
          <div className="info-item">
            <span className="info-label">Device:</span>
            <span className="info-value">{recording.device_name || 'Default'}</span>
          </div>
          
          {recording.file_size && (
            <div className="info-item">
              <span className="info-label">Size:</span>
              <span className="info-value">{getFileSize()}</span>
            </div>
          )}
        </div>
      </div>

      <div className="recording-progress">
        <div className="progress-bar">
          <div className="progress-indicator"></div>
        </div>
        <p className="progress-text">Recording in progress...</p>
      </div>
    </div>
  )
}

export default RecordingStatus