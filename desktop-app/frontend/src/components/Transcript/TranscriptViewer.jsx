import React, { useState, useEffect } from 'react'
import TranscriptSearch from './TranscriptSearch'

const TranscriptViewer = ({ chunks = [], activityId = null, searchable = true }) => {
  const [filteredChunks, setFilteredChunks] = useState(chunks)
  const [searchTerm, setSearchTerm] = useState('')
  const [highlightedText, setHighlightedText] = useState('')
  const [selectedChunk, setSelectedChunk] = useState(null)
  const [showTimestamps, setShowTimestamps] = useState(true)

  useEffect(() => {
    setFilteredChunks(chunks)
  }, [chunks])

  useEffect(() => {
    if (searchTerm) {
      const filtered = chunks.filter(chunk =>
        chunk.text.toLowerCase().includes(searchTerm.toLowerCase())
      )
      setFilteredChunks(filtered)
      setHighlightedText(searchTerm)
    } else {
      setFilteredChunks(chunks)
      setHighlightedText('')
    }
  }, [searchTerm, chunks])

  const formatTimestamp = (timestamp) => {
    try {
      const seconds = Math.floor(timestamp)
      const minutes = Math.floor(seconds / 60)
      const hours = Math.floor(minutes / 60)
      
      const formattedSeconds = (seconds % 60).toString().padStart(2, '0')
      const formattedMinutes = (minutes % 60).toString().padStart(2, '0')
      
      if (hours > 0) {
        const formattedHours = hours.toString().padStart(2, '0')
        return `${formattedHours}:${formattedMinutes}:${formattedSeconds}`
      }
      
      return `${formattedMinutes}:${formattedSeconds}`
    } catch {
      return '00:00'
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

  const highlightText = (text, searchTerm) => {
    if (!searchTerm || !text) return text
    
    const regex = new RegExp(`(${searchTerm})`, 'gi')
    const parts = text.split(regex)
    
    return parts.map((part, index) =>
      regex.test(part) ? (
        <mark key={index} className="search-highlight">{part}</mark>
      ) : (
        part
      )
    )
  }

  const handleChunkClick = (chunk) => {
    setSelectedChunk(selectedChunk?.id === chunk.id ? null : chunk)
  }

  const handleJumpToTime = (timestamp) => {
    // This would integrate with audio player if available
    console.log('Jump to time:', timestamp)
  }

  const exportTranscript = () => {
    const content = filteredChunks
      .map(chunk => {
        const timestamp = showTimestamps ? `[${formatTimestamp(chunk.start_time)}] ` : ''
        return `${timestamp}${chunk.text}`
      })
      .join('\n\n')
    
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `transcript-${activityId || 'export'}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  if (!chunks || chunks.length === 0) {
    return (
      <div className="transcript-viewer-empty">
        <p>No transcript available</p>
        <p className="text-muted">
          Transcripts will appear here after audio is processed
        </p>
      </div>
    )
  }

  return (
    <div className="transcript-viewer">
      {searchable && (
        <div className="transcript-controls">
          <TranscriptSearch
            onSearch={setSearchTerm}
            searchTerm={searchTerm}
            resultCount={filteredChunks.length}
            totalCount={chunks.length}
          />
          
          <div className="transcript-options">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={showTimestamps}
                onChange={(e) => setShowTimestamps(e.target.checked)}
              />
              Show timestamps
            </label>
            
            <button 
              onClick={exportTranscript}
              className="btn-secondary btn-small"
            >
              Export
            </button>
          </div>
        </div>
      )}

      <div className="transcript-content">
        <div className="transcript-stats">
          <span>
            Showing {filteredChunks.length} of {chunks.length} segments
          </span>
          {searchTerm && (
            <span>
              • Filtered by: "{searchTerm}"
            </span>
          )}
        </div>

        <div className="transcript-chunks">
          {filteredChunks.map((chunk, index) => (
            <div 
              key={chunk.id || index}
              className={`transcript-chunk ${selectedChunk?.id === chunk.id ? 'selected' : ''}`}
              onClick={() => handleChunkClick(chunk)}
            >
              <div className="chunk-header">
                {showTimestamps && (
                  <button 
                    className="timestamp-button"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleJumpToTime(chunk.start_time)
                    }}
                    title="Jump to this time"
                  >
                    {formatTimestamp(chunk.start_time)}
                  </button>
                )}
                
                <div className="chunk-meta">
                  {chunk.confidence && (
                    <span className={`confidence ${chunk.confidence > 0.8 ? 'high' : chunk.confidence > 0.6 ? 'medium' : 'low'}`}>
                      {Math.round(chunk.confidence * 100)}%
                    </span>
                  )}
                  
                  {chunk.speaker && (
                    <span className="speaker">
                      Speaker: {chunk.speaker}
                    </span>
                  )}
                </div>
              </div>
              
              <div className="chunk-text">
                {highlightText(chunk.text, highlightedText)}
              </div>
              
              {selectedChunk?.id === chunk.id && (
                <div className="chunk-details">
                  <div className="detail-grid">
                    <div className="detail-item">
                      <span className="detail-label">Start:</span>
                      <span className="detail-value">{formatTimestamp(chunk.start_time)}</span>
                    </div>
                    <div className="detail-item">
                      <span className="detail-label">End:</span>
                      <span className="detail-value">{formatTimestamp(chunk.end_time)}</span>
                    </div>
                    <div className="detail-item">
                      <span className="detail-label">Duration:</span>
                      <span className="detail-value">
                        {formatTimestamp(chunk.end_time - chunk.start_time)}
                      </span>
                    </div>
                    <div className="detail-item">
                      <span className="detail-label">Processed:</span>
                      <span className="detail-value">{formatDate(chunk.processed_at)}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default TranscriptViewer