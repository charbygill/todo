import React, { useState, useEffect } from 'react';
import axios from 'axios';
import {
  Container,
  Typography,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Paper,
  Box,
  Snackbar,
  Alert,
  Stack,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

function App() {
  const [todos, setTodos] = useState([]);
  const [newTodo, setNewTodo] = useState({
    title: '',
    date: new Date().toISOString().split('T')[0], // Today's date as default
    time: '12:00', // Noon as default
  });
  const [error, setError] = useState('');

  useEffect(() => {
    fetchTodos();
  }, []);

  const fetchTodos = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/todos');
      const sortedTodos = response.data.sort((a, b) => new Date(a.due_date) - new Date(b.due_date));
      setTodos(sortedTodos);
    } catch (error) {
      setError('Failed to fetch todos');
      console.error('Error fetching todos:', error);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      // Combine date and time into a single ISO string
      const dueDate = new Date(`${newTodo.date}T${newTodo.time}`);
      
      const todoToSubmit = {
        title: newTodo.title,
        due_date: dueDate.toISOString(),
        status: 'pending'
      };
      
      await axios.post('http://localhost:8080/api/todos', todoToSubmit);
      setNewTodo({
        title: '',
        date: new Date().toISOString().split('T')[0],
        time: '12:00'
      });
      fetchTodos();
    } catch (error) {
      setError('Failed to create todo');
      console.error('Error creating todo:', error);
    }
  };

  const formatDueDate = (dateString) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    }).format(date);
  };

  const handleDelete = async (id) => {
    try {
      await axios.delete(`http://localhost:8080/api/todos/${id}`);
      fetchTodos();
    } catch (error) {
      setError('Failed to delete todo');
      console.error('Error deleting todo:', error);
    }
  };

  const handleCloseError = () => {
    setError('');
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Todo List
        </Typography>

        <form onSubmit={handleSubmit}>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mb: 3 }}>
            <TextField
              fullWidth
              label="Title"
              value={newTodo.title}
              onChange={(e) => setNewTodo({ ...newTodo, title: e.target.value })}
              required
            />
            <Stack direction="row" spacing={2}>
              <TextField
                type="date"
                label="Due Date"
                value={newTodo.date}
                onChange={(e) => setNewTodo({ ...newTodo, date: e.target.value })}
                InputLabelProps={{ shrink: true }}
                required
                sx={{ flex: 2 }}
              />
              <TextField
                type="time"
                label="Time"
                value={newTodo.time}
                onChange={(e) => setNewTodo({ ...newTodo, time: e.target.value })}
                InputLabelProps={{ shrink: true }}
                required
                sx={{ flex: 1 }}
              />
            </Stack>
            <Button 
              variant="contained" 
              color="primary" 
              type="submit"
              sx={{ alignSelf: 'flex-start' }}
            >
              Add Todo
            </Button>
          </Box>
        </form>
      </Paper>

      <Paper elevation={3} sx={{ p: 3 }}>
        <List>
          {todos.length === 0 ? (
            <Typography color="textSecondary" align="center">
              No todos yet. Add one above!
            </Typography>
          ) : (
            todos.map((todo) => (
              <ListItem 
                key={todo.id}
                sx={{
                  bgcolor: new Date(todo.due_date) < new Date() ? '#fff4f4' : 'inherit',
                  borderRadius: 1,
                  mb: 1
                }}
              >
                <ListItemText
                  primary={todo.title}
                  secondary={
                    <Typography 
                      component="span" 
                      variant="body2" 
                      color={new Date(todo.due_date) < new Date() ? 'error' : 'textSecondary'}
                    >
                      Due: {formatDueDate(todo.due_date)}
                    </Typography>
                  }
                />
                <ListItemSecondaryAction>
                  <IconButton edge="end" onClick={() => handleDelete(todo.id)}>
                    <DeleteIcon />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))
          )}
        </List>
      </Paper>

      <Snackbar open={!!error} autoHideDuration={6000} onClose={handleCloseError}>
        <Alert onClose={handleCloseError} severity="error" sx={{ width: '100%' }}>
          {error}
        </Alert>
      </Snackbar>
    </Container>
  );
}

export default App;
