import React, { useState } from 'react';

export default function Dashboard({ onLogout }) {
  const [file, setFile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
    setMessage('');
    setError('');
  };

  const handleUpload = async (e) => {
    e.preventDefault();
    if (!file) {
      setError('Будь ласка, виберіть файл відео');
      return;
    }

    setLoading(true);
    setError('');
    setMessage('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      const token = localStorage.getItem('token');

      const response = await fetch('http://localhost:3000/api/media/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Помилка під час завантаження відео');
      }

      setMessage(`Відео успішно завантажено! ID файлу в базі: ${data.id || 'створено'}. Бекенд почав обробку.`);
      setFile(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col font-sans">
      {/* Навігаційна панель */}
      <header className="bg-gray-900 border-b border-gray-800 p-4 flex justify-between items-center px-8">
        <h1 className="text-2xl font-bold font-mono text-indigo-400 m-0">ContentForge 🤔</h1>
        <button 
          onClick={onLogout}
          className="bg-gray-800 hover:bg-red-900 text-xs text-gray-300 hover:text-white px-4 py-2 rounded transition border border-gray-700"
        >
          Вийти
        </button>
      </header>

      {/* Основна секція */}
      <main className="flex-1 flex flex-col items-center justify-center p-6">
        <div className="bg-gray-900 border border-gray-800 rounded-xl p-8 max-w-xl w-full shadow-2xl text-center">
          <h2 className="text-2xl font-bold text-white mb-2">Генератор контенту</h2>
          <p className="text-gray-400 text-sm mb-6">Завантажте відео-тест, і ШІ автоматично розпізнає мову та згенерує пости</p>

          {message && (
            <div className="bg-emerald-500/10 border border-emerald-500 text-emerald-400 p-4 rounded-lg mb-6 text-sm text-left">
              {message}
            </div>
          )}

          {error && (
            <div className="bg-red-500/10 border border-red-500 text-red-400 p-4 rounded-lg mb-6 text-sm text-left">
              {error}
            </div>
          )}

          <form onSubmit={handleUpload} className="space-y-6">
            <div className="border-2 border-dashed border-gray-700 hover:border-indigo-500 rounded-lg p-8 cursor-pointer transition bg-gray-950/50 relative">
              <input 
                type="file" 
                accept="video/*"
                onChange={handleFileChange}
                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
              />
              <div className="space-y-2">
                <span className="block text-4xl">📁</span>
                <span className="block text-sm text-gray-300 font-medium">
                  {file ? file.name : 'Клацніть або перетягніть файл відео сюди'}
                </span>
                <span className="block text-xs text-gray-500">MP4, MKV або MOV (до 500MB)</span>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || !file}
              className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:hover:bg-indigo-600 text-white font-bold py-3 px-6 rounded-lg transition duration-150 shadow-lg shadow-indigo-600/20"
            >
              {loading ? 'Надсилання файлу та запуск ШІ...' : 'Запустити конвеєр обробки'}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}