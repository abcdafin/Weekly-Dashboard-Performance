import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Handle 401 Unauthorized
      if (error.response.status === 401) {
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      }
      
      // Handle 403 Forbidden (no spreadsheet access)
      if (error.response.status === 403) {
        console.error('Access denied:', error.response.data?.error)
      }
    }
    return Promise.reject(error)
  }
)

export default api

// API helper functions
export const dashboardApi = {
  getDashboard: (month, year, refresh = false) => 
    api.get('/dashboard', { params: { month, year, ...(refresh ? { refresh: 'true' } : {}) } }),
  
  getMonths: () => 
    api.get('/months'),
  
  compare: (month, year, compareWith = 'previous_month') => 
    api.get('/dashboard/compare', { params: { month, year, compareWith } }),
  
  saveSnapshot: (month, year, week) =>
    api.post('/dashboard/snapshot', null, { params: { month, year, week } }),
  
  getSnapshots: (month, year) =>
    api.get('/dashboard/snapshots', { params: { month, year } }),
  
  deleteSnapshot: (month, year, week) =>
    api.delete('/dashboard/snapshot', { params: { month, year, week } }),
}

export const screenshotApi = {
  upload: (formData) =>
    api.post('/dashboard/screenshot', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    }),
  
  list: (month, year) =>
    api.get('/dashboard/screenshots', { params: { month, year } }),
  
  getImage: (id) =>
    api.get(`/dashboard/screenshot/${id}`),
  
  getImageUrl: (id) =>
    `${import.meta.env.VITE_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}/api/v1/screenshot/image/${id}`
}

export const authApi = {
  getCurrentUser: () => 
    api.get('/auth/me'),
  
  logout: () => 
    api.post('/auth/logout'),
}

export const settingsApi = {
  getSpreadsheet: () =>
    api.get('/settings/spreadsheet'),
  
  updateSpreadsheet: (data) =>
    api.put('/settings/spreadsheet', data),
}

