import React, { useState, useEffect } from 'react'
import { SearchTranscripts } from '../../../wailsjs/go/main/App'
import LoadingSpinner from '../ui/LoadingSpinner'

const TranscriptSearch = ({ 
  onSearch, 
  searchTerm = '', 
  resultCount = 0, 
  totalCount = 0,
  globalSearch = false 
}) => {
  const [query, setQuery] = useState(searchTerm)
  const [globalResults, setGlobalResults] = useState([])
  const [isSearching, setIsSearching] = useState(false)
  const [searchError, setSearchError] = useState(null)
  const [showGlobalResults, setShowGlobalResults] = useState(false)
  const [searchHistory, setSearchHistory] = useState([])

  useEffect(() => {
    setQuery(searchTerm)
  }, [searchTerm])

  useEffect(() => {
    // Load search history from localStorage
    const history = JSON.parse(localStorage.getItem('transcriptSearchHistory') || '[]')
    setSearchHistory(history)
  }, [])

  const handleSearch = (searchQuery) => {
    if (onSearch) {
      onSearch(searchQuery)
    }
    
    if (searchQuery && !searchHistory.includes(searchQuery)) {
      const newHistory = [searchQuery, ...searchHistory.slice(0, 9)] // Keep last 10 searches
      setSearchHistory(newHistory)
      localStorage.setItem('transcriptSearchHistory', JSON.stringify(newHistory))
    }
  }

  const handleGlobalSearch = async (searchQuery) => {
    if (!searchQuery.trim()) return

    try {
      setIsSearching(true)
      setSearchError(null)
      
      const results = await SearchTranscripts(searchQuery.trim())
      setGlobalResults(results || [])
      setShowGlobalResults(true)
    } catch (error) {
      console.error('Global search failed:', error)
      setSearchError('Search failed: ' + error.message)
    } finally {
      setIsSearching(false)
    }
  }

  const handleInputChange = (e) => {
    const value = e.target.value
    setQuery(value)
    
    // Debounce local search
    const timeoutId = setTimeout(() => {
      handleSearch(value)
    }, 300)
    
    return () => clearTimeout(timeoutId)
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleSearch(query)
      
      if (globalSearch) {
        handleGlobalSearch(query)
      }
    }
  }

  const handleClearSearch = () => {
    setQuery('')
    handleSearch('')
    setShowGlobalResults(false)
    setGlobalResults([])
    setSearchError(null)
  }

  const handleHistorySelect = (historyItem) => {
    setQuery(historyItem)
    handleSearch(historyItem)
    
    if (globalSearch) {
      handleGlobalSearch(historyItem)
    }
  }

  const formatTimestamp = (timestamp) => {
    const seconds = Math.floor(timestamp)
    const minutes = Math.floor(seconds / 60)
    const formattedSeconds = (seconds % 60).toString().padStart(2, '0')
    const formattedMinutes = (minutes % 60).toString().padStart(2, '0')
    return `${formattedMinutes}:${formattedSeconds}`
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

  return (
    <div className="transcript-search">
      <div className="search-input-container">
        <input
          type="text"
          value={query}
          onChange={handleInputChange}
          onKeyPress={handleKeyPress}
          placeholder={globalSearch ? "Search all transcripts..." : "Search this transcript..."}
          className="search-input"
        />
        
        <div className="search-actions">
          {query && (
            <button
              onClick={handleClearSearch}
              className="search-clear"
              title="Clear search"
            >
              ✕
            </button>
          )}
          
          {globalSearch && (
            <button
              onClick={() => handleGlobalSearch(query)}
              disabled={isSearching || !query.trim()}
              className="search-button"
              title="Search all transcripts"
            >
              {isSearching ? <LoadingSpinner size="small" /> : '🔍'}
            </button>
          )}
        </div>
      </div>

      {/* Search results summary */}
      {query && (
        <div className="search-summary">
          <span>
            Found {resultCount} result{resultCount !== 1 ? 's' : ''} 
            {totalCount > 0 && ` of ${totalCount} total segments`}
          </span>
        </div>
      )}

      {/* Search history */}
      {!query && searchHistory.length > 0 && (
        <div className="search-history">
          <label className="search-history-label">Recent searches:</label>
          <div className="search-history-items">
            {searchHistory.slice(0, 5).map((item, index) => (
              <button
                key={index}
                onClick={() => handleHistorySelect(item)}
                className="search-history-item"
              >
                {item}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Search error */}
      {searchError && (
        <div className="search-error">
          <p className="error-message">{searchError}</p>
        </div>
      )}

      {/* Global search results */}
      {showGlobalResults && globalResults.length > 0 && (
        <div className="global-search-results">
          <div className="global-results-header">
            <h4>Search Results ({globalResults.length})</h4>
            <button
              onClick={() => setShowGlobalResults(false)}
              className="btn-secondary btn-small"
            >
              Close
            </button>
          </div>
          
          <div className="global-results-list">
            {globalResults.map((result, index) => (
              <div key={`${result.activity_id}-${index}`} className="global-result-item">
                <div className="result-meta">
                  <span className="result-activity">Activity: {result.activity_id}</span>
                  <span className="result-timestamp">{formatTimestamp(result.start_time)}</span>
                </div>
                <div className="result-text">
                  {highlightText(result.text, query)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {showGlobalResults && globalResults.length === 0 && !isSearching && (
        <div className="no-global-results">
          <p>No results found across all transcripts for "{query}"</p>
        </div>
      )}
    </div>
  )
}

export default TranscriptSearch