// Inicializa
const loginSection = document.getElementById('login-section');
const adminSection = document.getElementById('admin-section');

// LOGIN
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  try {
    const res = await fetch('https://volurya.onrender.com/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const data = await res.json();
    if (res.ok) {
      localStorage.setItem('token', data.access_token);
      alert('Login success!');
      showAdmin();
    } else {
      alert(data.error);
    }
  } catch (err) {
    console.error(err);
  }
});

// Mostrar painel admin
function showAdmin() {
  loginSection.style.display = 'none';
  adminSection.style.display = 'block';
  loadProducts();
}

// FETCH autenticado
async function authFetch(url, options = {}) {
  const token = localStorage.getItem('token');
  options.headers = { 
    ...options.headers, 
    'Authorization': `Bearer ${token}`, 
    'Content-Type': 'application/json' 
  };
  return fetch(url, options);
}

// ADICIONAR PRODUTO
document.getElementById('productForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const name = document.getElementById('productName').value;
  const desc = document.getElementById('productDesc').value;
  const price = parseFloat(document.getElementById('productPrice').value);
  const stock = parseInt(document.getElementById('productStock').value);

  try {
    const res = await authFetch('https://volurya.onrender.com/products', {
      method: 'POST',
      body: JSON.stringify({ name, description: desc, price, stock })
    });
    const data = await res.json();
    if (res.ok) {
      alert('Product added!');
      loadProducts();
    } else {
      alert(data.error);
    }
  } catch (err) {
    console.error(err);
  }
});

// CARREGAR PRODUTOS EXISTENTES
async function loadProducts() {
  const res = await fetch('https://volurya.onrender.com/products');
  const products = await res.json();
  const container = document.getElementById('products-list');
  container.innerHTML = '';

  products.forEach(p => {
    const div = document.createElement('div');
    div.classList.add('mb-2');
    div.innerHTML = `
      <strong>${p.name}</strong> - $${p.price} - Stock: ${p.stock}
      <button class="btn btn-sm btn-warning" onclick="updateProduct(${p.id})">Edit</button>
      <button class="btn btn-sm btn-danger" onclick="deleteProduct(${p.id})">Delete</button>
    `;
    container.appendChild(div);
  });
}

// ATUALIZAR PRODUTO (simplificado prompt)
async function updateProduct(id) {
  const name = prompt('New name:');
  const desc = prompt('New description:');
  const price = parseFloat(prompt('New price:'));
  const stock = parseInt(prompt('New stock:'));

  try {
    await authFetch(`https://volurya.onrender.com/products/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name, description: desc, price, stock })
    });
    loadProducts();
  } catch (err) {
    console.error(err);
  }
}

// DELETAR PRODUTO
async function deleteProduct(id) {
  if (!confirm('Delete this product?')) return;
  try {
    await authFetch(`https://volurya.onrender.com/products/${id}`, { method: 'DELETE' });
    loadProducts();
  } catch (err) {
    console.error(err);
  }
}

// AUTO-CHECK login
if (localStorage.getItem('token')) {
  showAdmin();
}
