import React, { useState, useEffect } from 'react';

export default function Auth() {
  const [isLoginMode, setIsLoginMode] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (localStorage.getItem('token')) {
      window.location.href = '/';
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !password) {
      alert('Будь ласка, заповніть усі поля!');
      return;
    }

    setLoading(true);
    const endpoint = isLoginMode ? '/api/auth/login' : '/api/auth/register';
    const targetUrl = `https://contentforge-backend-mdzb.onrender.com${endpoint}`;

    try {
      const res = await fetch(targetUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        if (isLoginMode) {
          localStorage.setItem('token', data.token);
          window.location.href = '/';
        } else {
          alert('Реєстрація успішна! Тепер ви можете увійти.');
          setIsLoginMode(true);
          setPassword('');
        }
      } else {
        alert(data.message || `Помилка сервера: ${res.status}`);
      }
    } catch (err) {
      console.error(err);
      alert("Не вдалося зв'язатися з сервером.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#030712] text-gray-100 flex flex-col items-center justify-center p-6">
      <div className="w-full max-w-md bg-gray-950 border border-gray-800 rounded-2xl p-8 shadow-2xl">
        <h1 className="text-3xl font-black text-center tracking-wider bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent uppercase mb-2">
          ContentForge
        </h1>
        <p className="text-gray-400 text-sm text-center mb-8">
          {isLoginMode ? 'Увійдіть у свій акаунт' : 'Створіть новий акаунт для роботи'}
        </p>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-gray-400 mb-2">
              Email адреса
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@company.com"
              className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-sm text-gray-100 placeholder-gray-600 focus:outline-none focus:border-indigo-500 transition"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-gray-400 mb-2">
              Пароль
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full bg-gray-900 border border-gray-800 rounded-xl px-4 py-3 text-sm text-gray-100 placeholder-gray-600 focus:outline-none focus:border-indigo-500 transition"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:bg-gray-900 disabled:text-gray-600 py-3 rounded-xl font-bold tracking-wide transition duration-200 mt-2"
          >
            {loading ? 'Обробка...' : isLoginMode ? 'Увійти' : 'Зареєструватися'}
          </button>
        </form>

        <div className="mt-6 text-center">
          <button
            onClick={() => {
              setIsLoginMode(!isLoginMode);
              setPassword('');
            }}
            className="text-sm text-indigo-400 hover:text-indigo-300 font-medium transition duration-150"
          >
            {isLoginMode ? 'Немає акаунту? Зареєструватися' : 'Вже є акаунт? Увійти'}
          </button>
        </div>
      </div>
    </div>
  );
}