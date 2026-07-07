import React, { useState, useEffect } from 'react';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';

function App() {
  const [token, setToken] = useState(localStorage.getItem('token'));

  const handleLoginSuccess = () => {
    setToken(localStorage.getItem('token'));
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  return (
    <>
      {token ? (
        <Dashboard onLogout={handleLogout} />
      ) : (
        <Login onLoginSuccess={handleLoginSuccess} />
      )}
    </>
  );
}

export default App;