const API_PREFIX = '/api/v1';

// Вспомогательный метод для получения заголовков запроса
const getHeaders = (isMultipart = false) => {
  const headers = {};
  if (!isMultipart) {
    headers['Content-Type'] = 'application/json';
  }
  const token = localStorage.getItem('token');
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
};

// Вспомогательный метод обработки ответа сервера
const handleResponse = async (response) => {
  const contentType = response.headers.get('content-type');
  let data = null;
  
  if (contentType && contentType.includes('application/json')) {
    data = await response.json();
  } else {
    data = await response.text();
  }

  if (!response.ok) {
    const errorMsg = (data && data.error) || 'Произошла непредвиденная ошибка';
    throw new Error(errorMsg);
  }

  return data;
};

export const api = {
  // --- АУТЕНТИФИКАЦИЯ ---
  
  // Регистрация владельца магазина
  registerOwner: async (phone, password, iin, fullName) => {
    const res = await fetch(`${API_PREFIX}/auth/register-owner`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ phone, password, iin, full_name: fullName }),
    });
    return handleResponse(res);
  },

  // Инициация регистрации покупателя (отправка OTP)
  registerCustomer: async (phone, iin, fullName) => {
    const res = await fetch(`${API_PREFIX}/auth/register-customer`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ phone, iin, full_name: fullName }),
    });
    return handleResponse(res);
  },

  // Подтверждение OTP для покупателя
  verifyCustomerRegister: async (phone, code) => {
    const res = await fetch(`${API_PREFIX}/auth/verify-registration`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ phone, code }),
    });
    return handleResponse(res);
  },

  // Установка PIN-кода для покупателя
  setPin: async (pin) => {
    const res = await fetch(`${API_PREFIX}/auth/set-pin`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ pin }),
    });
    return handleResponse(res);
  },

  // Авторизация пользователя (пароль для В, PIN-код для П)
  login: async (phone, password) => {
    const res = await fetch(`${API_PREFIX}/auth/login`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ phone, password }),
    });
    return handleResponse(res);
  },

  // Получить данные текущего профиля
  getMe: async () => {
    const res = await fetch(`${API_PREFIX}/auth/me`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // --- МАГАЗИНЫ ---

  // Создать магазин (Администратор)
  createShop: async (ownerId, name, address) => {
    const res = await fetch(`${API_PREFIX}/shops`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ owner_id: ownerId, name, address }),
    });
    return handleResponse(res);
  },

  // Список магазинов владельца
  getShops: async () => {
    const res = await fetch(`${API_PREFIX}/shops`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // --- ДОГОВОРЫ ---

  // Создать договор (Владелец)
  createAgreement: async (shopId, customerPhone, creditLimit, dueDate) => {
    const res = await fetch(`${API_PREFIX}/agreements`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        shop_id: shopId,
        customer_phone: customerPhone,
        credit_limit: parseFloat(creditLimit),
        due_date: dueDate, // формат YYYY-MM-DD
      }),
    });
    return handleResponse(res);
  },

  // Подтвердить договор по SMS-коду (Покупатель)
  confirmAgreement: async (agreementId, code) => {
    const res = await fetch(`${API_PREFIX}/agreements/${agreementId}/confirm`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ code }),
    });
    return handleResponse(res);
  },

  // Список договоров (Владелец видит по своим магазинам, покупатель свои)
  getAgreements: async () => {
    const res = await fetch(`${API_PREFIX}/agreements`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // Детали договора
  getAgreementById: async (agreementId) => {
    const res = await fetch(`${API_PREFIX}/agreements/${agreementId}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // --- ТРАНЗАКЦИИ ---

  // Загрузка фото чека на сервер
  uploadReceipt: async (file) => {
    const formData = new FormData();
    formData.append('receipt', file);

    const res = await fetch(`${API_PREFIX}/upload`, {
      method: 'POST',
      headers: getHeaders(true),
      body: formData,
    });
    return handleResponse(res);
  },

  // Создать покупку или погашение долга (Владелец)
  createTransaction: async (agreementId, type, amount, receiptImageUrl = null) => {
    const body = {
      agreement_id: agreementId,
      type,
      amount: parseFloat(amount),
    };
    if (receiptImageUrl) {
      body.receipt_image_url = receiptImageUrl;
    }
    const res = await fetch(`${API_PREFIX}/transactions`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify(body),
    });
    return handleResponse(res);
  },

  // Подтвердить транзакцию (Покупатель)
  confirmTransaction: async (transactionId, code) => {
    const res = await fetch(`${API_PREFIX}/transactions/${transactionId}/confirm`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ code }),
    });
    return handleResponse(res);
  },

  // Отклонить транзакцию (Покупатель)
  rejectTransaction: async (transactionId) => {
    const res = await fetch(`${API_PREFIX}/transactions/${transactionId}/reject`, {
      method: 'POST',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // Получить историю транзакций по договору
  getTransactions: async (agreementId) => {
    const res = await fetch(`${API_PREFIX}/transactions?agreement_id=${agreementId}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  // --- АДМИНИСТРАТИВНЫЕ МЕТОДЫ ---
  getOwners: async () => {
    const res = await fetch(`${API_PREFIX}/admin/owners`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  getCustomers: async () => {
    const res = await fetch(`${API_PREFIX}/admin/customers`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  getAllShops: async () => {
    const res = await fetch(`${API_PREFIX}/admin/shops`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },

  getActiveAgreementByCID: async (cid, shopId) => {
    const res = await fetch(`${API_PREFIX}/agreements/active?cid=${cid}&shop_id=${shopId}`, {
      method: 'GET',
      headers: getHeaders(),
    });
    return handleResponse(res);
  },
};
