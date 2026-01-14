import React, { useState } from 'react'
import { CreateActivity } from '../../../wailsjs/go/main/App'
import Button from '../ui/Button'

const CreateActivityForm = ({ onActivityCreated, onCancel }) => {
  const [formData, setFormData] = useState({
    title: '',
    type: 'meeting',
    description: ''
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState(null)

  const activityTypes = [
    { value: 'meeting', label: 'Meeting' },
    { value: 'work_session', label: 'Work Session' },
    { value: 'call', label: 'Phone Call' },
    { value: 'interview', label: 'Interview' },
    { value: 'presentation', label: 'Presentation' },
    { value: 'other', label: 'Other' }
  ]

  const handleInputChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!formData.title.trim()) {
      setError('Activity title is required')
      return
    }

    try {
      setIsSubmitting(true)
      setError(null)
      
      const activity = await CreateActivity(formData.type, formData.title.trim())
      
      if (activity) {
        onActivityCreated(activity)
      } else {
        throw new Error('Failed to create activity')
      }
    } catch (err) {
      console.error('Failed to create activity:', err)
      setError(err.message || 'Failed to create activity')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="create-activity-form">
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="title">Activity Title *</label>
          <input
            type="text"
            id="title"
            name="title"
            value={formData.title}
            onChange={handleInputChange}
            placeholder="Enter activity title"
            disabled={isSubmitting}
            className="form-input"
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="type">Activity Type</label>
          <select
            id="type"
            name="type"
            value={formData.type}
            onChange={handleInputChange}
            disabled={isSubmitting}
            className="form-select"
          >
            {activityTypes.map(type => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="description">Description (optional)</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            placeholder="Enter activity description"
            disabled={isSubmitting}
            className="form-textarea"
            rows="3"
          />
        </div>

        {error && (
          <div className="form-error">
            <p className="error-message">{error}</p>
          </div>
        )}

        <div className="form-actions">
          <Button
            type="button"
            variant="secondary"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={isSubmitting || !formData.title.trim()}
          >
            {isSubmitting ? 'Creating...' : 'Create Activity'}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default CreateActivityForm