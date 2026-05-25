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

  const toast = document.createElement('div');
  toast.className = `toast align-items-center text-bg-${type} border-0`;
  toast.setAttribute('role', 'alert');

  const inner = document.createElement('div');
  inner.className = 'd-flex';

  const body = document.createElement('div');
  body.className = 'toast-body';
  body.textContent = message; // textContent — sem XSS

  const closeBtn = document.createElement('button');
  closeBtn.type = 'button';
  closeBtn.className = 'btn-close btn-close-white me-2 m-auto';
  closeBtn.dataset.bsDismiss = 'toast';

  inner.appendChild(body);
  inner.appendChild(closeBtn);
  toast.appendChild(inner);
  container.appendChild(toast);
  document.body.appendChild(container);

  new bootstrap.Toast(toast).show();
  setTimeout(() => container.remove(), 5000);
}

// Intercepta fetch para tratar 401
// Com cookies HttpOnly, o browser envia o cookie automaticamente
// Não precisamos mais adicionar Authorization header manualmente
const originalFetch = window.fetch;
window.fetch = async (url, options = {}) => {
  // Garante que cookies são enviados em todas as requisições
  options.credentials = 'include';

  // Adiciona Content-Type apenas se não for FormData
  if (!options.headers?.['Content-Type'] && !(options.body instanceof FormData)) {
    options.headers = {
      ...options.headers,
      'Content-Type': 'application/json'
    };
  }

  const res = await originalFetch(url, options);

  if (res.status === 401) {
    // Tenta refresh automático antes de redirecionar
    const refreshed = await tryRefreshToken();
    if (refreshed) {
      return originalFetch(url, options);
    }
    rememberCurrentPage();
    showToast('Sessão expirada. Faça login novamente.', 'danger');
    setTimeout(() => window.location.href = '/login', 1500);
    throw new Error('Unauthorized');
  }

  return res;
};

// Tenta renovar o access token usando o refresh token (cookie)
async function tryRefreshToken() {
  try {
    const res = await originalFetch(`${API_BASE}/refresh`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' }
    });
    return res.ok;
  } catch {
    return false;
  }
}

// Verifica se usuário está logado via endpoint leve
// Com cookies HttpOnly, não conseguimos ler o token em JS
// Usamos um endpoint que retorna 200 se o cookie for válido
async function checkAuthStatus() {
  try {
    const res = await originalFetch(`${API_BASE}/auth/me`, {
      credentials: 'include'
    });
    return res.ok;
  } catch {
    return false;
  }
}

// Cache do estado de autenticação para evitar múltiplas requisições
let _authStatus = null;
let _authStatusExpiry = 0;

async function isLoggedIn() {
  const now = Date.now();
  if (_authStatus !== null && now < _authStatusExpiry) {
    return _authStatus;
  }
  _authStatus = await checkAuthStatus();
  _authStatusExpiry = now + 60 * 1000; // cache por 1 minuto
  return _authStatus;
}

function invalidateAuthCache() {
  _authStatus = null;
  _authStatusExpiry = 0;
}

function rememberCurrentPage() {
  const path = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  if (!['/login', '/signup'].includes(window.location.pathname)) {
    sessionStorage.setItem(AUTH_REDIRECT_KEY, path || '/store');
  }
}

function getAuthRedirect(defaultPath = '/store') {
  const redirectTo = sessionStorage.getItem(AUTH_REDIRECT_KEY);
  sessionStorage.removeItem(AUTH_REDIRECT_KEY);
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

async function connectNotifications() {
  if (notificationSource) return;
  if (!await isLoggedIn()) return;

  // Com cookies, o EventSource envia o cookie automaticamente
  // Não precisamos mais do token na query string
  notificationSource = new EventSource(`${API_BASE}/events`, {
    withCredentials: true
  });

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
    setTimeout(async () => {
      if (await isLoggedIn()) connectNotifications();
    }, 5000);
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

async function setupNavbar() {
  const navLists = document.querySelectorAll('.navbar-nav');
  if (navLists.length === 0) return;

  const loggedIn = await isLoggedIn();

  navLists.forEach(nav => {
    const loginLink = nav.querySelector('a[href="/login"]');
    if (loginLink) {
      loginLink.addEventListener('click', () => rememberCurrentPage());
    }
    ensureCartLink(nav);
    syncAuthNav(nav, loggedIn);
  });

  updateCartBadge();
}

function ensureCartLink(nav) {
  let cartLink = nav.querySelector('a[href="/cart"]');
  if (!cartLink) {
    const item = document.createElement('li');
    item.className = 'nav-item';

    const a = document.createElement('a');
    a.className = 'btn btn-outline-danger btn-sm my-2 nav-link position-relative';
    a.href = '/cart';

    const icon = document.createElement('i');
    icon.className = 'fa fa-shopping-cart';

    const badge = document.createElement('span');
    badge.className = 'cart-badge badge rounded-pill bg-danger position-absolute top-0 start-100 translate-middle d-none';
    badge.textContent = '0';

    a.appendChild(icon);
    a.append(' CARRINHO');
    a.appendChild(badge);
    item.appendChild(a);

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

function syncAuthNav(nav, loggedIn) {
  const loginLink = nav.querySelector('a[href="/login"]');
  if (!loginLink) return;

  if (!loggedIn) {
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

  if (!await isLoggedIn()) {
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

async function logout() {
  if (notificationSource) {
    notificationSource.close();
    notificationSource = null;
  }

  try {
    await originalFetch(`${API_BASE}/logout`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' }
    });
  } catch {
    // logout local mesmo se o servidor falhar
  }

  invalidateAuthCache();
  setCartBadge(0);
  showToast('Você saiu!', 'info');
  setTimeout(() => window.location.href = '/login', 1500);
}

async function login(email, password) {
  const res = await originalFetch(`${API_BASE}/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Erro ao fazer login');
  }

  invalidateAuthCache();
  await connectNotifications();
  await updateCartBadge();
}

// Inicialização de toggle de senha — reutilizável em login e signup
function initPasswordToggles() {
  document.querySelectorAll('[data-password-toggle]').forEach((button) => {
    button.addEventListener('click', () => {
      const input = document.getElementById(button.dataset.passwordToggle);
      const icon = button.querySelector('.fa');
      const isHidden = input.type === 'password';
      input.type = isHidden ? 'text' : 'password';
      button.setAttribute('aria-label', isHidden ? 'Ocultar senha' : 'Mostrar senha');
      icon.classList.toggle('fa-eye', !isHidden);
      icon.classList.toggle('fa-eye-slash', isHidden);
    });
  });
}

document.addEventListener('DOMContentLoaded', () => {
  setupNavbar();
  connectNotifications();
  initPasswordToggles();
});