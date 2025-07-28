import React, { useEffect, useState, useContext } from 'react';
import { Box, Typography, Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField, IconButton, Snackbar, Alert, CircularProgress, InputAdornment, DialogContentText, Tooltip } from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import { Edit, Delete } from '@mui/icons-material';
import api from '../api/axios';
import { TextField as MuiTextField } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { RoleContext } from '../App';

export default function Products() {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editProduct, setEditProduct] = useState(null);
  const [form, setForm] = useState({ name: '', price: '', stock: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [deleteId, setDeleteId] = useState(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [detailProduct, setDetailProduct] = useState(null);
  const role = useContext(RoleContext);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      const res = await api.get('/products');
      setProducts(res.data);
    } catch (e) {
      setSnackbar({ open: true, message: 'Failed to fetch products', severity: 'error' });
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchProducts();
  }, []);

  const handleOpen = (product = null) => {
    setEditProduct(product);
    setForm(product ? { name: product.name, price: product.price, stock: product.stock } : { name: '', price: '', stock: '' });
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setEditProduct(null);
    setForm({ name: '', price: '', stock: '' });
  };

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    try {
      if (editProduct) {
        await api.put(`/products/${editProduct.id}`, { name: form.name, price: parseFloat(form.price), stock: parseInt(form.stock) });
        setSnackbar({ open: true, message: 'Product updated', severity: 'success' });
      } else {
        await api.post('/products', { name: form.name, price: parseFloat(form.price), stock: parseInt(form.stock) });
        setSnackbar({ open: true, message: 'Product created', severity: 'success' });
      }
      handleClose();
      fetchProducts();
    } catch (e) {
      setSnackbar({ open: true, message: 'Operation failed', severity: 'error' });
    }
  };

  const handleDelete = async () => {
    setDeleteLoading(true);
    try {
      await api.delete(`/products/${deleteId}`);
      setSnackbar({ open: true, message: 'Product deleted', severity: 'success' });
      setDeleteId(null);
      fetchProducts();
    } catch (e) {
      setSnackbar({ open: true, message: 'Delete failed', severity: 'error' });
    }
    setDeleteLoading(false);
  };

  const columns = [
    { field: 'id', headerName: 'ID', width: 90 },
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'price', headerName: 'Price', flex: 1, valueFormatter: ({ value }) => `$${value}` },
    { field: 'stock', headerName: 'Stock', flex: 1 },
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

  const filteredProducts = products.filter(p =>
    p.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Products</Typography>
      <MuiTextField
        placeholder="Search by name"
        value={search}
        onChange={e => setSearch(e.target.value)}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
        sx={{ mb: 2, width: 300 }}
        size="small"
      />
      {role === 'admin' ? (
        <Button variant="contained" sx={{ mb: 2, ml: 0 }} onClick={() => handleOpen()}>Add Product</Button>
      ) : (
        <Tooltip title="Admin only">
          <>
            <Button variant="contained" sx={{ mb: 2, ml: 0 }} disabled>Add Product</Button>
          </>
        </Tooltip>
      )}
      {loading ? <CircularProgress /> : (
        <DataGrid
          rows={filteredProducts}
          columns={columns}
          autoHeight
          pageSize={10}
          rowsPerPageOptions={[10, 20, 50]}
          onRowClick={(params, event) => {
            if (event.target.closest('button')) return;
            setDetailProduct(params.row);
          }}
        />
      )}
      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>{editProduct ? 'Edit Product' : 'Add Product'}</DialogTitle>
        <DialogContent>
          <TextField margin="normal" fullWidth label="Name" name="name" value={form.name} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Price" name="price" type="number" value={form.price} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Stock" name="stock" type="number" value={form.stock} onChange={handleChange} />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">{editProduct ? 'Update' : 'Create'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!deleteId} onClose={() => setDeleteId(null)}>
        <DialogTitle>Delete Product</DialogTitle>
        <DialogContent>Are you sure you want to delete this product?</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteId(null)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained" disabled={deleteLoading}>{deleteLoading ? <CircularProgress size={20} /> : 'Delete'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!detailProduct} onClose={() => setDetailProduct(null)}>
        <DialogTitle>Product Details</DialogTitle>
        <DialogContent>
          {detailProduct && (
            <DialogContentText>
              <b>ID:</b> {detailProduct.id}<br/>
              <b>Name:</b> {detailProduct.name}<br/>
              <b>Price:</b> ${detailProduct.price}<br/>
              <b>Stock:</b> {detailProduct.stock}<br/>
            </DialogContentText>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailProduct(null)}>Close</Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={snackbar.open} autoHideDuration={3000} onClose={() => setSnackbar({ ...snackbar, open: false })}>
        <Alert severity={snackbar.severity} sx={{ width: '100%' }}>{snackbar.message}</Alert>
      </Snackbar>
    </Box>
  );
} 