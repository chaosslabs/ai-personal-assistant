import React, { useState, useEffect } from 'react'
import { GetActivity, GetTranscript, ProcessActivityTranscription, GetTranscriptionStatus } from '../../../wailsjs/go/main/App'
import TranscriptViewer from '../Transcript/TranscriptViewer'
import LoadingSpinner from '../ui/LoadingSpinner'
import StatusIndicator from '../ui/StatusIndicator'

const ActivityDetail = ({ activityId, onClose }) => {
  const [activity, setActivity] = useState(null)
  const [transcript, setTranscript] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [activeTab, setActiveTab] = useState('overview')
  const [isTranscribing, setIsTranscribing] = useState(false)
  const [transcriptionStatus, setTranscriptionStatus] = useState(null)

  useEffect(() => {
    if (activityId) {
      loadActivityDetails()
    }
  }, [activityId])

  const loadActivityDetails = async () => {
    try {
      setLoading(true)
      setError(null)

      const [activityResult, transcriptResult] = await Promise.all([
        GetActivity(activityId).catch(() => null),
        GetTranscript(activityId).catch(() => [])
      ])

      if (activityResult) {
        setActivity(activityResult)
      } else {
        setError('Activity not found')
      }

      setTranscript(transcriptResult || [])
    } catch (err) {
      console.error('Failed to load activity details:', err)
      setError('Failed to load activity details')
    } finally {
      setLoading(false)
    }
  }

  const handleTranscribeActivity = async () => {
    try {
      setIsTranscribing(true)
      setTranscriptionStatus({ stage: 'processing', progress: 0 })

      console.log(`Starting transcription for activity ${activityId}`)
      await ProcessActivityTranscription(activityId)

      // Poll for transcription status
      const pollStatus = async () => {
        try {
          const status = await GetTranscriptionStatus(activityId)
          setTranscriptionStatus(status)

          if (status.stage === 'completed') {
            console.log(`Transcription completed for activity ${activityId}`)
            setIsTranscribing(false)
            // Reload transcript data
            const updatedTranscript = await GetTranscript(activityId)
            setTranscript(updatedTranscript || [])
          } else if (status.stage === 'failed') {
            console.error(`Transcription failed for activity ${activityId}:`, status.last_error)
            setIsTranscribing(false)
            alert(`Transcription failed: ${status.last_error}`)
          } else if (status.stage === 'processing') {
            // Continue polling
            setTimeout(pollStatus, 2000)
          }
        } catch (error) {
          console.error('Failed to get transcription status:', error)
          setIsTranscribing(false)
          setTranscriptionStatus(null)
        }
      }

      // Start polling after a short delay
      setTimeout(pollStatus, 1000)

    } catch (error) {
      console.error('Failed to start transcription:', error)
      alert('Failed to start transcription: ' + error.message)
      setIsTranscribing(false)
      setTranscriptionStatus(null)
    }
  }

  const canTranscribeActivity = (status) => {
    // Only allow transcription for completed activities
    return status === 'completed'
  }

  const formatDate = (dateString) => {
    try {
      const date = new Date(dateString)
      return date.toLocaleString()
    } catch {
      return dateString
    }
  }

  const formatDuration = (startTime, endTime) => {
    try {
      const start = new Date(startTime)
      const end = endTime ? new Date(endTime) : new Date()
      const durationMs = end - start
      const minutes = Math.floor(durationMs / (1000 * 60))
      const hours = Math.floor(minutes / 60)
      
      if (hours > 0) {
        return `${hours}h ${minutes % 60}m`
      }
      return `${minutes}m`
    } catch {
      return 'Unknown'
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'green'
      case 'completed':
        return 'blue'
      case 'paused':
        return 'yellow'
      default:
        return 'gray'
    }
  }

  if (loading) {
    return (
      <div className="activity-detail-loading">
        <LoadingSpinner />
        <p>Loading activity details...</p>
      </div>
    )
  }

  if (error || !activity) {
    return (
      <div className="activity-detail-error">
        <p className="error-message">{error || 'Activity not found'}</p>
        <button onClick={onClose} className="btn-secondary">
          Close
        </button>
      </div>
    )
  }

  return (
    <div className="activity-detail">
      <header className="activity-detail-header">
        <div className="activity-title-section">
          <h1>{activity.title}</h1>
          <div className="activity-meta">
            <span className="activity-type">{activity.type}</span>
            <StatusIndicator
              status={activity.status}
              color={getStatusColor(activity.status)}
            />
          </div>
        </div>
        <div className="activity-detail-actions">
          {canTranscribeActivity(activity.status) && (
            <button
              onClick={handleTranscribeActivity}
              disabled={isTranscribing}
              className="btn btn-primary"
              title="Transcribe this activity's audio recordings using local Whisper"
            >
              {isTranscribing ? (
                <>
                  {transcriptionStatus?.progress ?
                    `Transcribing... ${Math.round(transcriptionStatus.progress * 100)}%` :
                    'Transcribing...'
                  }
                </>
              ) : (
                'Transcribe Activity'
              )}
            </button>
          )}
          <button onClick={onClose} className="btn-secondary">
            Close
          </button>
        </div>
      </header>

      <nav className="activity-detail-nav">
        <button 
          className={activeTab === 'overview' ? 'active' : ''}
          onClick={() => setActiveTab('overview')}
        >
          Overview
        </button>
        <button 
          className={activeTab === 'transcript' ? 'active' : ''}
          onClick={() => setActiveTab('transcript')}
        >
          Transcript ({transcript.length})
        </button>
        <button 
          className={activeTab === 'recordings' ? 'active' : ''}
          onClick={() => setActiveTab('recordings')}
        >
          Recordings
        </button>
      </nav>

      <div className="activity-detail-content">
        {activeTab === 'overview' && (
          <div className="activity-overview">
            <div className="overview-section">
              <h3>Activity Information</h3>
              <div className="info-grid">
                <div className="info-item">
                  <span className="info-label">Started:</span>
                  <span className="info-value">{formatDate(activity.start_time)}</span>
                </div>
                {activity.end_time && (
                  <div className="info-item">
                    <span className="info-label">Ended:</span>
                    <span className="info-value">{formatDate(activity.end_time)}</span>
                  </div>
                )}
                <div className="info-item">
                  <span className="info-label">Duration:</span>
                  <span className="info-value">
                    {formatDuration(activity.start_time, activity.end_time)}
                  </span>
                </div>
                <div className="info-item">
                  <span className="info-label">Created:</span>
                  <span className="info-value">{formatDate(activity.created_at)}</span>
                </div>
                <div className="info-item">
                  <span className="info-label">Updated:</span>
                  <span className="info-value">{formatDate(activity.updated_at)}</span>
                </div>
              </div>
            </div>

            {activity.description && (
              <div className="overview-section">
                <h3>Description</h3>
                <p className="activity-description">{activity.description}</p>
              </div>
            )}

            {activity.metadata && Object.keys(activity.metadata).length > 0 && (
              <div className="overview-section">
                <h3>Additional Information</h3>
                <pre className="metadata-display">
                  {JSON.stringify(activity.metadata, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}

        {activeTab === 'transcript' && (
          <div className="activity-transcript">
            {isTranscribing && (
              <div className="transcription-progress">
                <h3>Transcription in Progress</h3>
                {transcriptionStatus && (
                  <div className="progress-info">
                    <p>Status: {transcriptionStatus.stage}</p>
                    {transcriptionStatus.progress > 0 && (
                      <div className="progress-bar">
                        <div
                          className="progress-fill"
                          style={{ width: `${transcriptionStatus.progress * 100}%` }}
                        ></div>
                      </div>
                    )}
                    {transcriptionStatus.current_file && (
                      <p className="text-muted">Processing: {transcriptionStatus.current_file}</p>
                    )}
                  </div>
                )}
              </div>
            )}

            {transcript.length > 0 ? (
              <TranscriptViewer chunks={transcript} />
            ) : !isTranscribing ? (
              <div className="empty-state">
                <p>No transcript available</p>
                {canTranscribeActivity(activity.status) ? (
                  <p className="text-muted">
                    Click "Transcribe Activity" to generate a transcript using local Whisper processing
                  </p>
                ) : (
                  <p className="text-muted">
                    Transcripts will appear here after audio is recorded and processed
                  </p>
                )}
              </div>
            ) : null}
          </div>
        )}

        {activeTab === 'recordings' && (
          <div className="activity-recordings">
            <div className="empty-state">
              <p>No recordings available</p>
              <p className="text-muted">
                Audio recordings will appear here when available
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default ActivityDetail