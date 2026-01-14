import React, { useState, useEffect } from 'react'
import { StartRecordingButtonAction, StopRecordingButtonAction, GetAudioDevices } from '../../../wailsjs/go/main/App'
import AudioDeviceSelector from './AudioDeviceSelector'
import Button from '../ui/Button'
import StatusIndicator from '../ui/StatusIndicator'

const RecordingControls = ({ onRecordingStart, onRecordingStop, activeRecording }) => {
  const [audioDevices, setAudioDevices] = useState([])
  const [selectedDevice, setSelectedDevice] = useState('default')
  const [isStarting, setIsStarting] = useState(false)
  const [isStopping, setIsStopping] = useState(false)
  const [showDeviceSelector, setShowDeviceSelector] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadAudioDevices()
  }, [])

  const loadAudioDevices = async () => {
    try {
      const devices = await GetAudioDevices()
      setAudioDevices(devices || [])
      
      // If no device is selected yet and we have devices, select the default one
      if (selectedDevice === 'default' && devices && devices.length > 0) {
        const defaultDevice = devices.find(device => device.default)
        if (defaultDevice) {
          setSelectedDevice(defaultDevice.id)
        } else {
          // If no device is marked as default, use the first one
          setSelectedDevice(devices[0].id)
        }
      }
    } catch (err) {
      console.error('Failed to load audio devices:', err)
    }
  }

  const handleStartRecording = async () => {
    try {
      setIsStarting(true)
      setError(null)

      const session = await StartRecordingButtonAction()

      if (session && onRecordingStart) {
        onRecordingStart(session.audio_recording)
      }
    } catch (err) {
      console.error('Failed to start recording:', err)
      setError(err.message || 'Failed to start recording')
    } finally {
      setIsStarting(false)
    }
  }

  const handleStopRecording = async () => {
    if (!activeRecording) return

    try {
      setIsStopping(true)
      setError(null)

      await StopRecordingButtonAction(activeRecording.id)

      // Always call the callback to clear recording state, even if backend succeeds
      if (onRecordingStop) {
        onRecordingStop()
      }
    } catch (err) {
      console.error('Failed to stop recording:', err)
      setError(err.message || 'Failed to stop recording')

      // Even if stopping fails, clear the recording state to prevent UI from staying stuck
      // The user can see the error and try again if needed
      if (onRecordingStop) {
        console.warn('Clearing recording state despite stop failure to prevent UI lock')
        onRecordingStop()
      }
    } finally {
      setIsStopping(false)
    }
  }

  const isRecording = !!activeRecording

  return (
    <div className="recording-controls">
      <div className="recording-controls-header">
        <h3>Recording Controls</h3>
        <StatusIndicator
          status={isRecording ? 'recording' : 'ready'}
          color={isRecording ? 'red' : 'green'}
        />
      </div>

      {error && (
        <div className="recording-error">
          <p className="error-message">{error}</p>
        </div>
      )}

      <div className="recording-settings">
        <div className="device-settings">
          <button
            onClick={() => setShowDeviceSelector(!showDeviceSelector)}
            className="btn btn-action"
            disabled={isRecording}
          >
            Audio Settings
          </button>
          {showDeviceSelector && (
            <AudioDeviceSelector
              devices={audioDevices}
              selectedDevice={selectedDevice}
              onDeviceSelect={setSelectedDevice}
              onClose={() => setShowDeviceSelector(false)}
            />
          )}
        </div>
      </div>

      <div className="recording-actions">
        {!isRecording ? (
          <Button
            onClick={handleStartRecording}
            variant="primary"
            size="large"
            disabled={isStarting}
            className="recording-start-btn"
          >
            {isStarting ? (
              <>
                <span className="btn-spinner"></span>
                Starting...
              </>
            ) : (
              <>
                <span className="record-icon">⏺</span>
                Start Recording
              </>
            )}
          </Button>
        ) : (
          <Button
            onClick={handleStopRecording}
            variant="danger"
            size="large"
            disabled={isStopping}
            className="recording-stop-btn"
          >
            {isStopping ? (
              <>
                <span className="btn-spinner"></span>
                Stopping...
              </>
            ) : (
              <>
                <span className="stop-icon">⏹</span>
                Stop Recording
              </>
            )}
          </Button>
        )}
      </div>

      {isRecording && (
        <div className="recording-info">
          <p className="text-muted">
            Recording Manual Session
          </p>
          <p className="text-muted">
            Device: {selectedDevice}
          </p>
        </div>
      )}
    </div>
  )
}

export default RecordingControls