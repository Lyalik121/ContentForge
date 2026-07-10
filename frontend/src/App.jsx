import React, { useState } from 'react';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Scenario2 from './pages/Scenario2';

function App() {
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [page, setPage] = useState('dashboard'); // 'dashboard' | 'scenario2'

  const handleLoginSuccess = () => {
    setToken(localStorage.getItem('token'));
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setPage('dashboard');
  };

  if (!token) {
    return <Login onLoginSuccess={handleLoginSuccess} />;
  }

  if (page === 'scenario2') {
    return <Scenario2 onLogout={handleLogout} onNavigate={setPage} />;
  }

  return <Dashboard onLogout={handleLogout} onNavigate={setPage} />;
}

export default App;