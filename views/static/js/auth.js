// static/js/auth.js - ÚNICO arquivo de autenticação

const API_BASE = window.location.hostname.includes('localhost')
  ? 'http://localhost:8080/api'
  : 'https://volurya.onrender.com/api';

const AUTH_REDIRECT_KEY = 'auth_redirect_to';

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
    rememberCurrentPage();
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

function rememberCurrentPage() {
  const path = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  if (!['/login', '/signup'].includes(window.location.pathname)) {
    localStorage.setItem(AUTH_REDIRECT_KEY, path || '/store');
  }
}

function getAuthRedirect(defaultPath = '/store') {
  const redirectTo = localStorage.getItem(AUTH_REDIRECT_KEY);
  localStorage.removeItem(AUTH_REDIRECT_KEY);

  if (!redirectTo || redirectTo.startsWith('/login') || redirectTo.startsWith('/signup')) {
    return defaultPath;
  }

  return redirectTo;
}

function goToLogin() {
  rememberCurrentPage();
  window.location.href = '/login';
}

let notificationSource;

function connectNotifications() {
  if (notificationSource || !isLoggedIn()) return;

  const token = localStorage.getItem('token');
  notificationSource = new EventSource(`${API_BASE}/events?access_token=${encodeURIComponent(token)}`);

  notificationSource.onmessage = handleNotificationEvent;

  [
    'cart_item_added',
    'cart_item_updated',
    'cart_item_removed',
    'checkout_created',
    'order_created'
  ].forEach(eventType => {
    notificationSource.addEventListener(eventType, handleNotificationEvent);
  });

  notificationSource.onerror = () => {
    notificationSource.close();
    notificationSource = null;
    if (isLoggedIn()) {
      setTimeout(connectNotifications, 5000);
    }
  };
}

function handleNotificationEvent(event) {
  try {
    const data = JSON.parse(event.data);
    if (data.type === 'connected') return;
    showToast(data.message || 'Nova notificação', 'info');
    updateCartBadge();
  } catch {
    showToast('Nova notificação', 'info');
  }
}

function setupNavbar() {
  const navLists = document.querySelectorAll('.navbar-nav');
  if (navLists.length === 0) return;

  navLists.forEach(nav => {
    const loginLink = nav.querySelector('a[href="/login"]');
    if (loginLink) {
      loginLink.addEventListener('click', () => rememberCurrentPage());
    }

    ensureCartLink(nav);
    syncAuthNav(nav);
  });

  updateCartBadge();
}

function ensureCartLink(nav) {
  let cartLink = nav.querySelector('a[href="/cart"]');
  if (!cartLink) {
    const item = document.createElement('li');
    item.className = 'nav-item';
    item.innerHTML = `
      <a class="btn btn-outline-danger btn-sm my-2 nav-link position-relative" href="/cart">
        <i class="fa fa-shopping-cart"></i>
        CARRINHO
        <span class="cart-badge badge rounded-pill bg-danger position-absolute top-0 start-100 translate-middle d-none">0</span>
      </a>
    `;

    const loginItem = nav.querySelector('a[href="/login"]')?.closest('li');
    nav.insertBefore(item, loginItem || null);
    return;
  }

  cartLink.classList.add('position-relative');
  if (!cartLink.querySelector('.cart-badge')) {
    const badge = document.createElement('span');
    badge.className = 'cart-badge badge rounded-pill bg-danger position-absolute top-0 start-100 translate-middle d-none';
    badge.textContent = '0';
    cartLink.appendChild(badge);
  }
}

function syncAuthNav(nav) {
  const loginLink = nav.querySelector('a[href="/login"]');
  if (!loginLink) return;

  if (!isLoggedIn()) {
    loginLink.textContent = 'Entrar';
    loginLink.classList.remove('btn-outline-secondary');
    loginLink.classList.add('btn-outline-primary');
    return;
  }

  loginLink.textContent = 'Sair';
  loginLink.href = '#logout';
  loginLink.classList.remove('btn-outline-primary');
  loginLink.classList.add('btn-outline-secondary');
  loginLink.addEventListener('click', (event) => {
    event.preventDefault();
    logout();
  });
}

async function updateCartBadge() {
  const badges = document.querySelectorAll('.cart-badge');
  if (badges.length === 0) return;

  if (!isLoggedIn()) {
    setCartBadge(0);
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/cart`);
    if (!res.ok) {
      setCartBadge(0);
      return;
    }

    const cart = await res.json();
    const count = (cart.items || []).reduce((total, item) => total + item.quantity, 0);
    setCartBadge(count);
  } catch {
    setCartBadge(0);
  }
}

function setCartBadge(count) {
  document.querySelectorAll('.cart-badge').forEach(badge => {
    badge.textContent = count;
    badge.classList.toggle('d-none', count === 0);
  });
}

// Logout
function logout() {
  if (notificationSource) {
    notificationSource.close();
    notificationSource = null;
  }
  localStorage.removeItem('token');
  localStorage.removeItem('refresh_token');
  setCartBadge(0);
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
  connectNotifications();
  updateCartBadge();
  return data;
}

document.addEventListener('DOMContentLoaded', () => {
  setupNavbar();
  connectNotifications();
});
