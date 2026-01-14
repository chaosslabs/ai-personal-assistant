import React, { useState, useEffect } from 'react'
import { GetActivities } from '../../../wailsjs/go/main/App'
import ActivityCard from './ActivityCard'
import LoadingSpinner from '../ui/LoadingSpinner'

const ActivityList = ({ limit = null, refreshTrigger = 0 }) => {
  const [activities, setActivities] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    loadActivities()
  }, [])

  // Refresh activities when refreshTrigger changes
  useEffect(() => {
    if (refreshTrigger > 0) {
      loadActivities()
    }
  }, [refreshTrigger])

  const loadActivities = async () => {
    try {
      setLoading(true)
      setError(null)
      console.log('Frontend: Loading activities...')
      
      const result = await GetActivities()
      console.log('Frontend: GetActivities result:', result)
      
      let activitiesList = result || []
      console.log('Frontend: Activities list length:', activitiesList.length)
      
      // Apply limit if specified
      if (limit && activitiesList.length > limit) {
        activitiesList = activitiesList.slice(0, limit)
        console.log('Frontend: Applied limit, new length:', activitiesList.length)
      }
      
      setActivities(activitiesList)
      console.log('Frontend: Activities loaded successfully')
    } catch (err) {
      console.error('Frontend: Failed to load activities:', err)
      console.error('Frontend: Error details:', {
        name: err.name,
        message: err.message,
        stack: err.stack
      })
      setError(`Failed to load activities: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleActivityDeleted = (deletedActivityId) => {
    console.log(`Frontend: Activity ${deletedActivityId} deleted, refreshing activities list`)
    // Refresh the activities list from the database to ensure consistency
    loadActivities()
  }

  const filteredActivities = activities.filter(activity =>
    activity.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    activity.type.toLowerCase().includes(searchTerm.toLowerCase())
  )

  if (loading) {
    return (
      <div className="activity-list-loading">
        <LoadingSpinner />
        <p>Loading activities...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="activity-list-error">
        <p className="error-message">{error}</p>
        <button onClick={loadActivities} className="btn btn-refresh">
          Try Again
        </button>
      </div>
    )
  }

  return (
    <div className="activity-list">
      {!limit && (
        <div className="activity-list-header">
          <div className="search-box">
            <input
              type="text"
              placeholder="Search activities..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="search-input"
            />
          </div>
          <button onClick={loadActivities} className="btn btn-refresh">
            🔄 Refresh
          </button>
        </div>
      )}

      {filteredActivities.length === 0 ? (
        <div className="activity-list-empty">
          {searchTerm ? (
            <p>No activities found matching "{searchTerm}"</p>
          ) : (
            <div>
              <p>No activities yet</p>
              <p className="text-muted">Start by creating your first activity</p>
            </div>
          )}
        </div>
      ) : (
        <div className="activity-grid">
          {filteredActivities.map(activity => (
            <ActivityCard 
              key={activity.id} 
              activity={activity} 
              onActivityDeleted={handleActivityDeleted}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default ActivityList