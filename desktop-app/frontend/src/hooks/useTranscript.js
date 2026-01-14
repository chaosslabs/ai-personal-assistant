import { useState, useEffect, useCallback } from 'react'
import { GetTranscript, SearchTranscripts } from '../../wailsjs/go/main/App'

export const useTranscript = (activityId) => {
  const [transcript, setTranscript] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const loadTranscript = useCallback(async (id) => {
    if (!id) {
      setTranscript([])
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const chunks = await GetTranscript(id)
      setTranscript(chunks || [])
    } catch (err) {
      console.error('Failed to load transcript:', err)
      setError(err.message || 'Failed to load transcript')
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshTranscript = useCallback(() => {
    if (activityId) {
      loadTranscript(activityId)
    }
  }, [activityId, loadTranscript])

  useEffect(() => {
    loadTranscript(activityId)
  }, [activityId, loadTranscript])

  return {
    transcript,
    loading,
    error,
    refreshTranscript
  }
}

export const useTranscriptSearch = () => {
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [query, setQuery] = useState('')

  const searchTranscripts = useCallback(async (searchQuery) => {
    if (!searchQuery?.trim()) {
      setResults([])
      setQuery('')
      return
    }

    try {
      setLoading(true)
      setError(null)
      setQuery(searchQuery)
      
      const searchResults = await SearchTranscripts(searchQuery.trim())
      setResults(searchResults || [])
    } catch (err) {
      console.error('Failed to search transcripts:', err)
      setError(err.message || 'Failed to search transcripts')
    } finally {
      setLoading(false)
    }
  }, [])

  const clearSearch = useCallback(() => {
    setResults([])
    setQuery('')
    setError(null)
  }, [])

  return {
    results,
    loading,
    error,
    query,
    searchTranscripts,
    clearSearch
  }
}

export const useLocalTranscriptSearch = (transcriptChunks) => {
  const [filteredChunks, setFilteredChunks] = useState(transcriptChunks)
  const [searchTerm, setSearchTerm] = useState('')

  const searchLocal = useCallback((query) => {
    setSearchTerm(query)
    
    if (!query?.trim()) {
      setFilteredChunks(transcriptChunks)
      return
    }

    const filtered = transcriptChunks.filter(chunk =>
      chunk.text.toLowerCase().includes(query.toLowerCase())
    )
    setFilteredChunks(filtered)
  }, [transcriptChunks])

  const clearSearch = useCallback(() => {
    setSearchTerm('')
    setFilteredChunks(transcriptChunks)
  }, [transcriptChunks])

  useEffect(() => {
    setFilteredChunks(transcriptChunks)
  }, [transcriptChunks])

  return {
    filteredChunks,
    searchTerm,
    searchLocal,
    clearSearch,
    resultCount: filteredChunks.length,
    totalCount: transcriptChunks.length
  }
}