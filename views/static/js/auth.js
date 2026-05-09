// static/js/auth.js - ÚNICO arquivo de autenticação

const API_BASE = window.location.hostname.includes('localhost')
  ? 'http://localhost:8080/api'
  : 'https://volurya.onrender.com/api';

// Toast global (reutilizável)
function showToast(message, type = 'success') {
  const container = document.createElement('div');
  container.className = 'position-fixed bottom-0 end-0 p-3';
  container.style.zIndex = '9999';
  container.innerHTML = `
    <div class="toast align-items-center text-bg-${type} border-0" role="alert">
      <div class="d-flex">
        <div class="toast-body">${message}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>
  `;
  document.body.appendChild(container);
  new bootstrap.Toast(container.querySelector('.toast')).show();
  setTimeout(() => container.remove(), 5000);
}

// Intercepta fetch para adicionar token + tratar 401
const originalFetch = window.fetch;
window.fetch = async (url, options = {}) => {
  const token = localStorage.getItem('token');
  if (token) {
    options.headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  }

  const res = await originalFetch(url, options);

  if (res.status === 401) {
    localStorage.removeItem('token');
    showToast('Sessão expirada. Faça login novamente.', 'danger');
    setTimeout(() => window.location.href = '/login', 1500);
    throw new Error('Unauthorized');
  }

  return res;
};

// Verifica token válido
function isLoggedIn() {
  const token = localStorage.getItem('token');
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 > Date.now();
  } catch {
    localStorage.removeItem('token');
    return false;
  }
}

// Logout
function logout() {
  localStorage.removeItem('token');
  showToast('Você saiu!', 'info');
  setTimeout(() => window.location.href = '/login', 1500);
}

// Login (reutilizável)
async function login(email, password) {
  const res = await fetch(`${API_BASE}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Erro ao fazer login');
  }

  const data = await res.json();
  localStorage.setItem('token', data.access_token);
  return data;
}