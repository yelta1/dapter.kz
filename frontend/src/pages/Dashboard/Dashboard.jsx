import React from 'react';
import { useAuth } from '../../context/AuthContext';
import OwnerDashboard from './OwnerDashboard';
import CustomerDashboard from './CustomerDashboard';
import Layout from '../../components/Layout/Layout';

const Dashboard = () => {
  const { user } = useAuth();

  return (
    <Layout>
      {user?.role === 'owner' ? <OwnerDashboard /> : <CustomerDashboard />}
    </Layout>
  );
};

export default Dashboard;
