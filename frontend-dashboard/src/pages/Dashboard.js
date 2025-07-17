import React, { useEffect, useState } from 'react';
import { Box, AppBar, Toolbar, Typography, Button, Drawer, List, ListItem, ListItemText, Grid, Card, CardContent, CircularProgress } from '@mui/material';
import api from '../api/axios';
import { useNavigate, useLocation } from 'react-router-dom';

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

function StatCard({ title, count, loading }) {
  return (
    <Card sx={{ minWidth: 200 }}>
      <CardContent>
        <Typography variant="h6">{title}</Typography>
        {loading ? <CircularProgress size={24} /> : <Typography variant="h4">{count}</Typography>}
      </CardContent>
    </Card>
  );
}

export default function Dashboard() {
  const [stats, setStats] = useState({ users: 0, products: 0, orders: 0, payments: 0 });
  const [loading, setLoading] = useState(true);

  const fetchStats = async () => {
    setLoading(true);
    try {
      const [users, products, orders, payments] = await Promise.all([
        api.get('/users'),
        api.get('/products'),
        api.get('/orders'),
        api.get('/payments'),
      ]);
      setStats({
        users: users.data.length,
        products: products.data.length,
        orders: orders.data.length,
        payments: payments.data.length,
      });
    } catch (e) {
      // handle error
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    window.location.href = '/login';
  };

  return (
    <Box sx={{ display: 'flex' }}>
      <Header onLogout={handleLogout} />
      <Sidebar />
      <Box component="main" sx={{ flexGrow: 1, p: 3, ml: `${drawerWidth}px`, mt: 8 }}>
        <Typography variant="h4" gutterBottom>Overview</Typography>
        <Grid container spacing={2}>
          <Grid item><StatCard title="Users" count={stats.users} loading={loading} /></Grid>
          <Grid item><StatCard title="Products" count={stats.products} loading={loading} /></Grid>
          <Grid item><StatCard title="Orders" count={stats.orders} loading={loading} /></Grid>
          <Grid item><StatCard title="Payments" count={stats.payments} loading={loading} /></Grid>
        </Grid>
      </Box>
    </Box>
  );
} 