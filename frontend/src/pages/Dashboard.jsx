import React, { useState, useEffect } from 'react';

export default function Dashboard({ onNavigate, onLogout }) {
  const [file, setFile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [mediaID, setMediaID] = useState(null);
  const [status, setStatus] = useState(null);
  const [generatedPosts, setGeneratedPosts] = useState(null);
  const [serverError, setServerError] = useState(null);

  useEffect(() => {
    if (window.location.pathname === '/login' && localStorage.getItem('token')) {
      window.history.replaceState({}, document.title, "/");
    }
  }, []);

  useEffect(() => {
    if (!mediaID) return;

    const interval = setInterval(async () => {
      try {
        const token = localStorage.getItem('token');
        const targetUrl = `https://contentforge-backend-mdzb.onrender.com/api/media/${mediaID}/status`;
        
        const res = await fetch(targetUrl, {
          method: 'GET',
          headers: { 'Authorization': `Bearer ${token}` }
        });

        if (res.status === 404) {
          setServerError(`Помилка 404: Маршрут не знайдено на сервері.`);
          clearInterval(interval);
          return;
        }

        const contentType = res.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
          setServerError("Сервер повернув відповідь не в форматі JSON.");
          clearInterval(interval);
          return;
        }

        const data = await res.json();

        if (res.ok && data.media_status) {
          setServerError(null);
          setStatus(data.media_status);

          const currentStatus = data.media_status.toLowerCase();

          if (currentStatus === 'completed') {
            clearInterval(interval);
            
            setGeneratedPosts({
              telegram: data.telegram || "Текст для Telegram відсутній у відповіді сервера.",
              instagram: data.instagram || "Текст для Instagram відсутній у відповіді сервера.",
              tiktok: data.tiktok || "Текст для TikTok відсутній у відповіді сервера.",
              threads: data.threads || "Текст для Threads відсутній у відповіді сервера."
            });
          } else if (currentStatus === 'failed') {
            clearInterval(interval);
          }
        }
      } catch (err) {
        console.error(err);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [mediaID]);

  const handleLogout = () => {
  if (typeof onLogout === 'function') {
    onLogout();
  }
  };

  const handleUpload = async () => {
    if (!file) {
      alert("Будь ласка, спочатку виберіть відеофайл!");
      return;
    }
    setLoading(true);
    setGeneratedPosts(null);
    setServerError(null);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const token = localStorage.getItem('token');
      const res = await fetch('https://contentforge-backend-mdzb.onrender.com/api/media/upload', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData,
      });

      const data = await res.json();
      
      if (res.ok && data.media_id) {
        setMediaID(data.media_id);
        setStatus('Uploaded');
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

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert("Текст скопійовано в буфер обміну! 📋");
  };

  return (
    <div className="min-h-screen bg-[#030712] text-gray-100 p-8 flex flex-col items-center">
      <div className="w-full max-w-4xl flex justify-between items-center mb-8 mt-4">
        <h1 className="text-3xl font-black tracking-wider bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent uppercase">
          ContentForge
        </h1>
       <div className="flex items-center gap-2">
          <button
            onClick={() => onNavigate('scenario2')}
            className="bg-indigo-950 hover:bg-indigo-900 text-indigo-300 px-4 py-2 rounded-xl text-sm font-medium transition border border-indigo-800 shadow-sm"
          >
            Генерація з нуля
          </button>
          <button 
            onClick={handleLogout}
            className="bg-gray-900 hover:bg-gray-800 text-gray-300 px-4 py-2 rounded-xl text-sm font-medium transition border border-gray-800 shadow-sm"
          >
            Вийти
          </button>
        </div>
      </div>

      <div className="w-full max-w-xl bg-gray-950 border border-gray-800 rounded-2xl p-6 shadow-2xl mb-8">
        <h2 className="text-2xl font-bold text-center mb-6 tracking-tight">Генератор контенту</h2>
        
        <div className="border-2 border-dashed border-gray-800 hover:border-indigo-500/40 rounded-xl p-8 text-center transition bg-gray-900/20">
          <input 
            type="file" 
            accept="video/*,audio/*"
            onChange={(e) => setFile(e.target.files[0])} 
            className="block w-full text-sm text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-indigo-950 file:text-indigo-300 hover:file:bg-indigo-900 mb-4 cursor-pointer" 
          />
          <button 
            onClick={handleUpload} 
            disabled={loading} 
            className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:bg-gray-900 disabled:text-gray-600 py-3 rounded-xl font-bold transition duration-200"
          >
            {loading ? "Завантаження файлу..." : "Запустити конвеєр обробки"}
          </button>
        </div>

        {serverError && (
          <div className="mt-4 p-4 bg-rose-950/40 border border-rose-800 text-rose-300 rounded-xl text-xs leading-relaxed text-center">
            {serverError}
          </div>
        )}

        {mediaID && !serverError && (
          <div className="mt-6 bg-gray-900/40 border border-gray-800 rounded-xl p-4 flex flex-col items-center justify-between gap-2 sm:flex-row">
            <span className="text-sm text-gray-400">ID Медіа: <strong className="text-gray-200 font-mono">{mediaID}</strong></span>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-400">Статус:</span>
              <span className={`px-3 py-1 text-xs font-bold rounded-full border tracking-wider uppercase ${
                status?.toLowerCase() === 'completed' ? 'bg-emerald-950/80 text-emerald-400 border-emerald-800' :
                status?.toLowerCase() === 'failed' ? 'bg-rose-950/80 text-rose-400 border-rose-800' :
                'bg-indigo-950/80 text-indigo-400 border-indigo-800 animate-pulse'
              }`}>
                {status}
              </span>
            </div>
          </div>
        )}
      </div>

      {generatedPosts && (
        <div className="w-full max-w-5xl grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-4">
          <div className="bg-gray-950 border border-gray-800 rounded-2xl p-5 flex flex-col justify-between shadow-xl">
            <div>
              <div className="flex items-center gap-2 mb-3 text-sky-400 font-bold text-sm uppercase tracking-wider">
                ✈️ Telegram
              </div>
              <p className="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap">{generatedPosts.telegram}</p>
            </div>
            <button 
              onClick={() => copyToClipboard(generatedPosts.telegram)}
              className="mt-5 w-full bg-gray-900 hover:bg-indigo-950 text-gray-300 hover:text-indigo-300 border border-gray-800 hover:border-indigo-800 py-2 rounded-xl text-xs font-semibold transition"
            >
              Копіювати
            </button>
          </div>

          <div className="bg-gray-950 border border-gray-800 rounded-2xl p-5 flex flex-col justify-between shadow-xl">
            <div>
              <div className="flex items-center gap-2 mb-3 text-pink-400 font-bold text-sm uppercase tracking-wider">
                📸 Instagram
              </div>
              <p className="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap">{generatedPosts.instagram}</p>
            </div>
            <button 
              onClick={() => copyToClipboard(generatedPosts.instagram)}
              className="mt-5 w-full bg-gray-900 hover:bg-indigo-950 text-gray-300 hover:text-indigo-300 border border-gray-800 hover:border-indigo-800 py-2 rounded-xl text-xs font-semibold transition"
            >
              Копіювати
            </button>
          </div>

          <div className="bg-gray-950 border border-gray-800 rounded-2xl p-5 flex flex-col justify-between shadow-xl">
            <div>
              <div className="flex items-center gap-2 mb-3 text-cyan-400 font-bold text-sm uppercase tracking-wider">
                🎵 TikTok
              </div>
              <p className="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap">{generatedPosts.tiktok}</p>
            </div>
            <button 
              onClick={() => copyToClipboard(generatedPosts.tiktok)}
              className="mt-5 w-full bg-gray-900 hover:bg-indigo-950 text-gray-300 hover:text-indigo-300 border border-gray-800 hover:border-indigo-800 py-2 rounded-xl text-xs font-semibold transition"
            >
              Копіювати
            </button>
          </div>

          <div className="bg-gray-950 border border-gray-800 rounded-2xl p-5 flex flex-col justify-between shadow-xl">
            <div>
              <div className="flex items-center gap-2 mb-3 text-purple-400 font-bold text-sm uppercase tracking-wider">
                🧵 Threads
              </div>
              <p className="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap">{generatedPosts.threads}</p>
            </div>
            <button 
              onClick={() => copyToClipboard(generatedPosts.threads)}
              className="mt-5 w-full bg-gray-900 hover:bg-indigo-950 text-gray-300 hover:text-indigo-300 border border-gray-800 hover:border-indigo-800 py-2 rounded-xl text-xs font-semibold transition"
            >
              Копіювати
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
