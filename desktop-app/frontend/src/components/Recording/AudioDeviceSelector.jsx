import React, { useState, useEffect } from 'react'
import { GetAudioDevices } from '../../../wailsjs/go/main/App'
import Modal from '../ui/Modal'
import Button from '../ui/Button'
import LoadingSpinner from '../ui/LoadingSpinner'

const AudioDeviceSelector = ({ devices = [], selectedDevice, onDeviceSelect, onClose }) => {
  const [availableDevices, setAvailableDevices] = useState(devices)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [tempSelected, setTempSelected] = useState(selectedDevice)

  useEffect(() => {
    if (devices.length === 0) {
      refreshDevices()
    } else {
      setAvailableDevices(devices)
    }
  }, [devices])

  const refreshDevices = async () => {
    try {
      setLoading(true)
      setError(null)
      const deviceList = await GetAudioDevices()
      setAvailableDevices(deviceList || [])
    } catch (err) {
      console.error('Failed to load audio devices:', err)
      setError('Failed to load audio devices')
    } finally {
      setLoading(false)
    }
  }

  const handleApply = () => {
    onDeviceSelect(tempSelected)
    onClose()
  }


  return (
    <Modal title="Audio Device Settings" onClose={onClose}>
      <div className="audio-device-selector">
        {loading && (
          <div className="loading-section">
            <LoadingSpinner />
            <p>Loading audio devices...</p>
          </div>
        )}

        {error && (
          <div className="error-section">
            <p className="error-message">{error}</p>
            <button onClick={refreshDevices} className="btn btn-secondary">
              🔄 Try Again
            </button>
          </div>
        )}

        {!loading && !error && (
          <>
            <div className="device-list-header">
              <h4>Available Input Devices</h4>
              <button onClick={refreshDevices} className="btn btn-refresh">
                🔄 Refresh
              </button>
            </div>

            <div className="device-list">
              {availableDevices.length === 0 ? (
                <p className="no-devices">No audio devices found</p>
              ) : (
                availableDevices.map((device) => (
                  <div
                    key={device.id}
                    className={`device-item ${tempSelected === device.id ? 'selected' : ''}`}
                  >
                    <label className="device-label">
                      <input
                        type="radio"
                        name="audioDevice"
                        value={device.id}
                        checked={tempSelected === device.id}
                        onChange={(e) => setTempSelected(e.target.value)}
                        className="device-radio"
                      />
                      <div className="device-info">
                        <span className="device-name">{device.name}</span>
                        <span className="device-type">{device.type}</span>
                        {device.default && (
                          <span className="device-badge">Default</span>
                        )}
                      </div>
                    </label>
                  </div>
                ))
              )}
            </div>

            <div className="device-settings">
              <h4>Recording Settings</h4>
              <div className="settings-grid">
                <div className="setting-item">
                  <label>Sample Rate</label>
                  <select className="form-select">
                    <option value="44100">44.1 kHz</option>
                    <option value="48000">48.0 kHz</option>
                  </select>
                </div>
                <div className="setting-item">
                  <label>Bit Depth</label>
                  <select className="form-select">
                    <option value="16">16 bit</option>
                    <option value="24">24 bit</option>
                  </select>
                </div>
              </div>
            </div>

            <div className="modal-actions">
              <button onClick={onClose} className="btn btn-secondary">
                Cancel
              </button>
              <button
                onClick={handleApply}
                className="btn btn-primary"
                disabled={!tempSelected}
              >
                Apply Settings
              </button>
            </div>
          </>
        )}
      </div>
    </Modal>
  )
}

export default AudioDeviceSelector