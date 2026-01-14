import React, { useState } from 'react'
import { StartActivity, StopActivity, DeleteActivity, ProcessActivityTranscription, GetTranscriptionStatus } from '../../../wailsjs/go/main/App'
import StatusIndicator from '../ui/StatusIndicator'

const ActivityCard = ({ activity, onActivityDeleted }) => {
  const [isStarting, setIsStarting] = useState(false)
  const [isStopping, setIsStopping] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showDetails, setShowDetails] = useState(false)
  const [showConfirmDelete, setShowConfirmDelete] = useState(false)
  const [isTranscribing, setIsTranscribing] = useState(false)
  const [transcriptionStatus, setTranscriptionStatus] = useState(null)

  const handleStartActivity = async () => {
    try {
      setIsStarting(true)
      await StartActivity(activity.id)
      // Refresh activity data
      window.location.reload() // Simple approach for now
    } catch (error) {
      console.error('Failed to start activity:', error)
      alert('Failed to start activity: ' + error.message)
    } finally {
      setIsStarting(false)
    }
  }

  const handleStopActivity = async () => {
    try {
      setIsStopping(true)
      await StopActivity(activity.id)
      // Refresh activity data
      window.location.reload() // Simple approach for now
    } catch (error) {
      console.error('Failed to stop activity:', error)
      alert('Failed to stop activity: ' + error.message)
    } finally {
      setIsStopping(false)
    }
  }

  const handleDeleteActivity = async () => {
    try {
      setIsDeleting(true)
      console.log(`Frontend: Starting delete for activity ${activity.id} (${activity.title})`)

      await DeleteActivity(activity.id)

      console.log(`Frontend: Successfully deleted activity ${activity.id}`)

      // Notify parent component if callback provided
      if (onActivityDeleted) {
        onActivityDeleted(activity.id)
      } else {
        // Fallback to page reload
        window.location.reload()
      }
    } catch (error) {
      console.error('Frontend: Failed to delete activity:', error)
      alert('Failed to delete activity: ' + error.message)
    } finally {
      setIsDeleting(false)
      setShowConfirmDelete(false)
    }
  }

  const handleTranscribeActivity = async () => {
    try {
      setIsTranscribing(true)
      setTranscriptionStatus({ stage: 'processing', progress: 0 })

      console.log(`Starting transcription for activity ${activity.id}`)
      await ProcessActivityTranscription(activity.id)

      // Poll for transcription status
      const pollStatus = async () => {
        try {
          const status = await GetTranscriptionStatus(activity.id)
          setTranscriptionStatus(status)

          if (status.stage === 'completed') {
            console.log(`Transcription completed for activity ${activity.id}`)
            setIsTranscribing(false)
            // Trigger refresh after successful transcription
            setTimeout(() => {
              window.location.reload()
            }, 1000)
          } else if (status.stage === 'failed') {
            console.error(`Transcription failed for activity ${activity.id}:`, status.last_error)
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

  const canDeleteActivity = (status) => {
    // Allow deletion of all activities (user requested this for stuck recordings)
    return true
  }

  const isActiveOrRecording = (status) => {
    return status === 'active' || status === 'recording'
  }

  const canTranscribeActivity = (status) => {
    // Only allow transcription for completed activities
    return status === 'completed'
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

  return (
    <div className={`activity-card ${activity.status}`}>
      <div className="activity-card-header">
        <div className="activity-title">
          <h3>{activity.title}</h3>
          <span className="activity-type">{activity.type}</span>
        </div>
        <StatusIndicator
          status={activity.status}
          color={getStatusColor(activity.status)}
        />
      </div>

      <div className="activity-card-details">
        <div className="activity-meta">
          <span className="activity-date">
            Started: {formatDate(activity.start_time)}
          </span>
          {activity.end_time && (
            <span className="activity-date">
              Ended: {formatDate(activity.end_time)}
            </span>
          )}
          <span className="activity-duration">
            Duration: {formatDuration(activity.start_time, activity.end_time)}
          </span>
        </div>

        {activity.description && (
          <p className="activity-description">{activity.description}</p>
        )}
      </div>

      <div className="activity-card-actions">
        {activity.status === 'active' ? (
          <button
            onClick={handleStopActivity}
            disabled={isStopping}
            className="btn btn-danger"
          >
            {isStopping ? 'Stopping...' : 'Stop Activity'}
          </button>
        ) : activity.status === 'created' ? (
          <button
            onClick={handleStartActivity}
            disabled={isStarting}
            className="btn btn-primary"
          >
            {isStarting ? 'Starting...' : 'Start Activity'}
          </button>
        ) : null}

        {canTranscribeActivity(activity.status) && (
          <button
            onClick={handleTranscribeActivity}
            disabled={isTranscribing}
            className="btn btn-secondary"
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
              'Transcribe'
            )}
          </button>
        )}

        <button
          onClick={() => setShowDetails(!showDetails)}
          className="btn btn-action"
        >
          {showDetails ? 'Hide' : 'View'} Details
        </button>

        <button
          onClick={() => {
            console.log(`Frontend: Delete button clicked for activity ${activity.id} (${activity.title})`)
            if (canDeleteActivity(activity.status)) {
              setShowConfirmDelete(true)
            }
          }}
          disabled={isDeleting || !canDeleteActivity(activity.status)}
          className={`btn-icon ${canDeleteActivity(activity.status) ? 'btn-danger' : 'btn-disabled'}`}
          title={canDeleteActivity(activity.status) ? 'Delete this activity and all associated files' : `Cannot delete ${activity.status} activity`}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
          </svg>
          {isDeleting && <span style={{ marginLeft: '4px' }}>Deleting...</span>}
        </button>
      </div>

      {showDetails && (
        <div className="activity-card-expanded">
          <div className="activity-details">
            <h4>Activity Details</h4>
            <div className="detail-row">
              <span className="detail-label">ID:</span>
              <span className="detail-value">{activity.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Created:</span>
              <span className="detail-value">{formatDate(activity.created_at)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Updated:</span>
              <span className="detail-value">{formatDate(activity.updated_at)}</span>
            </div>
            {activity.metadata && Object.keys(activity.metadata).length > 0 && (
              <div className="detail-row">
                <span className="detail-label">Metadata:</span>
                <pre className="detail-value">{JSON.stringify(activity.metadata, null, 2)}</pre>
              </div>
            )}
          </div>
        </div>
      )}

      {showConfirmDelete && (
        <div className="activity-card-expanded">
          <div className="delete-confirmation">
            <div className="confirmation-content">
              <p>Are you sure you want to delete "{activity.title}"?</p>
              <p className="warning-text">
                This will permanently delete the activity and all audio files. This cannot be undone.
              </p>
              <div className="confirmation-actions">
                <button
                  onClick={() => {
                    console.log(`Frontend: Delete cancelled for activity ${activity.id}`)
                    setShowConfirmDelete(false)
                  }}
                  className="btn btn-secondary"
                  disabled={isDeleting}
                >
                  Cancel
                </button>
                <button
                  onClick={() => {
                    console.log(`Frontend: Delete confirmed for activity ${activity.id}`)
                    handleDeleteActivity()
                  }}
                  className="btn btn-danger"
                  disabled={isDeleting}
                >
                  {isDeleting ? 'Deleting...' : 'Yes'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ActivityCard