import axios from 'axios';

// Use KrakenD gateway as the API base URL
const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Attach JWT token if present
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

export default api; 