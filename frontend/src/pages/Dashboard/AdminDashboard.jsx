import React, { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { useAuth } from '../../context/AuthContext';
import { Plus, Users, Store, UserCheck, ShieldCheck, LogOut, AlertCircle, RefreshCw } from 'lucide-react';
import Layout from '../../components/Layout/Layout';

const AdminDashboard = () => {
  const { logout } = useAuth();
  
  // Состояния
  const [activeTab, setActiveTab] = useState('owners'); // 'owners' | 'shops' | 'customers'
  const [owners, setOwners] = useState([]);
  const [shops, setShops] = useState([]);
  const [customers, setCustomers] = useState([]);
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Состояния модалок
  const [showOwnerModal, setShowOwnerModal] = useState(false);
  const [showShopModal, setShowShopModal] = useState(false);
  
  // Поля формы владельца
  const [ownerPhone, setOwnerPhone] = useState('');
  const [ownerPassword, setOwnerPassword] = useState('');
  const [ownerIin, setOwnerIin] = useState('');
  const [ownerName, setOwnerName] = useState('');
  
  // Поля формы магазина
  const [shopName, setShopName] = useState('');
  const [shopAddress, setShopAddress] = useState('');
  const [selectedOwnerId, setSelectedOwnerId] = useState('');

  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');
  const [formSuccess, setFormSuccess] = useState('');

  useEffect(() => {
    fetchData();
  }, [activeTab]);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      if (activeTab === 'owners') {
        const list = await api.getOwners();
        setOwners(list || []);
      } else if (activeTab === 'shops') {
        const list = await api.getAllShops();
        setShops(list || []);
        
        // Дополнительно подгружаем владельцев, чтобы выбрать при создании магазина
        const ownerList = await api.getOwners();
        setOwners(ownerList || []);
        if (ownerList && ownerList.length > 0) {
          setSelectedOwnerId(ownerList[0].id);
        }
      } else if (activeTab === 'customers') {
        const list = await api.getCustomers();
        setCustomers(list || []);
      }
    } catch (err) {
      setError(err.message || 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterOwner = async (e) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    setFormLoading(true);

    try {
      if (!ownerPhone || !ownerPassword || !ownerIin || !ownerName) {
        throw new Error('Заполните все обязательные поля');
      }
      if (ownerIin.length !== 12) {
        throw new Error('ИИН должен состоять из 12 цифр');
      }

      await api.registerOwner(ownerPhone, ownerPassword, ownerIin, ownerName);
      setFormSuccess('Владелец успешно зарегистрирован!');
      fetchData();
      
      setTimeout(() => {
        setShowOwnerModal(false);
        setOwnerPhone('');
        setOwnerPassword('');
        setOwnerIin('');
        setOwnerName('');
        setFormSuccess('');
      }, 1500);
    } catch (err) {
      setFormError(err.message);
    } finally {
      setFormLoading(false);
    }
  };

  const handleCreateShop = async (e) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    setFormLoading(true);

    try {
      if (!shopName || !selectedOwnerId) {
        throw new Error('Название магазина и владелец обязательны');
      }

      await api.createShop(selectedOwnerId, shopName, shopAddress);
      setFormSuccess('Магазин успешно создан!');
      fetchData();

      setTimeout(() => {
        setShowShopModal(false);
        setShopName('');
        setShopAddress('');
        setFormSuccess('');
      }, 1500);
    } catch (err) {
      setFormError(err.message);
    } finally {
      setFormLoading(false);
    }
  };

  // Метод для сопоставления имени владельца в списке магазинов
  const getOwnerName = (ownerId) => {
    const found = owners.find(o => o.id === ownerId);
    return found ? `${found.full_name} (${found.phone})` : ownerId;
  };

  return (
    <Layout>
      <div className="space-y-6">
        
        {/* Шапка админ-панели */}
        <div className="flex justify-between items-center bg-slate-900/40 backdrop-blur-md border border-slate-800/80 p-6 rounded-3xl">
          <div>
            <h1 className="text-xl font-bold text-slate-100 flex items-center gap-2">
              <ShieldCheck className="w-6 h-6 text-indigo-500" />
              <span>Панель Администратора</span>
            </h1>
            <p className="text-xs text-slate-400 mt-1">Регистрация участников системы, создание магазинов и мониторинг пользователей</p>
          </div>
          <button
            onClick={logout}
            className="flex items-center gap-1.5 py-2 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold text-xs rounded-xl cursor-pointer transition-colors border border-slate-700/50"
          >
            <LogOut className="w-4 h-4" />
            <span>Выйти</span>
          </button>
        </div>

        {error && (
          <div className="p-4 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm rounded-2xl flex items-center gap-2">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Навигационные вкладки и кнопки действий */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex p-1 bg-slate-950 rounded-xl border border-slate-800/80 max-w-fit">
            <button
              onClick={() => setActiveTab('owners')}
              className={`flex items-center gap-2 py-2 px-4 text-xs font-semibold rounded-lg transition-all ${
                activeTab === 'owners' 
                  ? 'bg-slate-800 text-slate-100 shadow-sm' 
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <UserCheck className="w-4 h-4" />
              <span>Владельцы</span>
            </button>
            <button
              onClick={() => setActiveTab('shops')}
              className={`flex items-center gap-2 py-2 px-4 text-xs font-semibold rounded-lg transition-all ${
                activeTab === 'shops' 
                  ? 'bg-slate-800 text-slate-100 shadow-sm' 
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Store className="w-4 h-4" />
              <span>Магазины</span>
            </button>
            <button
              onClick={() => setActiveTab('customers')}
              className={`flex items-center gap-2 py-2 px-4 text-xs font-semibold rounded-lg transition-all ${
                activeTab === 'customers' 
                  ? 'bg-slate-800 text-slate-100 shadow-sm' 
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Users className="w-4 h-4" />
              <span>Покупатели</span>
            </button>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={fetchData}
              disabled={loading}
              className="p-2.5 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-slate-200 rounded-xl border border-slate-700/50 cursor-pointer disabled:opacity-50"
              title="Обновить"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>

            {activeTab === 'owners' && (
              <button
                onClick={() => setShowOwnerModal(true)}
                className="flex items-center gap-1.5 py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-xs rounded-xl shadow-lg transition-all active:scale-95 cursor-pointer"
              >
                <Plus className="w-4 h-4" />
                <span>Создать владельца</span>
              </button>
            )}

            {activeTab === 'shops' && (
              <button
                onClick={() => setShowShopModal(true)}
                className="flex items-center gap-1.5 py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-xs rounded-xl shadow-lg transition-all active:scale-95 cursor-pointer"
              >
                <Plus className="w-4 h-4" />
                <span>Создать магазин</span>
              </button>
            )}
          </div>
        </div>

        {/* Списки данных */}
        <div className="bg-slate-900/20 border border-slate-800/80 rounded-3xl overflow-hidden backdrop-blur-sm">
          {loading ? (
            <div className="p-16 text-center">
              <div className="w-10 h-10 rounded-full border-2 border-slate-850 border-t-indigo-500 animate-spin mx-auto"></div>
              <p className="mt-4 text-xs text-slate-400">Загрузка данных...</p>
            </div>
          ) : activeTab === 'owners' ? (
            owners.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-slate-800 bg-slate-900/30 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                      <th className="p-4 pl-6">ФИО владельца</th>
                      <th className="p-4">Телефон</th>
                      <th className="p-4">ИИН</th>
                      <th className="p-4 pr-6 text-right">Зарегистрирован</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60 text-sm">
                    {owners.map(owner => (
                      <tr key={owner.id} className="hover:bg-slate-900/20 transition-colors">
                        <td className="p-4 pl-6 font-semibold text-slate-200">{owner.full_name}</td>
                        <td className="p-4 text-slate-300">{owner.phone}</td>
                        <td className="p-4 text-slate-400 font-mono">{owner.iin}</td>
                        <td className="p-4 pr-6 text-right text-xs text-slate-500">
                          {new Date(owner.created_at).toLocaleDateString('ru-RU')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="p-12 text-center text-slate-500 text-xs">Нет зарегистрированных владельцев в системе.</div>
            )
          ) : activeTab === 'shops' ? (
            shops.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-slate-800 bg-slate-900/30 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                      <th className="p-4 pl-6">Название магазина</th>
                      <th className="p-4">Адрес / Локация</th>
                      <th className="p-4 pr-6">Владелец</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60 text-sm">
                    {shops.map(shop => (
                      <tr key={shop.id} className="hover:bg-slate-900/20 transition-colors">
                        <td className="p-4 pl-6 font-semibold text-slate-200">{shop.name}</td>
                        <td className="p-4 text-slate-300">{shop.address || '—'}</td>
                        <td className="p-4 pr-6 text-slate-400 text-xs">{getOwnerName(shop.owner_id)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="p-12 text-center text-slate-500 text-xs">Нет созданных магазинов в системе.</div>
            )
          ) : (
            customers.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-slate-800 bg-slate-900/30 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                      <th className="p-4 pl-6">ID Покупателя (CID)</th>
                      <th className="p-4">ФИО</th>
                      <th className="p-4">Телефон</th>
                      <th className="p-4 pr-6">ИИН</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60 text-sm">
                    {customers.map(c => (
                      <tr key={c.id} className="hover:bg-slate-900/20 transition-colors">
                        <td className="p-4 pl-6 font-bold text-indigo-400 font-mono tracking-wider">{c.cid || '—'}</td>
                        <td className="p-4 font-semibold text-slate-200">{c.full_name}</td>
                        <td className="p-4 text-slate-300">{c.phone}</td>
                        <td className="p-4 pr-6 text-slate-400 font-mono">{c.iin}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="p-12 text-center text-slate-500 text-xs">Нет зарегистрированных покупателей в системе.</div>
            )
          )}
        </div>

        {/* МОДАЛЬНОЕ ОКНО: СОЗДАТЬ ВЛАДЕЛЬЦА */}
        {showOwnerModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowOwnerModal(false)}></div>
            <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-sm p-6 relative z-10 shadow-2xl">
              <h3 className="text-lg font-bold text-slate-100 mb-2">Создать аккаунт Владельца</h3>
              <p className="text-xs text-slate-400 mb-4">Зарегистрируйте нового владельца для ведения кредитного журнала</p>
              
              {formError && <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">{formError}</div>}
              {formSuccess && <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs rounded-xl">{formSuccess}</div>}
              
              <form onSubmit={handleRegisterOwner} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">ФИО полностью</label>
                  <input
                    type="text"
                    placeholder="Например, Ахметов Марат"
                    value={ownerName}
                    onChange={(e) => setOwnerName(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">ИИН (12 цифр)</label>
                  <input
                    type="text"
                    maxLength="12"
                    placeholder="Введите 12 цифр ИИН"
                    value={ownerIin}
                    onChange={(e) => setOwnerIin(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Номер телефона</label>
                  <input
                    type="tel"
                    placeholder="+77071234567"
                    value={ownerPhone}
                    onChange={(e) => setOwnerPhone(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Пароль владельца</label>
                  <input
                    type="password"
                    placeholder="Минимум 6 символов"
                    value={ownerPassword}
                    onChange={(e) => setOwnerPassword(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div className="flex gap-3 justify-end pt-2">
                  <button
                    type="button"
                    onClick={() => setShowOwnerModal(false)}
                    className="py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-all cursor-pointer"
                  >
                    Отмена
                  </button>
                  <button
                    type="submit"
                    disabled={formLoading || formSuccess !== ''}
                    className="py-2.5 px-5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/10 transition-all cursor-pointer disabled:opacity-50"
                  >
                    {formLoading ? 'Создание...' : 'Создать'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* МОДАЛЬНОЕ ОКНО: СОЗДАТЬ МАГАЗИН */}
        {showShopModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowShopModal(false)}></div>
            <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-sm p-6 relative z-10 shadow-2xl">
              <h3 className="text-lg font-bold text-slate-100 mb-2">Создать магазин</h3>
              <p className="text-xs text-slate-400 mb-4">Создайте торговую точку и привяжите ее к владельцу</p>
              
              {formError && <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">{formError}</div>}
              {formSuccess && <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs rounded-xl">{formSuccess}</div>}
              
              <form onSubmit={handleCreateShop} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Название магазина</label>
                  <input
                    type="text"
                    placeholder="Магазин 'Береке'"
                    value={shopName}
                    onChange={(e) => setShopName(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Адрес магазина</label>
                  <input
                    type="text"
                    placeholder="ул. Абая 12"
                    value={shopAddress}
                    onChange={(e) => setShopAddress(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Выберите владельца</label>
                  {owners.length > 0 ? (
                    <select
                      value={selectedOwnerId}
                      onChange={(e) => setSelectedOwnerId(e.target.value)}
                      className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
                    >
                      {owners.map(owner => (
                        <option key={owner.id} value={owner.id}>{owner.full_name} ({owner.phone})</option>
                      ))}
                    </select>
                  ) : (
                    <span className="block text-xs text-rose-400 font-semibold mt-1">Сначала создайте хотя бы одного владельца!</span>
                  )}
                </div>
                <div className="flex gap-3 justify-end pt-2">
                  <button
                    type="button"
                    onClick={() => setShowShopModal(false)}
                    className="py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-all cursor-pointer"
                  >
                    Отмена
                  </button>
                  <button
                    type="submit"
                    disabled={formLoading || owners.length === 0 || formSuccess !== ''}
                    className="py-2.5 px-5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/10 transition-all cursor-pointer disabled:opacity-50"
                  >
                    {formLoading ? 'Создание...' : 'Создать'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

      </div>
    </Layout>
  );
};

export default AdminDashboard;
