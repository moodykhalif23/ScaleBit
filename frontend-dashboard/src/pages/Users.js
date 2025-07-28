import React, { useEffect, useState, useContext } from 'react';
import { Box, Typography, Button, Dialog, DialogTitle, DialogContent, DialogActions, TextField, IconButton, Snackbar, Alert, CircularProgress, InputAdornment, DialogContentText, Tooltip, MenuItem, Select, FormControl, InputLabel } from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import { Edit, Delete } from '@mui/icons-material';
import api from '../api/axios';
import { TextField as MuiTextField } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { RoleContext } from '../App';

export default function Users() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editUser, setEditUser] = useState(null);
  const [form, setForm] = useState({ name: '', email: '', password: '' });
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [deleteId, setDeleteId] = useState(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [detailUser, setDetailUser] = useState(null);
  const role = useContext(RoleContext);
  const [roleUpdating, setRoleUpdating] = useState(false);

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await api.get('/users');
      setUsers(res.data);
    } catch (e) {
      setSnackbar({ open: true, message: 'Failed to fetch users', severity: 'error' });
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleOpen = (user = null) => {
    setEditUser(user);
    setForm(user ? { name: user.name, email: user.email, password: '' } : { name: '', email: '', password: '' });
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setEditUser(null);
    setForm({ name: '', email: '', password: '' });
  };

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    try {
      if (editUser) {
        await api.put(`/users/${editUser.id}`, { name: form.name, email: form.email });
        setSnackbar({ open: true, message: 'User updated', severity: 'success' });
      } else {
        await api.post('/users', form);
        setSnackbar({ open: true, message: 'User created', severity: 'success' });
      }
      handleClose();
      fetchUsers();
    } catch (e) {
      setSnackbar({ open: true, message: 'Operation failed', severity: 'error' });
    }
  };

  const handleDelete = async () => {
    setDeleteLoading(true);
    try {
      await api.delete(`/users/${deleteId}`);
      setSnackbar({ open: true, message: 'User deleted', severity: 'success' });
      setDeleteId(null);
      fetchUsers();
    } catch (e) {
      setSnackbar({ open: true, message: 'Delete failed', severity: 'error' });
    }
    setDeleteLoading(false);
  };

  const handleRoleChange = async (user, newRole) => {
    setRoleUpdating(true);
    try {
      await api.put(`/users/${user.id}`, { ...user, role: newRole });
      setSnackbar({ open: true, message: 'Role updated', severity: 'success' });
      setDetailUser({ ...user, role: newRole });
      fetchUsers();
    } catch (e) {
      setSnackbar({ open: true, message: 'Failed to update role', severity: 'error' });
    }
    setRoleUpdating(false);
  };

  const columns = [
    { field: 'id', headerName: 'ID', width: 90 },
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'email', headerName: 'Email', flex: 1 },
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

  const filteredUsers = users.filter(u =>
    u.name.toLowerCase().includes(search.toLowerCase()) ||
    u.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Users</Typography>
      <MuiTextField
        placeholder="Search by name or email"
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
        <Button variant="contained" sx={{ mb: 2, ml: 0 }} onClick={() => handleOpen()}>Add User</Button>
      ) : (
        <Tooltip title="Admin only">
          <span>
            <Button variant="contained" sx={{ mb: 2, ml: 0 }} disabled>Add User</Button>
          </span>
        </Tooltip>
      )}
      {loading ? <CircularProgress /> : (
        <DataGrid
          rows={filteredUsers}
          columns={columns}
          autoHeight
          pageSize={10}
          rowsPerPageOptions={[10, 20, 50]}
          onRowClick={(params, event) => {
            // Prevent opening detail on edit/delete click
            if (event.target.closest('button')) return;
            setDetailUser(params.row);
          }}
        />
      )}
      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>{editUser ? 'Edit User' : 'Add User'}</DialogTitle>
        <DialogContent>
          <TextField margin="normal" fullWidth label="Name" name="name" value={form.name} onChange={handleChange} />
          <TextField margin="normal" fullWidth label="Email" name="email" value={form.email} onChange={handleChange} />
          {!editUser && <TextField margin="normal" fullWidth label="Password" name="password" type="password" value={form.password} onChange={handleChange} />}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">{editUser ? 'Update' : 'Create'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!deleteId} onClose={() => setDeleteId(null)}>
        <DialogTitle>Delete User</DialogTitle>
        <DialogContent>Are you sure you want to delete this user?</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteId(null)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained" disabled={deleteLoading}>{deleteLoading ? <CircularProgress size={20} /> : 'Delete'}</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={!!detailUser} onClose={() => setDetailUser(null)}>
        <DialogTitle>User Details</DialogTitle>
        <DialogContent>
          {detailUser && (
            <DialogContentText>
              <b>ID:</b> {detailUser.id}<br/>
              <b>Name:</b> {detailUser.name}<br/>
              <b>Email:</b> {detailUser.email}<br/>
              <b>Role:</b> {role === 'admin' && detailUser.id !== users.find(u => u.email === localStorage.getItem('email'))?.id ? (
                <FormControl size="small" sx={{ minWidth: 120, ml: 1 }}>
                  <InputLabel>Role</InputLabel>
                  <Select
                    value={detailUser.role}
                    label="Role"
                    onChange={e => handleRoleChange(detailUser, e.target.value)}
                    disabled={roleUpdating}
                  >
                    <MenuItem value="admin">admin</MenuItem>
                    <MenuItem value="user">user</MenuItem>
                  </Select>
                </FormControl>
              ) : (
                <span style={{ marginLeft: 8 }}>{detailUser.role}</span>
              )}
            </DialogContentText>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailUser(null)}>Close</Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={snackbar.open} autoHideDuration={3000} onClose={() => setSnackbar({ ...snackbar, open: false })}>
        <Alert severity={snackbar.severity} sx={{ width: '100%' }}>{snackbar.message}</Alert>
      </Snackbar>
    </Box>
  );
} 