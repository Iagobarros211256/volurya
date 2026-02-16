// auth.js

// Salva token no localStorage ao logar (já existente)
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  try {
    const response = await fetch('https://volurya.onrender.com/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });

    const data = await response.json();
    if (response.ok) {
      localStorage.setItem('token', data.access_token);
      alert('Login realizado com sucesso!');
    } else {
      alert(data.error || 'Erro no login');
    }
  } catch (err) {
    console.error('Erro login:', err);
  }
});

// Função para obter token
function getToken() {
  return localStorage.getItem('token');
}

// Função utilitária para fetch autenticado
async function authFetch(url, options = {}) {
  const token = getToken();
  options.headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  return fetch(url, options);
}
