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
        console.log('Token payload:', payload);
        console.log('Token expires at:', new Date(payload.exp * 1000));
      } else {
        console.log('Token expired, removing from localStorage');
        console.log('Current time:', currentTime, 'Token exp:', payload?.exp);
        localStorage.removeItem('token');
      }
    } catch (error) {
      console.log('Invalid token format, removing from localStorage:', error);
      localStorage.removeItem('token');
    }
  } else {
    console.log('No token found in localStorage for request:', config.url);
  }
  return config;
});

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.log('API Error:', {
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data,
      url: error.config?.url,
      method: error.config?.method
    });
    
    if (error.response?.status === 401) {
      console.log('401 Unauthorized error detected');
      console.log('Current path:', window.location.pathname);
      console.log('Token in localStorage:', localStorage.getItem('token') ? 'Present' : 'Not present');
      
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