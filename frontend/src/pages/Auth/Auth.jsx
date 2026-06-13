import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { ShieldCheck, Phone, Lock, User, FileText, KeyRound } from 'lucide-react';

const Auth = () => {
  const { 
    user, 
    needsPin, 
    login, 
    registerOwner, 
    registerCustomerInitiate, 
    registerCustomerVerify, 
    setCustomerPin 
  } = useAuth();
  
  const navigate = useNavigate();

  // Состояние экранов: 'login' | 'register' | 'otp' | 'pin'
  const [view, setView] = useState('login');
  // Роль для входа: 'owner' | 'customer'
  const [loginRole, setLoginRole] = useState('owner');
  
  // Общие поля
  const [phone, setPhone] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  // Поля входа владельца / покупателя
  const [password, setPassword] = useState('');
  const [pin, setPin] = useState('');

  // Поля регистрации
  const [regRole, setRegRole] = useState('customer'); // 'customer' или 'owner'
  const [fullName, setFullName] = useState('');
  const [iin, setIin] = useState('');
  const [otpCode, setOtpCode] = useState('');
  const [newPin, setNewPin] = useState('');
  const [confirmPin, setConfirmPin] = useState('');

  useEffect(() => {
    if (user) {
      if (needsPin) {
        setView('pin');
      } else {
        if (user.role === 'admin') {
          navigate('/admin');
        } else {
          navigate(user.role === 'owner' ? '/owner' : '/customer');
        }
      }
    }
  }, [user, needsPin, navigate]);

  const resetState = () => {
    setError('');
    setMessage('');
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    resetState();
    setLoading(true);
    
    try {
      const credentials = (loginRole === 'owner' || loginRole === 'admin') ? password : pin;
      if (!phone || !credentials) {
        throw new Error('Пожалуйста, заполните все поля входа');
      }
      await login(phone, credentials);
      // Перенаправление произойдет в useEffect
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterInitiate = async (e) => {
    e.preventDefault();
    resetState();
    setLoading(true);

    try {
      if (!phone || !iin || !fullName) {
        throw new Error('Пожалуйста, заполните все поля');
      }
      if (iin.length !== 12) {
        throw new Error('ИИН должен состоять ровно из 12 цифр');
      }

      if (regRole === 'owner') {
        if (!password) {
          throw new Error('Укажите пароль для регистрации владельца');
        }
        await registerOwner(phone, password, iin, fullName);
      } else {
        await registerCustomerInitiate(phone, iin, fullName);
        setMessage('Код подтверждения отправлен. Пожалуйста, посмотрите логи сервера Go!');
        setView('otp');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOtp = async (e) => {
    e.preventDefault();
    resetState();
    setLoading(true);

    try {
      if (!otpCode) {
        throw new Error('Пожалуйста, введите код из SMS');
      }
      await registerCustomerVerify(phone, otpCode);
      setView('pin'); // После успешной валидации OTP переводим на установку PIN
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSetPin = async (e) => {
    e.preventDefault();
    resetState();
    setLoading(true);

    try {
      if (newPin.length !== 4 || isNaN(newPin)) {
        throw new Error('PIN-код должен состоять ровно из 4-х цифр');
      }
      if (newPin !== confirmPin) {
        throw new Error('PIN-коды не совпадают');
      }
      await setCustomerPin(newPin);
      navigate('/customer');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center py-12 sm:px-6 lg:px-8 relative selection:bg-indigo-500 selection:text-white">
      {/* Decorative Blur Backgrounds */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[350px] h-[350px] bg-indigo-500/10 blur-[90px] rounded-full pointer-events-none"></div>
      <div className="absolute bottom-1/4 left-1/3 w-[250px] h-[250px] bg-emerald-500/5 blur-[90px] rounded-full pointer-events-none"></div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md relative z-10">
        <div className="flex justify-center items-center space-x-3 mb-6">
          <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-indigo-600 to-indigo-400 flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <ShieldCheck className="w-6 h-6 text-white" />
          </div>
          <span className="font-extrabold text-2xl bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-emerald-400">
            Дәптер.kz
          </span>
        </div>
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md relative z-10 px-4">
        <div className="bg-slate-900/40 backdrop-blur-xl border border-slate-800/80 py-8 px-6 shadow-2xl rounded-3xl sm:px-10">
          
          {/* Заголовки */}
          <div className="mb-6 text-center">
            <h2 className="text-xl font-bold text-slate-100">
              {view === 'login' && 'Вход в личный кабинет'}
              {view === 'register' && 'Регистрация пользователя'}
              {view === 'otp' && 'Подтверждение телефона'}
              {view === 'pin' && 'Установка PIN-кода'}
            </h2>
            <p className="mt-1.5 text-xs text-slate-400">
              {view === 'login' && 'Введите свои данные для доступа к долговому журналу'}
              {view === 'register' && 'Введите личные данные для регистрации'}
              {view === 'otp' && `Код отправлен на номер ${phone}`}
              {view === 'pin' && 'Придумайте 4-значный PIN-код для последующего входа'}
            </p>
          </div>

          {/* Информационные сообщения / ошибки */}
          {error && (
            <div className="mb-4 p-3.5 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-400 text-xs font-medium">
              {error}
            </div>
          )}
          {message && (
            <div className="mb-4 p-3.5 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400 text-xs font-medium">
              {message}
            </div>
          )}

          {/* VIEW: LOGIN */}
          {view === 'login' && (
            <form onSubmit={handleLogin} className="space-y-4">
              {/* Переключатель ролей */}
              <div className="grid grid-cols-3 p-1 bg-slate-950 rounded-xl border border-slate-800/80 mb-2">
                <button
                  type="button"
                  onClick={() => { setLoginRole('owner'); resetState(); }}
                  className={`py-2 text-[10px] font-semibold rounded-lg transition-all ${
                    loginRole === 'owner' 
                      ? 'bg-slate-800 text-slate-100 shadow-sm' 
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Владелец
                </button>
                <button
                  type="button"
                  onClick={() => { setLoginRole('customer'); resetState(); }}
                  className={`py-2 text-[10px] font-semibold rounded-lg transition-all ${
                    loginRole === 'customer' 
                      ? 'bg-slate-800 text-slate-100 shadow-sm' 
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Покупатель
                </button>
                <button
                  type="button"
                  onClick={() => { setLoginRole('admin'); resetState(); }}
                  className={`py-2 text-[10px] font-semibold rounded-lg transition-all ${
                    loginRole === 'admin' 
                      ? 'bg-slate-800 text-slate-100 shadow-sm' 
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Админ
                </button>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Номер телефона</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <Phone className="w-4 h-4" />
                  </span>
                  <input
                    type="tel"
                    id="login-phone"
                    name="phone"
                    autoComplete="tel"
                    placeholder="Например, +77071234567"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              {(loginRole === 'owner' || loginRole === 'admin') ? (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Пароль</label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                      <Lock className="w-4 h-4" />
                    </span>
                    <input
                      type="password"
                      id="login-password"
                      name="password"
                      autoComplete="current-password"
                      placeholder="••••••••"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                    />
                  </div>
                </div>
              ) : (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">4-значный PIN-код</label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                      <KeyRound className="w-4 h-4" />
                    </span>
                    <input
                      type="password"
                      id="login-pin"
                      name="pin"
                      autoComplete="one-time-code"
                      maxLength="4"
                      placeholder="••••"
                      value={pin}
                      onChange={(e) => setPin(e.target.value)}
                      className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 tracking-[0.3em] focus:outline-none focus:border-indigo-500 transition-colors"
                    />
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-indigo-600/20 active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {loading ? 'Вход...' : 'Войти'}
              </button>

              <div className="mt-4 text-center">
                <span className="text-xs text-slate-500">Еще нет аккаунта? </span>
                <button
                  type="button"
                  onClick={() => { setView('register'); resetState(); }}
                  className="text-xs text-indigo-400 hover:text-indigo-300 font-semibold"
                >
                  Зарегистрироваться
                </button>
              </div>
            </form>
          )}

          {/* VIEW: REGISTER */}
          {view === 'register' && (
            <form onSubmit={handleRegisterInitiate} className="space-y-4">
              {/* Публичная регистрация доступна только для покупателей */}

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">ФИО полностью</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <User className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    id="register-name"
                    name="name"
                    autoComplete="name"
                    placeholder="Например, Ахметов Серик"
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">ИИН (12 цифр)</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <FileText className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    id="register-iin"
                    name="iin"
                    autoComplete="off"
                    maxLength="12"
                    placeholder="Введите 12 цифр ИИН"
                    value={iin}
                    onChange={(e) => setIin(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Номер телефона</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <Phone className="w-4 h-4" />
                  </span>
                  <input
                    type="tel"
                    id="register-phone"
                    name="phone"
                    autoComplete="tel"
                    placeholder="+77071234567"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              {regRole === 'owner' && (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Пароль владельца</label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                      <Lock className="w-4 h-4" />
                    </span>
                    <input
                      type="password"
                      id="register-password"
                      name="password"
                      autoComplete="new-password"
                      placeholder="Придумайте пароль"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                    />
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-emerald-600/20 active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {loading ? 'Регистрация...' : regRole === 'owner' ? 'Зарегистрироваться' : 'Получить SMS-код'}
              </button>

              <div className="mt-4 text-center">
                <span className="text-xs text-slate-500">Уже есть аккаунт? </span>
                <button
                  type="button"
                  onClick={() => { setView('login'); resetState(); }}
                  className="text-xs text-indigo-400 hover:text-indigo-300 font-semibold"
                >
                  Войти
                </button>
              </div>
            </form>
          )}

          {/* VIEW: OTP */}
          {view === 'otp' && (
            <form onSubmit={handleVerifyOtp} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Код из SMS-заглушки</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <KeyRound className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    id="otp-code"
                    name="otp"
                    autoComplete="one-time-code"
                    maxLength="4"
                    placeholder="Введите 4 цифры"
                    value={otpCode}
                    onChange={(e) => setOtpCode(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 tracking-[0.5em] focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <span className="block mt-2 text-[10px] text-amber-400/80 font-medium">
                  * Пожалуйста, скопируйте код из консоли бэкенда!
                </span>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-indigo-600/20 active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {loading ? 'Проверка...' : 'Подтвердить телефон'}
              </button>

              <button
                type="button"
                onClick={() => setView('register')}
                className="w-full py-2.5 bg-transparent text-slate-400 hover:text-slate-200 text-xs font-medium mt-1"
              >
                Вернуться назад
              </button>
            </form>
          )}

          {/* VIEW: SET PIN */}
          {view === 'pin' && (
            <form onSubmit={handleSetPin} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Придумайте новый PIN-код</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <Lock className="w-4 h-4" />
                  </span>
                  <input
                    type="password"
                    id="new-pin"
                    name="newPin"
                    autoComplete="new-password"
                    maxLength="4"
                    placeholder="4 цифры"
                    value={newPin}
                    onChange={(e) => setNewPin(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 tracking-[0.3em] focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Подтвердите PIN-код</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <Lock className="w-4 h-4" />
                  </span>
                  <input
                    type="password"
                    id="confirm-pin"
                    name="confirmPin"
                    autoComplete="new-password"
                    maxLength="4"
                    placeholder="Повторите 4 цифры"
                    value={confirmPin}
                    onChange={(e) => setConfirmPin(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2.5 bg-slate-950 border border-slate-800/80 rounded-xl text-sm text-slate-100 placeholder-slate-600 tracking-[0.3em] focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-emerald-600/20 active:scale-[0.98] transition-all disabled:opacity-50"
              >
                {loading ? 'Сохранение...' : 'Задать PIN-код и продолжить'}
              </button>
            </form>
          )}

        </div>
      </div>
    </div>
  );
};

export default Auth;
