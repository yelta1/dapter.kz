import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../../services/api';
import { useAuth } from '../../context/AuthContext';
import { ArrowLeft, Plus, Landmark, Receipt, AlertCircle, Calendar, ShieldCheck, CheckCircle2, XCircle, Clock } from 'lucide-react';
import Layout from '../../components/Layout/Layout';

const AgreementDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();

  // Состояния данных
  const [agreement, setAgreement] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Состояния модалок создания транзакций
  const [showTxModal, setShowTxModal] = useState(false);
  const [txType, setTxType] = useState('purchase'); // 'purchase' или 'repayment'
  const [amount, setAmount] = useState('');
  const [receiptFile, setReceiptFile] = useState(null);

  // Состояния отправки форм
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');
  const [formSuccess, setFormSuccess] = useState('');

  useEffect(() => {
    fetchData();
  }, [id]);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      // Запрашиваем детали договора
      const details = await api.getAgreementById(id);
      setAgreement(details);

      // Запрашиваем транзакции по договору
      const txHistory = await api.getTransactions(id);
      setTransactions(txHistory || []);
    } catch (err) {
      setError(err.message || 'Ошибка загрузки деталей договора');
    } finally {
      setLoading(false);
    }
  };

  const openCreateTxModal = (type) => {
    setTxType(type);
    setAmount('');
    setReceiptFile(null);
    setFormError('');
    setFormSuccess('');
    setShowTxModal(true);
  };

  const handleCreateTransaction = async (e) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    setFormLoading(true);

    try {
      if (!amount || parseFloat(amount) <= 0) {
        throw new Error('Укажите корректную сумму операции (> 0)');
      }

      let receiptUrl = null;

      if (txType === 'purchase') {
        if (!receiptFile) {
          throw new Error('Загрузка фото чека является обязательной для фиксации покупки в долг');
        }
        
        // 1. Сначала загружаем файл чека на сервер
        const uploadRes = await api.uploadReceipt(receiptFile);
        receiptUrl = uploadRes.url;
      }

      // 2. Создаем транзакцию с полученным URL чека
      await api.createTransaction(id, txType, amount, receiptUrl);
      
      setFormSuccess('Операция успешно создана! Код подтверждения отправлен покупателю по SMS.');
      fetchData();

      setTimeout(() => {
        setShowTxModal(false);
        setAmount('');
        setReceiptFile(null);
        setFormSuccess('');
      }, 2000);
    } catch (err) {
      setFormError(err.message);
    } finally {
      setFormLoading(false);
    }
  };

  if (loading) {
    return (
      <Layout>
        <div className="p-12 text-center">
          <div className="w-10 h-10 rounded-full border-2 border-slate-800 border-t-indigo-500 animate-spin mx-auto"></div>
          <p className="mt-4 text-xs text-slate-400">Загрузка долговой книги...</p>
        </div>
      </Layout>
    );
  }

  if (error || !agreement) {
    return (
      <Layout>
        <div className="max-w-md mx-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl text-center">
          <AlertCircle className="w-12 h-12 text-rose-500 mx-auto mb-4" />
          <h3 className="font-bold text-slate-200">Ошибка операции</h3>
          <p className="text-xs text-slate-400 mt-1.5">{error || 'Договор не найден'}</p>
          <button 
            onClick={() => navigate(user?.role === 'owner' ? '/owner' : '/customer')}
            className="mt-6 py-2 px-4 bg-slate-800 hover:bg-slate-700 text-xs font-semibold rounded-xl"
          >
            Вернуться назад
          </button>
        </div>
      </Layout>
    );
  }

  // Расчет прогресса лимита долга
  const limitPercent = Math.min((agreement.balance / agreement.credit_limit) * 100, 100);

  const getTxStatusLabel = (status) => {
    switch (status) {
      case 'completed':
        return <span className="inline-flex items-center gap-1 text-emerald-400 text-xs font-semibold"><CheckCircle2 className="w-3.5 h-3.5" /> Подтверждено</span>;
      case 'pending':
        return <span className="inline-flex items-center gap-1 text-amber-400 text-xs font-semibold"><Clock className="w-3.5 h-3.5 animate-pulse" /> Ожидает SMS</span>;
      case 'rejected':
        return <span className="inline-flex items-center gap-1 text-rose-400 text-xs font-semibold"><XCircle className="w-3.5 h-3.5" /> Отклонено покупателем</span>;
      case 'expired':
        return <span className="inline-flex items-center gap-1 text-slate-500 text-xs font-semibold"><XCircle className="w-3.5 h-3.5" /> Время вышло</span>;
      default:
        return <span>{status}</span>;
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        
        {/* Кнопка назад */}
        <button
          onClick={() => navigate(user?.role === 'owner' ? '/owner' : '/customer')}
          className="flex items-center space-x-1 text-slate-400 hover:text-slate-200 text-xs font-semibold transition-colors cursor-pointer"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад в кабинет</span>
        </button>

        {/* Информационная карточка Договора */}
        <div className="bg-slate-900/40 backdrop-blur-md border border-slate-800/80 p-6 rounded-3xl relative overflow-hidden">
          {/* Background glow */}
          <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/5 blur-2xl rounded-full"></div>

          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 pb-6 border-b border-slate-800">
            <div>
              <span className={`text-[9px] px-2 py-0.5 rounded font-bold uppercase tracking-wider ${
                agreement.status === 'active' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-amber-500/10 text-amber-400'
              }`}>
                Договор: {agreement.status === 'active' ? 'Активен' : 'На оформлении'}
              </span>
              <h1 className="text-xl font-bold text-slate-100 mt-2">
                Журнал долга: {agreement.customer_name || 'Неизвестный'}
              </h1>
              <p className="text-xs text-slate-400 mt-0.5">Магазин: «{agreement.shop_name}» • Телефон покупателя: {agreement.customer_phone}</p>
            </div>

            {/* Блок действий для Владельца магазина */}
            {user?.role === 'owner' && agreement.status === 'active' && (
              <div className="flex items-center gap-3">
                <button
                  onClick={() => openCreateTxModal('repayment')}
                  className="flex items-center space-x-1 py-2.5 px-4 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs rounded-xl shadow-lg transition-all active:scale-95 cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Принять платеж</span>
                </button>
                <button
                  onClick={() => openCreateTxModal('purchase')}
                  className="flex items-center space-x-1 py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-xs rounded-xl shadow-lg transition-all active:scale-95 cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Покупка в долг</span>
                </button>
              </div>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-6">
            <div>
              <span className="block text-slate-500 text-xs font-semibold uppercase tracking-wider">Текущий долг</span>
              <span className="block text-2xl font-bold text-indigo-400 mt-1">{agreement.balance.toLocaleString()} ₸</span>
            </div>
            
            <div>
              <span className="block text-slate-500 text-xs font-semibold uppercase tracking-wider">Кредитный лимит</span>
              <span className="block text-2xl font-bold text-slate-200 mt-1">{agreement.credit_limit.toLocaleString()} ₸</span>
            </div>

            <div>
              <span className="block text-slate-500 text-xs font-semibold uppercase tracking-wider flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5" /> Срок погашения
              </span>
              <span className="block text-sm font-semibold text-slate-300 mt-2">
                до {new Date(agreement.due_date).toLocaleDateString('ru-RU')}
              </span>
            </div>
          </div>

          {/* Шкала лимита */}
          {agreement.status === 'active' && (
            <div className="mt-6">
              <div className="flex justify-between text-[10px] text-slate-500 font-semibold mb-1">
                <span>ИСПОЛЬЗОВАНО: {limitPercent.toFixed(0)}%</span>
                <span>ДОСТУПНО: {(agreement.credit_limit - agreement.balance).toLocaleString()} ₸</span>
              </div>
              <div className="w-full h-2 bg-slate-950 rounded-full overflow-hidden border border-slate-800">
                <div 
                  className={`h-full rounded-full transition-all duration-300 ${
                    limitPercent > 85 ? 'bg-rose-500' : limitPercent > 50 ? 'bg-amber-500' : 'bg-indigo-500'
                  }`}
                  style={{ width: `${limitPercent}%` }}
                ></div>
              </div>
            </div>
          )}
        </div>

        {/* История Транзакций */}
        <div className="bg-slate-900/20 border border-slate-800/80 rounded-3xl overflow-hidden backdrop-blur-sm">
          <div className="p-6 border-b border-slate-800">
            <h2 className="text-lg font-bold text-slate-200">История операций по договору</h2>
            <p className="text-xs text-slate-400">Список всех зафиксированных покупок и платежей</p>
          </div>

          {transactions.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 bg-slate-900/30 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                    <th className="p-4 pl-6">Дата операции</th>
                    <th className="p-4">Тип</th>
                    <th className="p-4">Сумма</th>
                    <th className="p-4">Документ/Чек</th>
                    <th className="p-4">Статус согласия</th>
                    <th className="p-4 pr-6 text-right">ЭЦП (SMS ID)</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60">
                  {transactions.map((tx) => (
                    <tr key={tx.id} className="hover:bg-slate-900/10 transition-colors">
                      <td className="p-4 pl-6 text-xs text-slate-400">
                        {new Date(tx.created_at).toLocaleString('ru-RU')}
                      </td>
                      <td className="p-4">
                        <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded ${
                          tx.type === 'purchase' ? 'bg-rose-500/10 text-rose-400' : 'bg-emerald-500/10 text-emerald-400'
                        }`}>
                          {tx.type === 'purchase' ? 'Покупка' : 'Платеж'}
                        </span>
                      </td>
                      <td className={`p-4 font-bold text-sm ${
                        tx.type === 'purchase' ? 'text-slate-200' : 'text-emerald-400'
                      }`}>
                        {tx.type === 'purchase' ? '+' : '-'}{tx.amount.toLocaleString()} ₸
                      </td>
                      <td className="p-4 text-xs">
                        {tx.receipt_image_url ? (
                          <a 
                            href={`${api.BASE_URL}${tx.receipt_image_url}`} 
                            target="_blank" 
                            rel="noreferrer" 
                            className="text-indigo-400 hover:text-indigo-300 hover:underline flex items-center gap-1"
                          >
                            <Receipt className="w-3.5 h-3.5" />
                            <span>Смотреть чек</span>
                          </a>
                        ) : (
                          <span className="text-slate-600">—</span>
                        )}
                      </td>
                      <td className="p-4">{getTxStatusLabel(tx.status)}</td>
                      <td className="p-4 pr-6 text-right text-[10px] text-slate-500 font-mono">
                        {tx.signature_sms_id ? tx.signature_sms_id.substring(0, 8) + '...' : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="p-12 text-center text-slate-500">
              <Landmark className="w-12 h-12 mx-auto mb-3 text-slate-700" />
              <p className="text-xs">Операций по данному договору пока не проводилось.</p>
            </div>
          )}
        </div>

        {/* МОДАЛЬНОЕ ОКНО: ЗАФИКСИРОВАТЬ ТРАНЗАКЦИЮ */}
        {showTxModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowTxModal(false)}></div>
            <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-sm p-6 relative z-10 shadow-2xl">
              <h3 className="text-lg font-bold text-slate-100 mb-2">
                {txType === 'purchase' ? 'Фиксация покупки в долг' : 'Прием платежа (погашение)'}
              </h3>
              <p className="text-xs text-slate-400 mb-4">
                Заполните детали операции. Проведение транзакции потребует отправки SMS-кода покупателю.
              </p>

              {formError && (
                <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">
                  {formError}
                </div>
              )}
              {formSuccess && (
                <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs rounded-xl flex items-center gap-1.5">
                  <ShieldCheck className="w-4 h-4 flex-shrink-0" />
                  <span>{formSuccess}</span>
                </div>
              )}

              <form onSubmit={handleCreateTransaction} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1">Сумма операции (₸)</label>
                  <input
                    type="number"
                    placeholder="Например, 5000"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    className="block w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>

                {txType === 'purchase' && (
                  <div>
                    <label className="block text-xs font-medium text-slate-400 mb-1">Фото чека (Обязательно)</label>
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) => setReceiptFile(e.target.files[0])}
                      className="block w-full text-xs text-slate-400 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-slate-800 file:text-slate-300 hover:file:bg-slate-700 cursor-pointer"
                    />
                  </div>
                )}

                <div className="flex gap-3 justify-end pt-2">
                  <button
                    type="button"
                    onClick={() => setShowTxModal(false)}
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

      </div>
    </Layout>
  );
};

export default AgreementDetail;
