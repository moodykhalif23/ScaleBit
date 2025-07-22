import React, { useEffect, useState, useContext } from 'react';
import { Box, Typography, Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField, IconButton, Snackbar, Alert, CircularProgress, InputAdornment, DialogContentText, Tooltip } from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import { Edit, Delete } from '@mui/icons-material';
import api from '../api/axios';
import { TextField as MuiTextField } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { RoleContext } from '../App';

export default function Orders() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editOrder, setEditOrder] = useState(null);
  const [form, setForm] = useState({ user_id: '', product_id: '', quantity: '', status: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [deleteId, setDeleteId] = useState(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [detailOrder, setDetailOrder] = useState(null);
  const role = useContext(RoleContext);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await api.get('/orders');
      setOrders(res.data);
    } catch (e) {
      setSnackbar({ open: true, message: 'Failed to fetch orders', severity: 'error' });
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchOrders();
  }, []);

  const handleOpen = (order = null) => {
    setEditOrder(order);
    setForm(order ? { user_id: order.user_id, product_id: order.product_id, quantity: order.quantity, status: order.status } : { user_id: '', product_id: '', quantity: '', status: '' });
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setEditOrder(null);
    setForm({ user_id: '', product_id: '', quantity: '', status: '' });
  };

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    try {
      if (editOrder) {
        await api.put(`/orders/${editOrder.id}`, { user_id: parseInt(form.user_id), product_id: parseInt(form.product_id), quantity: parseInt(form.quantity), status: form.status });
        setSnackbar({ open: true, message: 'Order updated', severity: 'success' });
      } else {
        await api.post('/orders', { user_id: parseInt(form.user_id), product_id: parseInt(form.product_id), quantity: parseInt(form.quantity), status: form.status });
        setSnackbar({ open: true, message: 'Order created', severity: 'success' });
      }
      handleClose();
      fetchOrders();
    } catch (e) {
      setSnackbar({ open: true, message: 'Operation failed', severity: 'error' });
    }
  };

  const handleDelete = async () => {
    setDeleteLoading(true);
    try {
      await api.delete(`/orders/${deleteId}`);
      setSnackbar({ open: true, message: 'Order deleted', severity: 'success' });
      setDeleteId(null);
      fetchOrders();
    } catch (e) {
      setSnackbar({ open: true, message: 'Delete failed', severity: 'error' });
    }
    setDeleteLoading(false);
  };

  const columns = [
    { field: 'id', headerName: 'ID', width: 90 },
    { field: 'user_id', headerName: 'User ID', flex: 1 },
    { field: 'product_id', headerName: 'Product ID', flex: 1 },
    { field: 'quantity', headerName: 'Quantity', flex: 1 },
    { field: 'status', headerName: 'Status', flex: 1 },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 120,
      renderCell: (params) => (
        <>
          {role === 'admin' ? (
            <>
              <IconButton onClick={() => handleOpen(params.row)}><Edit /></IconButton>
              <IconButton color="error" onClick={() => setDeleteId(params.row.id)}><Delete /></IconButton>
            </>
          ) : (
            <Tooltip title="Admin only">
              <span>
                <IconButton disabled><Edit /></IconButton>
                <IconButton disabled color="error"><Delete /></IconButton>
              </span>
            </Tooltip>
          )}
        </>
      ),
      sortable: false,
      filterable: false,
    },
  ];

  const filteredOrders = orders.filter(o =>
    o.id.toString().includes(search.toLowerCase())
  );

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Orders</Typography>
      <MuiTextField
        placeholder="Search by ID"
        value={search}
        onChange={e => setSearch(e.target.value)}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
        sx={{ mb: 2, width: 350 }}
        size="small"
      />
      {role === 'admin' ? (
        <Button variant="contained" sx={{ mb: 2, ml: 0 }} onClick={() => handleOpen()}>Add Order</Button>
      ) : (
        <Tooltip title="Admin only">
          <span>
            <Button variant="contained" sx={{ mb: 2, ml: 0 }} disabled>Add Order</Button>
          </span>
        </Tooltip>
      )}
      {loading ? <CircularProgress /> : (
        <DataGrid
          rows={filteredOrders}
          columns={columns}
          autoHeight
          pageSize={10}
          rowsPerPageOptions={[10, 20, 50]}
          onRowClick={(params, event) => {
            if (event.target.closest('button')) return;
            setDetailOrder(params.row);
          }}
        />
      )}
      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>{editOrder ? 'Edit Order' : 'Add Order'}</DialogTitle>
        <DialogContent>
          <TextField margin="normal" fullWidth label="User ID" name="user_id" type="number" value={form.user_id} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Product ID" name="product_id" type="number" value={form.product_id} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Quantity" name="quantity" type="number" value={form.quantity} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Status" name="status" value={form.status} onChange={handleChange} />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">{editOrder ? 'Update' : 'Create'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!deleteId} onClose={() => setDeleteId(null)}>
        <DialogTitle>Delete Order</DialogTitle>
        <DialogContent>Are you sure you want to delete this order?</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteId(null)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained" disabled={deleteLoading}>{deleteLoading ? <CircularProgress size={20} /> : 'Delete'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!detailOrder} onClose={() => setDetailOrder(null)}>
        <DialogTitle>Order Details</DialogTitle>
        <DialogContent>
          {detailOrder && (
            <DialogContentText>
              <b>ID:</b> {detailOrder.id}<br/>
              <b>User ID:</b> {detailOrder.user_id}<br/>
              <b>Product ID:</b> {detailOrder.product_id}<br/>
              <b>Quantity:</b> {detailOrder.quantity}<br/>
              <b>Status:</b> {detailOrder.status}<br/>
            </DialogContentText>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOrder(null)}>Close</Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={snackbar.open} autoHideDuration={3000} onClose={() => setSnackbar({ ...snackbar, open: false })}>
        <Alert severity={snackbar.severity} sx={{ width: '100%' }}>{snackbar.message}</Alert>
      </Snackbar>
    </Box>
  );
} 