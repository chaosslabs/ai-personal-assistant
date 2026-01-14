import { useState, useEffect, useCallback } from 'react'
import { 
  GetActivities, 
  GetActivity, 
  CreateActivity, 
  StartActivity, 
  StopActivity 
} from '../../wailsjs/go/main/App'

export const useActivities = () => {
  const [activities, setActivities] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadActivities = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await GetActivities()
      setActivities(result || [])
    } catch (err) {
      console.error('Failed to load activities:', err)
      setError(err.message || 'Failed to load activities')
    } finally {
      setLoading(false)
    }
  }, [])

  const createActivity = useCallback(async (type, title) => {
    try {
      const activity = await CreateActivity(type, title)
      setActivities(prev => [activity, ...prev])
      return activity
    } catch (err) {
      console.error('Failed to create activity:', err)
      throw err
    }
  }, [])

  const startActivity = useCallback(async (activityId) => {
    try {
      await StartActivity(activityId)
      await loadActivities() // Refresh to get updated status
    } catch (err) {
      console.error('Failed to start activity:', err)
      throw err
    }
  }, [loadActivities])

  const stopActivity = useCallback(async (activityId) => {
    try {
      await StopActivity(activityId)
      await loadActivities() // Refresh to get updated status
    } catch (err) {
      console.error('Failed to stop activity:', err)
      throw err
    }
  }, [loadActivities])

  const refreshActivities = useCallback(() => {
    loadActivities()
  }, [loadActivities])

  useEffect(() => {
    loadActivities()
  }, [loadActivities])

  return {
    activities,
    loading,
    error,
    createActivity,
    startActivity,
    stopActivity,
    refreshActivities
  }
}

export const useActivity = (activityId) => {
  const [activity, setActivity] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadActivity = useCallback(async (id) => {
    if (!id) {
      setActivity(null)
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await GetActivity(id)
      setActivity(result)
    } catch (err) {
      console.error('Failed to load activity:', err)
      setError(err.message || 'Failed to load activity')
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshActivity = useCallback(() => {
    if (activityId) {
      loadActivity(activityId)
    }
  }, [activityId, loadActivity])

  useEffect(() => {
    loadActivity(activityId)
  }, [activityId, loadActivity])

  return {
    activity,
    loading,
    error,
    refreshActivity
  }
}