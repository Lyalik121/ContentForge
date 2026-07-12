import React, { useState } from 'react';

export default function Scenario2({ onLogout, onNavigate }) {
  const [niche, setNiche] = useState('');
  const [audience, setAudience] = useState('');
  const [requirements, setRequirements] = useState('');
  const [contentType, setContentType] = useState('posts'); // 'posts' | 'script'

  const [loading, setLoading] = useState(false);
  const [serverError, setServerError] = useState(null);
  const [posts, setPosts] = useState(null);   // result for 'posts'
  const [script, setScript] = useState(null); // result for 'script'

  const handleGenerate = async () => {
    if (!niche.trim() || !audience.trim() || !requirements.trim()) {
      alert('Будь ласка, заповніть усі поля!');
      return;
    }

    setLoading(true);
    setServerError(null);
    setPosts(null);
    setScript(null);

    try {
      const token = localStorage.getItem('token');
      const res = await fetch('https://contentforge-backend-mdzb.onrender.com/api/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          niche,
          audience,
          requirements,
          content_type: contentType, // backend: "script" -> script, else -> posts
        }),
      });

      const data = await res.json();

      if (res.ok && data.status === 'success') {
        if (contentType === 'script') {
          setScript(data.result);
        } else {
          setPosts(data.result);
        }
      } else {
        setServerError(data.message || `Помилка сервера: ${res.status}`);
      }
    } catch (err) {
      console.error(err);
      setServerError("Не вдалося зв'язатися з сервером.");
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('Текст скопійовано в буфер обміну! 📋');
  };

  // Small reusable card — same style as Dashboard result cards
  const ResultCard = ({ icon, title, colorClass, text }) => (
    <div className="bg-gray-950 border border-gray-800 rounded-2xl p-5 flex flex-col justify-between shadow-xl">
      <div>
        <div className={`flex items-center gap-2 mb-3 font-bold text-sm uppercase tracking-wider ${colorClass}`}>
          {icon} {title}
        </div>
        <p className="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap">{text}</p>
      </div>
      <button
        onClick={() => copyToClipboard(text)}
        className="mt-5 w-full bg-gray-900 hover:bg-indigo-950 text-gray-300 hover:text-indigo-300 border border-gray-800 hover:border-indigo-800 py-2 rounded-xl text-xs font-semibold transition"
      >
        Копіювати
      </button>
    </div>
  );

  return (
    <div className="min-h-screen bg-[#030712] text-gray-100 p-8 flex flex-col items-center">
      <div className="w-full max-w-4xl flex justify-between items-center mb-8 mt-4">
        <h1 className="text-3xl font-black tracking-wider bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent uppercase">
          ContentForge
        </h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => onNavigate('dashboard')}
            className="bg-gray-900 hover:bg-gray-800 text-gray-300 px-4 py-2 rounded-xl text-sm font-medium transition border border-gray-800 shadow-sm"
          >
            ← До завантаження
          </button>
          <button
            onClick={onLogout}
            className="bg-gray-900 hover:bg-gray-800 text-gray-300 px-4 py-2 rounded-xl text-sm font-medium transition border border-gray-800 shadow-sm"
          >
            Вийти
          </button>
        </div>
      </div>

      <div className="w-full max-w-xl bg-gray-950 border border-gray-800 rounded-2xl p-6 shadow-2xl mb-8">
        <h2 className="text-2xl font-bold text-center mb-6 tracking-tight">Генерація з нуля</h2>

        <div className="flex gap-2 mb-6 bg-gray-900/40 border border-gray-800 rounded-xl p-1">
          <button
            onClick={() => setContentType('posts')}
            className={`flex-1 py-2 rounded-lg text-sm font-semibold transition ${
              contentType === 'posts' ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            Пости для соцмереж
          </button>
          <button
            onClick={() => setContentType('script')}
            className={`flex-1 py-2 rounded-lg text-sm font-semibold transition ${
              contentType === 'script' ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            Сценарій відео
          </button>
        </div>

        <label className="block text-sm text-gray-400 mb-2">Ніша</label>
        <input
          type="text"
          value={niche}
          onChange={(e) => setNiche(e.target.value)}
          placeholder="Наприклад: фітнес для початківців"
          className="w-full bg-gray-900/40 border border-gray-800 focus:border-indigo-500/60 rounded-xl px-4 py-3 mb-4 text-sm text-gray-100 outline-none transition placeholder-gray-600"
        />

        <label className="block text-sm text-gray-400 mb-2">Цільова аудиторія</label>
        <input
          type="text"
          value={audience}
          onChange={(e) => setAudience(e.target.value)}
          placeholder="Наприклад: жінки 25–35 років"
          className="w-full bg-gray-900/40 border border-gray-800 focus:border-indigo-500/60 rounded-xl px-4 py-3 mb-4 text-sm text-gray-100 outline-none transition placeholder-gray-600"
        />

        <label className="block text-sm text-gray-400 mb-2">Вимоги до контенту</label>
        <textarea
          value={requirements}
          onChange={(e) => setRequirements(e.target.value)}
          placeholder="Наприклад: короткі Shorts, дружній тон, заклик до дії"
          rows={4}
          className="w-full bg-gray-900/40 border border-gray-800 focus:border-indigo-500/60 rounded-xl px-4 py-3 mb-6 text-sm text-gray-100 outline-none transition placeholder-gray-600 resize-none"
        />

        <button
          onClick={handleGenerate}
          disabled={loading}
          className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:bg-gray-900 disabled:text-gray-600 py-3 rounded-xl font-bold transition duration-200"
        >
          {loading ? 'Генерація... (може зайняти до хвилини)' : 'Згенерувати'}
        </button>

        {serverError && (
          <div className="mt-4 p-4 bg-rose-950/40 border border-rose-800 text-rose-300 rounded-xl text-xs leading-relaxed text-center">
            {serverError}
          </div>
        )}
      </div>

      {posts && (
        <div className="w-full max-w-5xl grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-4">
          <ResultCard icon="✈️" title="Telegram" colorClass="text-sky-400" text={posts.telegram} />
          <ResultCard icon="📸" title="Instagram" colorClass="text-pink-400" text={posts.instagram} />
          <ResultCard icon="🎵" title="TikTok" colorClass="text-cyan-400" text={posts.tiktok} />
          <ResultCard icon="🧵" title="Threads" colorClass="text-purple-400" text={posts.threads} />
        </div>
      )}

      {script && (
        <div className="w-full max-w-5xl grid grid-cols-1 sm:grid-cols-3 gap-6 mt-4">
          <ResultCard icon="🎣" title="Hook (перші 3 сек)" colorClass="text-amber-400" text={script.hook} />
          <ResultCard icon="🎬" title="Основна частина" colorClass="text-indigo-400" text={script.body} />
          <ResultCard icon="👉" title="Заклик до дії" colorClass="text-emerald-400" text={script.cta} />
        </div>
      )}
    </div>
  );
}