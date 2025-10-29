import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Box,
  Typography,
  Avatar,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Chip,
  CircularProgress,
  Alert,
} from '@mui/material';
import { User } from '../types/user';

interface TaskAssignmentProps {
  open: boolean;
  onClose: () => void;
  onAssign: (userId: string) => void;
  taskId: string;
  taskTitle: string;
  currentAssignee?: User;
  users: User[];
  loading?: boolean;
}

const TaskAssignment: React.FC<TaskAssignmentProps> = ({
  open,
  onClose,
  onAssign,
  taskId,
  taskTitle,
  currentAssignee,
  users,
  loading = false,
}) => {
  const [selectedUserId, setSelectedUserId] = useState<string>('');
  const [filteredUsers, setFilteredUsers] = useState<User[]>([]);

  useEffect(() => {
    if (currentAssignee) {
      setSelectedUserId(currentAssignee.id.toString());
    } else {
      setSelectedUserId('');
    }
  }, [currentAssignee, open]);

  useEffect(() => {
    setFilteredUsers(users);
  }, [users]);

  const handleAssign = () => {
    if (selectedUserId) {
      onAssign(selectedUserId);
    }
  };

  const handleClose = () => {
    setSelectedUserId(currentAssignee?.id.toString() || '');
    onClose();
  };

  const getAssigneeInfo = (userId: string) => {
    return users.find(user => user.id.toString() === userId);
  };

  return (
    <Dialog 
      open={open} 
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle>
        Assign Task
        <Typography variant="subtitle1" color="text.secondary">
          {taskTitle}
        </Typography>
      </DialogTitle>
      
      <DialogContent>
        {loading ? (
          <Box display="flex" justifyContent="center" p={3}>
            <CircularProgress />
          </Box>
        ) : (
          <Box>
            {/* Current Assignee Display */}
            {currentAssignee && (
              <Box mb={3}>
                <Typography variant="subtitle2" gutterBottom>
                  Current Assignee:
                </Typography>
                <Chip
                  avatar={<Avatar src={currentAssignee.avatar} />}
                  label={`${currentAssignee.name} (${currentAssignee.email})`}
                  variant="outlined"
                  color="primary"
                />
              </Box>
            )}

            {/* Assignee Selection */}
            <FormControl fullWidth variant="outlined">
              <InputLabel id="assignee-select-label">Select Assignee</InputLabel>
              <Select
                labelId="assignee-select-label"
                value={selectedUserId}
                onChange={(e) => setSelectedUserId(e.target.value)}
                label="Select Assignee"
              >
                <MenuItem value="">
                  <em>Unassigned</em>
                </MenuItem>
                {filteredUsers.map((user) => (
                  <MenuItem key={user.id} value={user.id.toString()}>
                    <Box display="flex" alignItems="center">
                      <Avatar 
                        src={user.avatar} 
                        sx={{ width: 24, height: 24, mr: 2 }}
                      >
                        {user.name.charAt(0)}
                      </Avatar>
                      <Box>
                        <Typography variant="body1">{user.name}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {user.email}
                        </Typography>
                      </Box>
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* Selected User Preview */}
            {selectedUserId && (
              <Box mt={3} p={2} sx={{ backgroundColor: 'action.hover', borderRadius: 1 }}>
                <Typography variant="subtitle2" gutterBottom>
                  Selected Assignee:
                </Typography>
                {(() => {
                  const user = getAssigneeInfo(selectedUserId);
                  return user ? (
                    <List dense>
                      <ListItem>
                        <ListItemAvatar>
                          <Avatar src={user.avatar}>{user.name.charAt(0)}</Avatar>
                        </ListItemAvatar>
                        <ListItemText
                          primary={user.name}
                          secondary={
                            <Box>
                              <Typography variant="body2">{user.email}</Typography>
                              {user.role && (
                                <Chip
                                  label={user.role}
                                  size="small"
                                  color={user.role === 'admin' ? 'secondary' : 'default'}
                                  sx={{ mt: 0.5 }}
                                />
                              )}
                            </Box>
                          }
                        />
                      </ListItem>
                    </List>
                  ) : null;
                })()}
              </Box>
            )}

            {/* No Users Available */}
            {users.length === 0 && !loading && (
              <Alert severity="warning" sx={{ mt: 2 }}>
                No users available for assignment.
              </Alert>
            )}
          </Box>
        )}
      </DialogContent>
      
      <DialogActions>
        <Button onClick={handleClose}>
          Cancel
        </Button>
        <Button 
          onClick={handleAssign}
          variant="contained"
          disabled={!selectedUserId || loading}
        >
          Assign Task
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default TaskAssignment;