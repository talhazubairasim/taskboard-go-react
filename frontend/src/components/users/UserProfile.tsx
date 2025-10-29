import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Avatar,
  Button,
  Grid,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  Divider,
  Alert,
  CircularProgress,
  Card,
  CardContent,
  List,
  ListItem,
  ListItemText,
  IconButton,
} from '@mui/material';
import {
  Edit as EditIcon,
  Save as SaveIcon,
  Cancel as CancelIcon,
  Email as EmailIcon,
  CalendarToday as CalendarIcon,
  Badge as RoleIcon,
} from '@mui/icons-material';
import { User } from '../../types/user';

interface UserProfileProps {
  user: User;
  onUpdateUser: (userData: Partial<User>) => Promise<void>;
  loading?: boolean;
  isEditable?: boolean;
}

const UserProfile: React.FC<UserProfileProps> = ({
  user,
  onUpdateUser,
  loading = false,
  isEditable = false,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editedUser, setEditedUser] = useState<Partial<User>>(user);
  const [saveError, setSaveError] = useState<string>('');
  const [saveSuccess, setSaveSuccess] = useState(false);

  useEffect(() => {
    setEditedUser(user);
  }, [user]);

  const handleEditToggle = () => {
    if (isEditing) {
      // Cancel editing
      setEditedUser(user);
      setSaveError('');
      setSaveSuccess(false);
    }
    setIsEditing(!isEditing);
  };

  const handleSave = async () => {
    try {
      setSaveError('');
      await onUpdateUser(editedUser);
      setSaveSuccess(true);
      setIsEditing(false);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'Failed to update user');
    }
  };

  const handleFieldChange = (field: keyof User, value: any) => {
    setEditedUser(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const getRoleColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return 'error';
      case 'manager':
        return 'warning';
      case 'user':
        return 'primary';
      default:
        return 'default';
    }
  };

  const getStatusColor = (isActive: boolean) => {
    return isActive ? 'success' : 'default';
  };

  return (
    <Box>
      {/* Header with Actions */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          User Profile
        </Typography>
        
        {isEditable && (
          <Box>
            {isEditing ? (
              <>
                <Button
                  startIcon={<CancelIcon />}
                  onClick={handleEditToggle}
                  sx={{ mr: 1 }}
                >
                  Cancel
                </Button>
                <Button
                  startIcon={<SaveIcon />}
                  variant="contained"
                  onClick={handleSave}
                  disabled={loading}
                >
                  {loading ? <CircularProgress size={24} /> : 'Save Changes'}
                </Button>
              </>
            ) : (
              <Button
                startIcon={<EditIcon />}
                variant="outlined"
                onClick={handleEditToggle}
              >
                Edit Profile
              </Button>
            )}
          </Box>
        )}
      </Box>

      {/* Alerts */}
      {saveError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {saveError}
        </Alert>
      )}
      
      {saveSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Profile updated successfully!
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Left Column - Profile Information */}
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, textAlign: 'center' }}>
            <Avatar
              src={user.avatar}
              sx={{
                width: 120,
                height: 120,
                mx: 'auto',
                mb: 2,
                border: 4,
                borderColor: 'background.paper',
                boxShadow: 1,
              }}
            >
              {user.name.charAt(0).toUpperCase()}
            </Avatar>
            
            <Typography variant="h5" gutterBottom>
              {user.name}
            </Typography>
            
            <Box mb={2}>
              <Chip
                label={user.role}
                color={getRoleColor(user.role)}
                sx={{ mr: 1 }}
              />
              <Chip
                label={user.isActive ? 'Active' : 'Inactive'}
                color={getStatusColor(user.isActive)}
                variant={user.isActive ? 'filled' : 'outlined'}
              />
            </Box>

            <Divider sx={{ my: 2 }} />

            {/* Quick Stats */}
            <Card variant="outlined" sx={{ mb: 2 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Quick Stats
                </Typography>
                <List dense>
                  <ListItem>
                    <ListItemText 
                      primary="Member since" 
                      secondary={new Date(user.createdAt).toLocaleDateString()}
                    />
                  </ListItem>
                  <ListItem>
                    <ListItemText 
                      primary="Last updated" 
                      secondary={new Date(user.updatedAt).toLocaleDateString()}
                    />
                  </ListItem>
                </List>
              </CardContent>
            </Card>
          </Paper>
        </Grid>

        {/* Right Column - Details */}
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Personal Information
            </Typography>
            
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Full Name"
                  value={isEditing ? editedUser.name : user.name}
                  onChange={(e) => handleFieldChange('name', e.target.value)}
                  disabled={!isEditing}
                  margin="normal"
                />
              </Grid>
              
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Email"
                  type="email"
                  value={isEditing ? editedUser.email : user.email}
                  onChange={(e) => handleFieldChange('email', e.target.value)}
                  disabled={!isEditing}
                  margin="normal"
                  InputProps={{
                    startAdornment: <EmailIcon sx={{ mr: 1, color: 'text.secondary' }} />,
                  }}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <FormControl fullWidth margin="normal" disabled={!isEditing}>
                  <InputLabel>Role</InputLabel>
                  <Select
                    value={isEditing ? editedUser.role : user.role}
                    label="Role"
                    onChange={(e) => handleFieldChange('role', e.target.value)}
                    startAdornment={<RoleIcon sx={{ mr: 1, color: 'text.secondary' }} />}
                  >
                    <MenuItem value="user">User</MenuItem>
                    <MenuItem value="manager">Manager</MenuItem>
                    <MenuItem value="admin">Admin</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item xs={12} sm={6}>
                <FormControl fullWidth margin="normal" disabled={!isEditing}>
                  <InputLabel>Status</InputLabel>
                  <Select
                    value={isEditing ? editedUser.isActive : user.isActive}
                    label="Status"
                    onChange={(e) => handleFieldChange('isActive', e.target.value)}
                  >
                    <MenuItem value="true">Active</MenuItem>
                    <MenuItem value="false">Inactive</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              {!isEditing && (
                <>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="Created At"
                      value={new Date(user.createdAt).toLocaleString()}
                      disabled
                      margin="normal"
                      InputProps={{
                        startAdornment: <CalendarIcon sx={{ mr: 1, color: 'text.secondary' }} />,
                      }}
                    />
                  </Grid>
                  
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="Updated At"
                      value={new Date(user.updatedAt).toLocaleString()}
                      disabled
                      margin="normal"
                      InputProps={{
                        startAdornment: <CalendarIcon sx={{ mr: 1, color: 'text.secondary' }} />,
                      }}
                    />
                  </Grid>
                </>
              )}
            </Grid>

            {/* Additional Information Section */}
            {!isEditing && user.department && (
              <>
                <Divider sx={{ my: 3 }} />
                <Typography variant="h6" gutterBottom>
                  Additional Information
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      label="Department"
                      value={user.department}
                      disabled
                      margin="normal"
                    />
                  </Grid>
                  {user.position && (
                    <Grid item xs={12} sm={6}>
                      <TextField
                        fullWidth
                        label="Position"
                        value={user.position}
                        disabled
                        margin="normal"
                      />
                    </Grid>
                  )}
                </Grid>
              </>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default UserProfile;