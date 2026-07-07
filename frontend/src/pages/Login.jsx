import React, { useState } from 'react';

export default function Login({ onLoginSuccess }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await fetch('http://localhost:3000/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Щось пішло не так');
      }

      localStorage.setItem('token', data.token);
      alert('Успішний вхід! Токен збережено.');
      
      if (onLoginSuccess) {
        onLoginSuccess();
      }
      
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white p-4">
      <div className="bg-gray-800 p-8 rounded-lg shadow-xl w-full max-w-md border border-gray-700">
        <h2 className="text-3xl font-bold mb-6 text-center text-indigo-400 font-mono">ContentForge</h2>
        <p className="text-gray-400 text-sm text-center mb-6">Вхід у кабінет розробника</p>

        {error && (
          <div className="bg-red-500/20 border border-red-500 text-red-400 p-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-gray-400 uppercase mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full p-3 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-indigo-500 text-white"
              placeholder="testuser@gmail.com"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-400 uppercase mb-1">Пароль</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full p-3 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-indigo-500 text-white"
              placeholder="••••••••"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold p-3 rounded transition duration-200 disabled:opacity-50 mt-2"
          >
            {loading ? 'Вхід...' : 'Увійти'}
          </button>
        </form>
      </div>
    </div>
  );
}