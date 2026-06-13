import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

// ProtectedRoute защищает маршруты от неавторизованных пользователей и ролевого доступа
const ProtectedRoute = ({ children, allowedRoles }) => {
  const { user, loading, needsPin } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0b0f19] flex items-center justify-center">
        <div className="relative flex flex-col items-center">
          <div className="w-12 h-12 rounded-full border-4 border-slate-800 border-t-indigo-500 animate-spin"></div>
          <span className="mt-4 text-slate-400 font-medium text-sm">Загрузка сессии...</span>
        </div>
      </div>
    );
  }

  if (!user) {
    // Пользователь не авторизован, перенаправляем на вход
    return <Navigate to="/auth" replace />;
  }

  if (needsPin && window.location.pathname !== '/auth/set-pin') {
    // Покупатель только что зарегистрировался, но еще не ввел PIN
    return <Navigate to="/auth/set-pin" replace />;
  }

  if (allowedRoles && !allowedRoles.includes(user.role)) {
    // У пользователя нет прав доступа к этой странице (например, покупатель ломится к владельцу)
    const fallbackPath = user.role === 'admin' 
      ? '/admin' 
      : user.role === 'owner' 
        ? '/owner' 
        : '/customer';
    return <Navigate to={fallbackPath} replace />;
  }

  return children;
};

export default ProtectedRoute;
