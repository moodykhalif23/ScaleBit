import React, { useMemo } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Products from './pages/Products';
import Orders from './pages/Orders';
import Payments from './pages/Payments';
import { getUserRoleFromToken } from './api/axios';
import Layout from './pages/Layout';

export const RoleContext = React.createContext(null);

function PrivateRoute({ children }) {
  const token = localStorage.getItem('token');

  // Check if token exists and is not expired
  if (!token) {
    return <Navigate to="/login" replace />;
  }

  try {
    // Parse JWT to check expiration
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Date.now() / 1000;

    if (payload.exp && payload.exp < currentTime) {
      // Token is expired, remove it and redirect to login
      localStorage.removeItem('token');
      return <Navigate to="/login" replace />;
    }
  } catch (error) {
    // Invalid token format, remove it and redirect to login
    localStorage.removeItem('token');
    return <Navigate to="/login" replace />;
  }

  return children;
}

function App() {
  const role = useMemo(() => getUserRoleFromToken(), []);

  return (
    <RoleContext.Provider value={role}>
      <Router>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route element={<PrivateRoute><Layout /></PrivateRoute>}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/users" element={<Users />} />
            <Route path="/products" element={<Products />} />
            <Route path="/orders" element={<Orders />} />
            <Route path="/payments" element={<Payments />} />
          </Route>
          {/* Redirect any unknown routes to login if not authenticated, or dashboard if authenticated */}
          <Route path="*" element={<Navigate to={localStorage.getItem('token') ? "/" : "/login"} replace />} />
        </Routes>
      </Router>
    </RoleContext.Provider>
  );
}

export default App; 