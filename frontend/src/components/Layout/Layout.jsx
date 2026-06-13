import React from 'react';
import { useAuth } from '../../context/AuthContext';
import { LogOut, User as UserIcon, BookOpen, Store } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const Layout = ({ children }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/auth');
  };

  const getRoleLabel = (role) => {
    if (role === 'admin') return 'Администратор';
    return role === 'owner' ? 'Владелец' : 'Покупатель';
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-indigo-500 selection:text-white">
      {/* Background gradients for premium glassmorphism vibe */}
      <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] rounded-full bg-indigo-900/10 blur-[120px] pointer-events-none"></div>
      <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] rounded-full bg-emerald-900/10 blur-[120px] pointer-events-none"></div>

      {/* Header */}
      <header className="sticky top-0 z-50 backdrop-blur-md bg-slate-900/60 border-b border-slate-800/80">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          
          {/* Logo */}
          <div 
            onClick={() => {
              if (user?.role === 'admin') navigate('/admin');
              else navigate(user?.role === 'owner' ? '/owner' : '/customer');
            }} 
            className="flex items-center space-x-3 cursor-pointer group"
          >
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-indigo-400 flex items-center justify-center shadow-lg shadow-indigo-500/20 group-hover:scale-105 transition-transform duration-200">
              <BookOpen className="w-5 h-5 text-white" />
            </div>
            <div>
              <span className="font-bold text-lg bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-emerald-400">
                Дәптер.kz
              </span>
              <span className="block text-[10px] text-slate-400 font-medium tracking-wider uppercase">
                Цифровая Тетрадь
              </span>
            </div>
          </div>

          {/* User profile actions */}
          {user && (
            <div className="flex items-center space-x-6">
              {/* Profile summary */}
              <div className="hidden sm:flex items-center space-x-3 border-r border-slate-800/80 pr-6">
                <div className="w-9 h-9 rounded-lg bg-slate-800 flex items-center justify-center text-slate-300">
                  <UserIcon className="w-4 h-4" />
                </div>
                <div className="text-left">
                  <span className="block text-xs font-semibold text-slate-200">
                    {user.full_name || 'Пользователь'}
                  </span>
                  <div className="flex items-center space-x-2">
                    <span className="text-[10px] text-slate-400 font-medium">
                      {user.phone}
                    </span>
                    <span className={`text-[9px] px-1.5 py-0.5 rounded font-bold uppercase tracking-wider ${
                      user.role === 'admin'
                        ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                        : user.role === 'owner' 
                          ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20' 
                          : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                    }`}>
                      {getRoleLabel(user.role)}
                    </span>
                  </div>
                </div>
              </div>

              {/* Logout button */}
              <button 
                onClick={handleLogout}
                className="flex items-center space-x-2 text-slate-400 hover:text-rose-400 bg-slate-900 hover:bg-rose-500/10 border border-slate-800 hover:border-rose-500/20 px-3.5 py-1.8 rounded-xl text-sm font-medium transition-all duration-200"
              >
                <LogOut className="w-4 h-4" />
                <span className="hidden md:inline">Выйти</span>
              </button>
            </div>
          )}
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 relative z-10">
        {children}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 bg-slate-950/80 text-center py-6 text-xs text-slate-500">
        <div className="max-w-7xl mx-auto px-4">
          <p>© {new Date().getFullYear()} Дәптер.kz — Микрофинансовая система локального кредитования.</p>
          <p className="mt-1 text-slate-600">Разработано для цифровизации сельских магазинов.</p>
        </div>
      </footer>
    </div>
  );
};

export default Layout;
