import React, { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { useNavigate } from 'react-router-dom';
import { Plus, Store, User, CreditCard, Calendar, ArrowRight, ShieldCheck, AlertCircle, RefreshCw, Landmark, Receipt } from 'lucide-react';

const OwnerDashboard = () => {
  const navigate = useNavigate();
  
  // Состояние
  const [shops, setShops] = useState([]);
  const [selectedShopId, setSelectedShopId] = useState('');
  const [agreements, setAgreements] = useState([]);
  
  const [loadingShops, setLoadingShops] = useState(true);
  const [loadingAgreements, setLoadingAgreements] = useState(false);
  const [error, setError] = useState('');
  
  // Модальные окна
  const [showAgreementModal, setShowAgreementModal] = useState(false);
  const [showNewOpModal, setShowNewOpModal] = useState(false);

  // Поля формы создания договора
  const [custCid, setCustCid] = useState('');
  const [creditLimit, setCreditLimit] = useState('');
  const [dueDate, setDueDate] = useState('');

  // Поля формы "Новая операция"
  const [newOpCid, setNewOpCid] = useState('');
  const [newOpType, setNewOpType] = useState('purchase'); // 'purchase' | 'repayment'
  const [newOpAmount, setNewOpAmount] = useState('');
  const [newOpReceiptFile, setNewOpReceiptFile] = useState(null);
  
  // Состояния проверки договора в "Новой операции"
  const [newOpResolvedCustomerName, setNewOpResolvedCustomerName] = useState('');
  const [newOpAgreementId, setNewOpAgreementId] = useState('');

  const [formError, setFormError] = useState('');
  const [formSuccess, setFormSuccess] = useState('');
  const [formLoading, setFormLoading] = useState(false);

  useEffect(() => {
    fetchShops();
  }, []);

  useEffect(() => {
    if (selectedShopId) {
      fetchAgreements(selectedShopId);
    } else {
      setAgreements([]);
    }
  }, [selectedShopId]);

  // Эффект для автоматического разрешения договора при вводе 6-значного CID в "Новой операции"
  useEffect(() => {
    if (newOpCid.length === 6 && selectedShopId) {
      resolveActiveAgreement();
    } else {
      setNewOpResolvedCustomerName('');
      setNewOpAgreementId('');
    }
  }, [newOpCid, selectedShopId]);

  const fetchShops = async () => {
    setLoadingShops(true);
    setError('');
    try {
      const shopList = await api.getShops();
      setShops(shopList || []);
      if (shopList && shopList.length > 0) {
        setSelectedShopId(shopList[0].id);
      }
    } catch (err) {
      setError('Не удалось загрузить список магазинов');
    } finally {
      setLoadingShops(false);
    }
  };

  const fetchAgreements = async (shopId) => {
    setLoadingAgreements(true);
    try {
      const allAgreements = await api.getAgreements();
      const filtered = (allAgreements || []).filter(a => a.shop_id === shopId);
      setAgreements(filtered || []);
    } catch (err) {
      setError('Не удалось загрузить договоры магазина');
    } finally {
      setLoadingAgreements(false);
    }
  };

  const resolveActiveAgreement = async () => {
    setFormError('');
    try {
      const activeAg = await api.getActiveAgreementByCID(newOpCid, selectedShopId);
      setNewOpAgreementId(activeAg.id);
      setNewOpResolvedCustomerName(activeAg.customer_name);
    } catch (err) {
      setFormError(err.message || 'Договор не найден');
      setNewOpResolvedCustomerName('');
      setNewOpAgreementId('');
    }
  };

  const handleCreateAgreement = async (e) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    setFormLoading(true);
    try {
      if (!custCid || !creditLimit || !dueDate) {
        throw new Error('Пожалуйста, заполните все поля');
      }
      
      await api.createAgreement(selectedShopId, custCid, creditLimit, dueDate);
      setFormSuccess('Договор создан! Код подтверждения отправлен клиенту по SMS.');
      fetchAgreements(selectedShopId);
      
      setTimeout(() => {
        setShowAgreementModal(false);
        setCustCid('');
        setCreditLimit('');
        setDueDate('');
        setFormSuccess('');
      }, 2000);
    } catch (err) {
      setFormError(err.message);
    } finally {
      setFormLoading(false);
    }
  };

  const handleCreateNewOp = async (e) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    setFormLoading(true);

    try {
      if (!newOpAgreementId) {
        throw new Error('Пожалуйста, укажите корректный 6-значный ID покупателя с активным договором');
      }
      if (!newOpAmount || parseFloat(newOpAmount) <= 0) {
        throw new Error('Сумма операции должна быть больше нуля');
      }

      let receiptUrl = null;
      if (newOpType === 'purchase') {
        if (!newOpReceiptFile) {
          throw new Error('Фото чека обязательно при фиксации покупки в долг');
        }
        const uploadRes = await api.uploadReceipt(newOpReceiptFile);
        receiptUrl = uploadRes.url;
      }

      await api.createTransaction(newOpAgreementId, newOpType, newOpAmount, receiptUrl);
      setFormSuccess('Операция успешно создана! Код подтверждения отправлен покупателю по SMS.');
      
      // Обновляем баланс договоров
      fetchAgreements(selectedShopId);

      setTimeout(() => {
        setShowNewOpModal(false);
        setNewOpCid('');
        setNewOpAmount('');
        setNewOpReceiptFile(null);
        setNewOpResolvedCustomerName('');
        setNewOpAgreementId('');
        setFormSuccess('');
      }, 2000);
    } catch (err) {
      setFormError(err.message);
    } finally {
      setFormLoading(false);
    }
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case 'active':
        return <span className="bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px] px-2 py-0.5 rounded-full font-bold">Активен</span>;
      case 'pending_confirmation':
        return <span className="bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[10px] px-2 py-0.5 rounded-full font-bold">Ожидает подписи</span>;
      case 'closed':
        return <span className="bg-slate-500/10 text-slate-400 border border-slate-500/20 text-[10px] px-2 py-0.5 rounded-full font-bold">Закрыт (Архив)</span>;
      case 'suspended':
        return <span className="bg-rose-500/10 text-rose-400 border border-rose-500/20 text-[10px] px-2 py-0.5 rounded-full font-bold">Заблокирован</span>;
      default:
        return <span className="bg-slate-800 text-slate-300 text-[10px] px-2 py-0.5 rounded-full">{status}</span>;
    }
  };

  return (
    <div className="space-y-6">
      
      {/* Главный заголовок и выбор магазина */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 bg-slate-900/40 backdrop-blur-md border border-slate-800/80 p-6 rounded-3xl">
        <div className="flex items-center space-x-4">
          <div className="w-12 h-12 rounded-2xl bg-indigo-600/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
            <Store className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-100">Панель управления владельца</h1>
            <p className="text-xs text-slate-400">Ведите учет долгов и регистрируйте новые продажи</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {shops.length > 0 ? (
            <select
              value={selectedShopId}
              onChange={(e) => setSelectedShopId(e.target.value)}
              className="bg-slate-950 border border-slate-800 text-slate-200 py-2.5 px-4 rounded-xl text-sm font-medium focus:outline-none focus:border-indigo-500 cursor-pointer min-w-[200px]"
            >
              {shops.map(shop => (
                <option key={shop.id} value={shop.id}>{shop.name}</option>
              ))}
            </select>
          ) : (
            <span className="text-xs text-slate-500 font-medium">Нет доступных магазинов</span>
          )}

          <button
            onClick={() => setShowNewOpModal(true)}
            disabled={shops.length === 0}
            className="flex items-center space-x-1.5 py-2.5 px-4 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-sm rounded-xl transition-all shadow-md active:scale-95 cursor-pointer disabled:opacity-50"
          >
            <Plus className="w-4 h-4" />
            <span>Новая операция</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm rounded-2xl flex items-center gap-2">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Список Договоров по выбранному магазину */}
      {selectedShopId && (
        <div className="bg-slate-900/20 border border-slate-800/80 rounded-3xl overflow-hidden backdrop-blur-sm">
          <div className="p-6 border-b border-slate-800 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-bold text-slate-200">Долговые договоры</h2>
              <p className="text-xs text-slate-400">Список лимитов и текущего баланса покупателей</p>
            </div>
            <button
              onClick={() => setShowAgreementModal(true)}
              className="flex items-center space-x-1.5 py-2 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-xs rounded-xl transition-all shadow-md active:scale-95 cursor-pointer"
            >
              <Plus className="w-4 h-4" />
              <span>Создать договор</span>
            </button>
          </div>

          {loadingAgreements ? (
            <div className="p-12 text-center">
              <div className="w-8 h-8 rounded-full border-2 border-slate-800 border-t-indigo-500 animate-spin mx-auto"></div>
              <p className="mt-4 text-xs text-slate-400">Загрузка договоров...</p>
            </div>
          ) : agreements.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 bg-slate-900/30 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                    <th className="p-4 pl-6">Покупатель</th>
                    <th className="p-4">ID (CID)</th>
                    <th className="p-4">Телефон</th>
                    <th className="p-4">Текущий долг</th>
                    <th className="p-4">Кредитный лимит</th>
                    <th className="p-4">Срок погашения</th>
                    <th className="p-4">Статус</th>
                    <th className="p-4 pr-6 text-right">Действия</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60">
                  {agreements.map((agreement) => (
                    <tr key={agreement.id} className="hover:bg-slate-900/20 transition-colors group">
                      <td className="p-4 pl-6">
                        <div className="font-semibold text-slate-200 text-sm">{agreement.customer_name || 'Неизвестный'}</div>
                      </td>
                      <td className="p-4 text-xs text-slate-300 font-mono tracking-wider">{agreement.customer_cid || agreement.cid || '—'}</td>
                      <td className="p-4 text-xs text-slate-400">{agreement.customer_phone}</td>
                      <td className="p-4 font-bold text-indigo-400 text-sm">
                        {agreement.balance.toLocaleString('ru-RU')} ₸
                      </td>
                      <td className="p-4 text-slate-300 text-sm">
                        {agreement.credit_limit.toLocaleString('ru-RU')} ₸
                      </td>
                      <td className="p-4 text-xs text-slate-400">
                        {new Date(agreement.due_date).toLocaleDateString('ru-RU')}
                      </td>
                      <td className="p-4">{getStatusBadge(agreement.status)}</td>
                      <td className="p-4 pr-6 text-right">
                        <button
                          onClick={() => navigate(`/agreements/${agreement.id}`)}
                          className="inline-flex items-center space-x-1 py-1.5 px-3 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white rounded-lg text-xs font-semibold border border-slate-700/50 transition-colors"
                        >
                          <span>Открыть журнал</span>
                          <ArrowRight className="w-3.5 h-3.5 transition-transform group-hover:translate-x-0.5" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="p-12 text-center">
              <div className="w-12 h-12 rounded-2xl bg-slate-900 flex items-center justify-center text-slate-500 mx-auto mb-4 border border-slate-800/50">
                <AlertCircle className="w-6 h-6" />
              </div>
              <h3 className="text-sm font-semibold text-slate-300">Нет договоров</h3>
              <p className="text-xs text-slate-400 mt-1">В этом магазине пока не оформлено ни одного договора.</p>
            </div>
          )}
        </div>
      )}

      {/* МОДАЛЬНОЕ ОКНО: СОЗДАТЬ ДОГОВОР */}
      {showAgreementModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowAgreementModal(false)}></div>
          <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-md p-6 relative z-10 shadow-2xl">
            <h3 className="text-lg font-bold text-slate-100 mb-2">Создать долговой договор</h3>
            <p className="text-xs text-slate-400 mb-4">Установите лимит долга для покупателя. Требуется подтверждение по SMS.</p>
            
            {formError && <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">{formError}</div>}
            {formSuccess && <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs rounded-xl flex items-center gap-1.5"><ShieldCheck className="w-4 h-4 flex-shrink-0" /> <span>{formSuccess}</span></div>}
            
            <form onSubmit={handleCreateAgreement} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">6-значный ID Покупателя (CID)</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <User className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    maxLength="6"
                    placeholder="Например, 104523"
                    value={custCid}
                    onChange={(e) => setCustCid(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <span className="block mt-1 text-[9px] text-slate-500">* Покупатель должен предоставить свой ID из личного кабинета</span>
              </div>
              
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Лимит долга (₸)</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <CreditCard className="w-4 h-4" />
                  </span>
                  <input
                    type="number"
                    placeholder="Например, 50000"
                    value={creditLimit}
                    onChange={(e) => setCreditLimit(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Крайний срок погашения</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <Calendar className="w-4 h-4" />
                  </span>
                  <input
                    type="date"
                    value={dueDate}
                    onChange={(e) => setDueDate(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
              </div>

              <div className="flex gap-3 justify-end pt-2">
                <button
                  type="button"
                  onClick={() => setShowAgreementModal(false)}
                  className="py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-all cursor-pointer"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={formLoading || formSuccess !== ''}
                  className="py-2.5 px-5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/10 transition-all cursor-pointer disabled:opacity-50"
                >
                  {formLoading ? 'Отправка...' : 'Отправить код'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* МОДАЛЬНОЕ ОКНО: БЫСТРАЯ ОПЕРАЦИЯ (НОВАЯ ОПЕРАЦИЯ ПО CID) */}
      {showNewOpModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowNewOpModal(false)}></div>
          <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-sm p-6 relative z-10 shadow-2xl">
            <h3 className="text-lg font-bold text-slate-100 mb-2">Провести операцию</h3>
            <p className="text-xs text-slate-400 mb-4">Быстрая покупка в долг или погашение по CID покупателя</p>
            
            {formError && <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">{formError}</div>}
            {formSuccess && <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs rounded-xl flex items-center gap-1.5"><ShieldCheck className="w-4 h-4 flex-shrink-0" /> <span>{formSuccess}</span></div>}
            
            <form onSubmit={handleCreateNewOp} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">ID Покупателя (6 знаков)</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <User className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    maxLength="6"
                    placeholder="Например, 102439"
                    value={newOpCid}
                    onChange={(e) => setNewOpCid(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                {newOpResolvedCustomerName && (
                  <span className="block mt-1.5 text-xs text-emerald-400 font-semibold bg-emerald-500/10 p-2 rounded-xl border border-emerald-500/20">
                    Покупатель: {newOpResolvedCustomerName}
                  </span>
                )}
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Тип операции</label>
                <div className="grid grid-cols-2 p-1 bg-slate-950 rounded-xl border border-slate-800/80">
                  <button
                    type="button"
                    onClick={() => setNewOpType('purchase')}
                    className={`py-2 text-xs font-semibold rounded-lg transition-all ${
                      newOpType === 'purchase'
                        ? 'bg-slate-800 text-slate-100 shadow-sm'
                        : 'text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    Покупка в долг
                  </button>
                  <button
                    type="button"
                    onClick={() => setNewOpType('repayment')}
                    className={`py-2 text-xs font-semibold rounded-lg transition-all ${
                      newOpType === 'repayment'
                        ? 'bg-slate-800 text-slate-100 shadow-sm'
                        : 'text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    Погашение
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Сумма (₸)</label>
                <input
                  type="number"
                  placeholder="Введите сумму"
                  value={newOpAmount}
                  onChange={(e) => setNewOpAmount(e.target.value)}
                  className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>

              {newOpType === 'purchase' && (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Фото чека (Обязательно)</label>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => setNewOpReceiptFile(e.target.files[0])}
                    className="block w-full text-xs text-slate-400 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-slate-800 file:text-slate-300 hover:file:bg-slate-700 cursor-pointer"
                  />
                </div>
              )}

              <div className="flex gap-3 justify-end pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowNewOpModal(false);
                    setNewOpCid('');
                    setNewOpAmount('');
                    setNewOpReceiptFile(null);
                    setNewOpResolvedCustomerName('');
                    setNewOpAgreementId('');
                    setFormError('');
                  }}
                  className="py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-all cursor-pointer"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={formLoading || !newOpAgreementId || formSuccess !== ''}
                  className="py-2.5 px-5 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-emerald-600/10 transition-all cursor-pointer disabled:opacity-50"
                >
                  {formLoading ? 'Отправка...' : 'Отправить код'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
};

export default OwnerDashboard;
