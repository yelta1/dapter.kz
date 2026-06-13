import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Auth from './pages/Auth/Auth';
import Dashboard from './pages/Dashboard/Dashboard';
import AdminDashboard from './pages/Dashboard/AdminDashboard';
import AgreementDetail from './pages/Agreements/AgreementDetail';
import ProtectedRoute from './components/Layout/ProtectedRoute';
import { useAuth } from './context/AuthContext';

function App() {
  const { user } = useAuth();

  return (
    <Routes>
      {/* Публичные пути */}
      <Route path="/auth" element={<Auth />} />
      
      {/* Принудительная установка PIN после OTP */}
      <Route 
        path="/auth/set-pin" 
        element={
          <ProtectedRoute>
            <Auth />
          </ProtectedRoute>
        } 
      />

      {/* Защищенные разделы */}
      <Route 
        path="/admin" 
        element={
          <ProtectedRoute allowedRoles={['admin']}>
            <AdminDashboard />
          </ProtectedRoute>
        } 
      />
      <Route 
        path="/owner" 
        element={
          <ProtectedRoute allowedRoles={['owner']}>
            <Dashboard />
          </ProtectedRoute>
        } 
      />
      <Route 
        path="/customer" 
        element={
          <ProtectedRoute allowedRoles={['customer']}>
            <Dashboard />
          </ProtectedRoute>
        } 
      />
      <Route 
        path="/agreements/:id" 
        element={
          <ProtectedRoute>
            <AgreementDetail />
          </ProtectedRoute>
        } 
      />

      {/* Дефолтный редирект */}
      <Route 
        path="*" 
        element={
          user 
            ? <Navigate to={user.role === 'admin' ? '/admin' : (user.role === 'owner' ? '/owner' : '/customer')} replace /> 
            : <Navigate to="/auth" replace />
        } 
      />
    </Routes>
  );
}

export default App;
