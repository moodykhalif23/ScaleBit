import axios from 'axios';

// Use KrakenD gateway as the API base URL
const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8000';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Attach JWT token if present
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    // Validate token before sending
    try {
      const payload = parseJwt(token);
      const currentTime = Date.now() / 1000;

      if (payload && payload.exp && payload.exp > currentTime) {
        config.headers['Authorization'] = `Bearer ${token}`;
        console.log('Adding valid token to request:', config.url);
      } else {
        console.log('Token expired, removing from localStorage');
        localStorage.removeItem('token');
      }
    } catch (error) {
      console.log('Invalid token format, removing from localStorage');
      localStorage.removeItem('token');
    }
  }
  return config;
});

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Only redirect if we're not already on the login page
      if (!window.location.pathname.includes('/login')) {
        // Token is invalid or expired, remove it and redirect to login
        localStorage.removeItem('token');
        console.log('Authentication failed, redirecting to login');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

function parseJwt(token) {
  try {
    return JSON.parse(atob(token.split('.')[1]));
  } catch (e) {
    return null;
  }
}

export function getUserRoleFromToken() {
  const token = localStorage.getItem('token');
  if (!token) return null;
  const payload = parseJwt(token);
  return payload && payload.role ? payload.role : null;
}

export default api; 