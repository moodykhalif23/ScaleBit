import React, { useEffect, useState, useContext } from 'react';
import { Box, Typography, Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField, IconButton, Snackbar, Alert, CircularProgress, InputAdornment, DialogContentText, Tooltip } from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import { Edit, Delete } from '@mui/icons-material';
import api from '../api/axios';
import { TextField as MuiTextField } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { RoleContext } from '../App';

export default function Payments() {
  const [payments, setPayments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editPayment, setEditPayment] = useState(null);
  const [form, setForm] = useState({ order_id: '', amount: '', status: '', timestamp: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [deleteId, setDeleteId] = useState(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [detailPayment, setDetailPayment] = useState(null);
  const role = useContext(RoleContext);

  const fetchPayments = async () => {
    setLoading(true);
    try {
      const res = await api.get('/payments');
      setPayments(res.data);
    } catch (e) {
      setSnackbar({ open: true, message: 'Failed to fetch payments', severity: 'error' });
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchPayments();
  }, []);

  const handleOpen = (payment = null) => {
    setEditPayment(payment);
    setForm(payment ? { order_id: payment.order_id, amount: payment.amount, status: payment.status, timestamp: payment.timestamp } : { order_id: '', amount: '', status: '', timestamp: '' });
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setEditPayment(null);
    setForm({ order_id: '', amount: '', status: '', timestamp: '' });
  };

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    try {
      if (editPayment) {
        await api.put(`/payments/${editPayment.id}`, { order_id: parseInt(form.order_id), amount: parseFloat(form.amount), status: form.status, timestamp: form.timestamp });
        setSnackbar({ open: true, message: 'Payment updated', severity: 'success' });
      } else {
        await api.post('/payments', { order_id: parseInt(form.order_id), amount: parseFloat(form.amount), status: form.status, timestamp: form.timestamp });
        setSnackbar({ open: true, message: 'Payment created', severity: 'success' });
      }
      handleClose();
      fetchPayments();
    } catch (e) {
      setSnackbar({ open: true, message: 'Operation failed', severity: 'error' });
    }
  };

  const handleDelete = async () => {
    setDeleteLoading(true);
    try {
      await api.delete(`/payments/${deleteId}`);
      setSnackbar({ open: true, message: 'Payment deleted', severity: 'success' });
      setDeleteId(null);
      fetchPayments();
    } catch (e) {
      setSnackbar({ open: true, message: 'Delete failed', severity: 'error' });
    }
    setDeleteLoading(false);
  };

  const columns = [
    { field: 'id', headerName: 'ID', width: 90 },
    { field: 'order_id', headerName: 'Order ID', flex: 1 },
    { field: 'amount', headerName: 'Amount', flex: 1, valueFormatter: ({ value }) => `$${value}` },
    { field: 'status', headerName: 'Status', flex: 1 },
    { field: 'timestamp', headerName: 'Timestamp', flex: 1 },
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
              <>
                <IconButton disabled><Edit /></IconButton>
                <IconButton disabled color="error"><Delete /></IconButton>
              </>
            </Tooltip>
          )}
        </>
      ),
      sortable: false,
      filterable: false,
    },
  ];

  const filteredPayments = payments.filter(p =>
    p.id.toString().includes(search.toLowerCase())
  );

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Payments</Typography>
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
        <Button variant="contained" sx={{ mb: 2, ml: 0 }} onClick={() => handleOpen()}>Add Payment</Button>
      ) : (
        <Tooltip title="Admin only">
          <span>
            <Button variant="contained" sx={{ mb: 2, ml: 0 }} disabled>Add Payment</Button>
          </span>
        </Tooltip>
      )}
      {loading ? <CircularProgress /> : (
        <DataGrid
          rows={filteredPayments}
          columns={columns}
          autoHeight
          pageSize={10}
          rowsPerPageOptions={[10, 20, 50]}
          onRowClick={(params, event) => {
            if (event.target.closest('button')) return;
            setDetailPayment(params.row);
          }}
        />
      )}
      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>{editPayment ? 'Edit Payment' : 'Add Payment'}</DialogTitle>
        <DialogContent>
          <TextField margin="normal" fullWidth label="Order ID" name="order_id" type="number" value={form.order_id} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Amount" name="amount" type="number" value={form.amount} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Status" name="status" value={form.status} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Timestamp" name="timestamp" value={form.timestamp} onChange={handleChange} helperText="Format: YYYY-MM-DDTHH:mm:ss.sssZ" />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">{editPayment ? 'Update' : 'Create'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!deleteId} onClose={() => setDeleteId(null)}>
        <DialogTitle>Delete Payment</DialogTitle>
        <DialogContent>Are you sure you want to delete this payment?</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteId(null)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained" disabled={deleteLoading}>{deleteLoading ? <CircularProgress size={20} /> : 'Delete'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!detailPayment} onClose={() => setDetailPayment(null)}>
        <DialogTitle>Payment Details</DialogTitle>
        <DialogContent>
          {detailPayment && (
            <DialogContentText>
              <b>ID:</b> {detailPayment.id}<br/>
              <b>Order ID:</b> {detailPayment.order_id}<br/>
              <b>Amount:</b> ${detailPayment.amount}<br/>
              <b>Status:</b> {detailPayment.status}<br/>
              <b>Timestamp:</b> {detailPayment.timestamp}<br/>
            </DialogContentText>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailPayment(null)}>Close</Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={snackbar.open} autoHideDuration={3000} onClose={() => setSnackbar({ ...snackbar, open: false })}>
        <Alert severity={snackbar.severity} sx={{ width: '100%' }}>{snackbar.message}</Alert>
      </Snackbar>
    </Box>
  );
} 