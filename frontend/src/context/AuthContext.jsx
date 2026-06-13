import React, { createContext, useState, useEffect, useContext } from 'react';
import { api } from '../services/api';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [needsPin, setNeedsPin] = useState(false);

  // Функция для запроса профиля пользователя при загрузке токена
  const fetchProfile = async () => {
    try {
      const profile = await api.getMe();
      setUser(profile);
      // Если у покупателя отсутствует PIN-код (например, сессия осталась, но пин не настроен)
      if (profile.role === 'customer' && !profile.has_pin) {
        // Заглушка: бэкенд возвращает пользователя. Если он сохранен в БД без PIN-кода,
        // то мы помечаем это. В нашей Go структуре, если pin_code_hash == nil, бэкенд не пустит на логин,
        // но при регистрации выдает токен до установки PIN.
      }
    } catch (err) {
      console.error('Ошибка получения профиля:', err);
      logout();
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      fetchProfile();
    } else {
      setLoading(false);
    }
  }, []);

  // Вход
  const login = async (phone, password) => {
    setLoading(true);
    try {
      const data = await api.login(phone, password);
      localStorage.setItem('token', data.token);
      await fetchProfile();
      return { success: true };
    } catch (err) {
      setLoading(false);
      throw err;
    }
  };

  // Выход
  const logout = () => {
    localStorage.removeItem('token');
    setUser(null);
    setNeedsPin(false);
  };

  // Регистрация владельца
  const registerOwner = async (phone, password, iin, fullName) => {
    setLoading(true);
    try {
      const data = await api.registerOwner(phone, password, iin, fullName);
      // После успешной регистрации владельца сразу логиним его
      return login(phone, password);
    } catch (err) {
      setLoading(false);
      throw err;
    }
  };

  // Инициация регистрации покупателя (отправка OTP)
  const registerCustomerInitiate = async (phone, iin, fullName) => {
    return api.registerCustomer(phone, iin, fullName);
  };

  // Завершение регистрации покупателя (код из SMS)
  const registerCustomerVerify = async (phone, code) => {
    setLoading(true);
    try {
      const data = await api.verifyCustomerRegister(phone, code);
      localStorage.setItem('token', data.token);
      setNeedsPin(data.needs_pin || false);
      
      // Запрашиваем мелкими шагами профиль
      const profile = await api.getMe();
      setUser(profile);
      return { success: true };
    } catch (err) {
      setLoading(false);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Задать PIN-код
  const setCustomerPin = async (pin) => {
    setLoading(true);
    try {
      await api.setPin(pin);
      setNeedsPin(false);
      // Обновляем профиль, так как PIN установлен
      await fetchProfile();
    } catch (err) {
      setLoading(false);
      throw err;
    }
  };

  const value = {
    user,
    loading,
    needsPin,
    login,
    logout,
    registerOwner,
    registerCustomerInitiate,
    registerCustomerVerify,
    setCustomerPin,
    fetchProfile,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth должен использоваться внутри AuthProvider');
  }
  return context;
};
