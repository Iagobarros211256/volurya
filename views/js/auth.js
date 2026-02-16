// js/auth.js

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
