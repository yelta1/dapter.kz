import React, { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { ShieldCheck, ShieldAlert, Check, X, KeyRound, AlertCircle, FileSignature, Landmark } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

const CustomerDashboard = () => {
  const navigate = useNavigate();
  const { user } = useAuth();

  // Состояние
  const [agreements, setAgreements] = useState([]);
  const [pendingTransactions, setPendingTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Состояние OTP-модалки
  const [showOtpModal, setShowOtpModal] = useState(false);
  const [otpCode, setOtpCode] = useState('');
  const [otpLoading, setOtpLoading] = useState(false);
  const [otpError, setOtpError] = useState('');
  
  // Какую сущность мы сейчас подписываем/подтверждаем
  const [confirmTarget, setConfirmTarget] = useState(null); // agreement или transaction
  const [confirmType, setConfirmType] = useState(''); // 'agreement' или 'transaction'

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      // 1. Получаем все договоры покупателя
      const agreementList = await api.getAgreements();
      setAgreements(agreementList || []);

      // 2. Для каждого активного договора получаем транзакции и фильтруем ожидающие
      const activeAgreements = (agreementList || []).filter(a => a.status === 'active');
      let pendingTxs = [];
      
      for (const agreement of activeAgreements) {
        const txList = await api.getTransactions(agreement.id);
        const pending = (txList || []).filter(t => t.status === 'pending');
        // Обогащаем транзакции названием магазина для отображения в алерте
        const enriched = pending.map(t => ({
          ...t,
          shop_name: agreement.shop_name,
        }));
        pendingTxs = [...pendingTxs, ...enriched];
      }

      setPendingTransactions(pendingTxs);
    } catch (err) {
      setError('Ошибка при загрузке данных кабинета покупателя');
    } finally {
      setLoading(false);
    }
  };

  // Инициировать подтверждение договора
  const startConfirmAgreement = (agreement) => {
    setConfirmTarget(agreement);
    setConfirmType('agreement');
    setOtpCode('');
    setOtpError('');
    setShowOtpModal(true);
  };

  // Инициировать подтверждение транзакции (покупки/погашения)
  const startConfirmTransaction = (transaction) => {
    setConfirmTarget(transaction);
    setConfirmType('transaction');
    setOtpCode('');
    setOtpError('');
    setShowOtpModal(true);
  };

  // Отклонить транзакцию
  const handleRejectTransaction = async (transactionId) => {
    if (!window.confirm('Вы действительно хотите отклонить эту покупку/погашение?')) return;
    try {
      await api.rejectTransaction(transactionId);
      fetchData();
    } catch (err) {
      alert(`Не удалось отклонить: ${err.message}`);
    }
  };

  // Отправка OTP кода
  const handleOtpSubmit = async (e) => {
    e.preventDefault();
    setOtpError('');
    setOtpLoading(true);
    try {
      if (!otpCode) throw new Error('Пожалуйста, введите код');
      
      if (confirmType === 'agreement') {
        await api.confirmAgreement(confirmTarget.id, otpCode);
      } else {
        await api.confirmTransaction(confirmTarget.id, otpCode);
      }
      
      setShowOtpModal(false);
      fetchData();
    } catch (err) {
      setOtpError(err.message);
    } finally {
      setOtpLoading(false);
    }
  };

  // Подсчет сводной информации
  const activeAgreements = (agreements || []).filter(a => a.status === 'active');
  const totalLimit = activeAgreements.reduce((sum, a) => sum + a.credit_limit, 0);
  const totalDebt = activeAgreements.reduce((sum, a) => sum + a.balance, 0);
  const totalAvailable = totalLimit - totalDebt;

  const pendingAgreements = (agreements || []).filter(a => a.status === 'pending_confirmation');

  if (loading) {
    return (
      <div className="p-12 text-center">
        <div className="w-10 h-10 rounded-full border-2 border-slate-800 border-t-indigo-500 animate-spin mx-auto"></div>
        <p className="mt-4 text-xs text-slate-400">Загрузка информации кабинета...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      
      {/* Шапка кабинета покупателя */}
      <div className="bg-slate-900/40 backdrop-blur-md border border-slate-800/80 p-6 rounded-3xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Кабинет Покупателя</h1>
          <p className="text-xs text-slate-400">Просматривайте лимиты, контролируйте долги и подписывайте покупки</p>
        </div>
        {user?.cid && (
          <div className="flex flex-col items-start sm:items-end">
            <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-1">Ваш ID Покупателя</span>
            <span className="text-sm font-extrabold text-indigo-400 bg-indigo-500/10 px-3 py-1 rounded-xl border border-indigo-500/20 font-mono tracking-wider">{user.cid}</span>
          </div>
        )}
      </div>

      {error && (
        <div className="p-4 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm rounded-2xl flex items-center gap-2">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Сводный баланс лимитов */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-gradient-to-br from-indigo-950/40 to-slate-900/40 backdrop-blur-md border border-indigo-500/10 p-6 rounded-3xl relative overflow-hidden">
          <div className="absolute top-0 right-0 w-24 h-24 bg-indigo-500/5 blur-xl rounded-full"></div>
          <span className="block text-slate-400 text-xs font-semibold uppercase tracking-wider">Общий долг</span>
          <span className="block mt-2 text-2xl font-bold text-indigo-400">{totalDebt.toLocaleString('ru-RU')} ₸</span>
          <span className="block mt-1 text-[10px] text-slate-500">Сумма всех активных задолженностей</span>
        </div>

        <div className="bg-gradient-to-br from-emerald-950/20 to-slate-900/40 backdrop-blur-md border border-emerald-500/10 p-6 rounded-3xl relative overflow-hidden">
          <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 blur-xl rounded-full"></div>
          <span className="block text-slate-400 text-xs font-semibold uppercase tracking-wider">Доступный кредит</span>
          <span className="block mt-2 text-2xl font-bold text-emerald-400">{totalAvailable.toLocaleString('ru-RU')} ₸</span>
          <span className="block mt-1 text-[10px] text-slate-500">Оставшийся свободный лимит</span>
        </div>

        <div className="bg-slate-900/30 backdrop-blur-md border border-slate-800/80 p-6 rounded-3xl">
          <span className="block text-slate-400 text-xs font-semibold uppercase tracking-wider">Общий лимит</span>
          <span className="block mt-2 text-2xl font-bold text-slate-200">{totalLimit.toLocaleString('ru-RU')} ₸</span>
          <span className="block mt-1 text-[10px] text-slate-500">Совокупный лимит по всем магазинам</span>
        </div>
      </div>

      {/* Очередь требующих внимания действий (Договоры на подпись, Покупки на подтверждение) */}
      {(pendingAgreements.length > 0 || pendingTransactions.length > 0) && (
        <div className="space-y-3">
          <h2 className="text-sm font-bold text-amber-400 uppercase tracking-wider flex items-center gap-1.5">
            <ShieldAlert className="w-4 h-4" />
            <span>Требуется подтверждение (Простая ЭЦП)</span>
          </h2>

          <div className="grid grid-cols-1 gap-3">
            {/* Договоры */}
            {pendingAgreements.map(agreement => (
              <div key={agreement.id} className="bg-slate-900/50 border border-amber-500/20 p-5 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-400 flex-shrink-0">
                    <FileSignature className="w-5 h-5" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-slate-200">Договор с магазином «{agreement.shop_name}»</h3>
                    <p className="text-xs text-slate-400 mt-0.5">
                      Предлагаемый кредитный лимит: <strong className="text-slate-200">{agreement.credit_limit.toLocaleString()} ₸</strong>. 
                      Срок погашения: {new Date(agreement.due_date).toLocaleDateString('ru-RU')}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => startConfirmAgreement(agreement)}
                  className="py-2 px-4 bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-xs rounded-xl shadow-md cursor-pointer transition-all hover:scale-[1.01]"
                >
                  Подписать по SMS
                </button>
              </div>
            ))}

            {/* Транзакции (Покупки в долг / Погашения) */}
            {pendingTransactions.map(tx => (
              <div key={tx.id} className="bg-slate-900/50 border border-indigo-500/20 p-5 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 ${
                    tx.type === 'purchase' ? 'bg-rose-500/10 text-rose-400' : 'bg-emerald-500/10 text-emerald-400'
                  }`}>
                    <Landmark className="w-5 h-5" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-slate-200">
                      {tx.type === 'purchase' ? 'Покупка в долг' : 'Внесение погашения'} в «{tx.shop_name}»
                    </h3>
                    <p className="text-xs text-slate-400 mt-0.5">
                      Сумма операции: <strong className="text-slate-200">{tx.amount.toLocaleString()} ₸</strong>
                      {tx.receipt_image_url && (
                        <a 
                          href={tx.receipt_image_url} 
                          target="_blank" 
                          rel="noreferrer" 
                          className="ml-2 text-indigo-400 hover:underline inline-flex items-center"
                        >
                          (посмотреть фото чека)
                        </a>
                      )}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-auto">
                  <button
                    onClick={() => handleRejectTransaction(tx.id)}
                    className="flex items-center justify-center p-2 rounded-xl bg-slate-800 hover:bg-rose-500/10 hover:text-rose-400 text-slate-400 border border-slate-700/50 hover:border-rose-500/20 transition-all cursor-pointer"
                    title="Отклонить"
                  >
                    <X className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => startConfirmTransaction(tx)}
                    className="py-2 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl shadow-md cursor-pointer transition-all hover:scale-[1.01]"
                  >
                    Подтвердить по SMS
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Список магазинов и активных договоров */}
      <div className="bg-slate-900/20 border border-slate-800/80 rounded-3xl overflow-hidden backdrop-blur-sm">
        <div className="p-6 border-b border-slate-800">
          <h2 className="text-lg font-bold text-slate-200">Мои кредитные линии</h2>
          <p className="text-xs text-slate-400">Список магазинов, где у вас открыты лимиты</p>
        </div>

        {activeAgreements.length > 0 ? (
          <div className="divide-y divide-slate-800/60">
            {activeAgreements.map(agreement => (
              <div 
                key={agreement.id} 
                onClick={() => navigate(`/agreements/${agreement.id}`)}
                className="p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-slate-900/10 transition-all cursor-pointer group"
              >
                <div>
                  <h3 className="font-semibold text-slate-200 group-hover:text-indigo-400 transition-colors">{agreement.shop_name}</h3>
                  <span className="block text-[10px] text-slate-500 mt-0.5">Срок договора до {new Date(agreement.due_date).toLocaleDateString('ru-RU')}</span>
                </div>

                <div className="flex items-center gap-6">
                  <div className="text-right">
                    <span className="block text-[10px] text-slate-400 uppercase tracking-wider font-medium">Ваш долг</span>
                    <span className="font-bold text-sm text-indigo-400">{agreement.balance.toLocaleString()} ₸</span>
                  </div>
                  <div className="text-right border-l border-slate-800 pl-6">
                    <span className="block text-[10px] text-slate-400 uppercase tracking-wider font-medium">Кредитный лимит</span>
                    <span className="font-bold text-sm text-slate-300">{agreement.credit_limit.toLocaleString()} ₸</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-12 text-center text-slate-500">
            <Landmark className="w-12 h-12 mx-auto mb-3 text-slate-700" />
            <p className="text-xs">У вас пока нет активных договоров лимита.</p>
          </div>
        )}
      </div>

      {/* МОДАЛЬНОЕ ОКНО: SMS ПОДТВЕРЖДЕНИЕ */}
      {showOtpModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={() => setShowOtpModal(false)}></div>
          <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-sm p-6 relative z-10 shadow-2xl">
            <h3 className="text-lg font-bold text-slate-100 mb-2">
              {confirmType === 'agreement' ? 'Подписание договора' : 'Подтверждение операции'}
            </h3>
            <p className="text-xs text-slate-400 mb-4">
              {confirmType === 'agreement' 
                ? `Подтвердите подписание кредитного договора в магазине «${confirmTarget?.shop_name}»`
                : `Подтвердите покупку/погашение в магазине «${confirmTarget?.shop_name}» на сумму ${confirmTarget?.amount.toLocaleString()} ₸`}
            </p>

            {otpError && (
              <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs rounded-xl">
                {otpError}
              </div>
            )}

            <form onSubmit={handleOtpSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Код подтверждения из SMS</label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-500">
                    <KeyRound className="w-4 h-4" />
                  </span>
                  <input
                    type="text"
                    maxLength="4"
                    placeholder="••••"
                    value={otpCode}
                    onChange={(e) => setOtpCode(e.target.value)}
                    className="block w-full pl-10 pr-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 tracking-[0.5em] focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                </div>
                <span className="block mt-2 text-[9px] text-amber-400/80 font-medium">
                  * Пожалуйста, скопируйте код из консоли бэкенда!
                </span>
              </div>

              <div className="flex gap-3 justify-end pt-2">
                <button
                  type="button"
                  onClick={() => setShowOtpModal(false)}
                  className="py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-all cursor-pointer"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={otpLoading}
                  className="py-2.5 px-5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/10 transition-all cursor-pointer disabled:opacity-50"
                >
                  {otpLoading ? 'Проверка...' : 'Подтвердить'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
};

export default CustomerDashboard;
