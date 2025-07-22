import React from 'react';
import { Box, AppBar, Toolbar, Typography, Button, Drawer, List, ListItem, ListItemText } from '@mui/material';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';

const drawerWidth = 200;

function Sidebar() {
  const navigate = useNavigate();
  const location = useLocation();
  const items = [
    { text: 'Dashboard', path: '/' },
    { text: 'Users', path: '/users' },
    { text: 'Products', path: '/products' },
    { text: 'Orders', path: '/orders' },
    { text: 'Payments', path: '/payments' },
  ];
  return (
    <Drawer variant="permanent" sx={{ width: drawerWidth, flexShrink: 0, [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box' } }}>
      <Toolbar />
      <Box sx={{ overflow: 'auto' }}>
        <List>
          {items.map(({ text, path }) => (
            <ListItem button key={text} selected={location.pathname === path} onClick={() => navigate(path)}>
              <ListItemText primary={text} />
            </ListItem>
          ))}
        </List>
      </Box>
    </Drawer>
  );
}

function Header({ onLogout }) {
  return (
    <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
      <Toolbar>
        <Typography variant="h6" noWrap sx={{ flexGrow: 1 }}>
          ScaleBit Dashboard
        </Typography>
        <Button color="inherit" onClick={onLogout}>Logout</Button>
      </Toolbar>
    </AppBar>
  );
}

export default function Layout() {
  const handleLogout = () => {
    localStorage.removeItem('token');
    window.location.href = '/login';
  };

  return (
    <Box sx={{ display: 'flex' }}>
      <Header onLogout={handleLogout} />
      <Sidebar />
      <Box component="main" sx={{ flexGrow: 1, p: 3, ml: `${drawerWidth}px`, mt: '64px' }}>
        <Outlet />
      </Box>
    </Box>
  );
}