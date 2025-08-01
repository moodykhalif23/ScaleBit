import React, { useEffect, useState } from 'react';
import { Box, Typography, Grid, Card, CardContent, CircularProgress } from '@mui/material';
import api from '../api/axios';

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
      console.log('Fetching dashboard stats...');
      const [users, products, orders, payments] = await Promise.all([
        api.get('/users'),
        api.get('/products'),
        api.get('/orders'),
        api.get('/payments'),
      ]);
      console.log('Dashboard stats fetched successfully');
      setStats({
        users: users.data.length,
        products: products.data.length,
        orders: orders.data.length,
        payments: payments.data.length,
      });
    } catch (e) {
      console.error('Error fetching dashboard stats:', e);
      // Set default values on error
      setStats({ users: 0, products: 0, orders: 0, payments: 0 });
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchStats();
  }, []);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Overview</Typography>
      <Grid container spacing={2}>
        <Grid><StatCard title="Users" count={stats.users} loading={loading} /></Grid>
        <Grid><StatCard title="Products" count={stats.products} loading={loading} /></Grid>
        <Grid><StatCard title="Orders" count={stats.orders} loading={loading} /></Grid>
        <Grid><StatCard title="Payments" count={stats.payments} loading={loading} /></Grid>
      </Grid>
    </Box>
  );
} 