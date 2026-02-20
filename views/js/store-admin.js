// static/js/store-admin.js

if (!isLoggedIn()) {
  loginSection.style.display = 'block';
  adminSection.style.display = 'none';
} else {
  loginSection.style.display = 'none';
  adminSection.style.display = 'block';
  loadProducts();
}

// Login
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('email').value.trim();
  const password = document.getElementById('password').value.trim();

  try {
    await login(email, password);
    showToast('Login realizado!', 'success');
    loginSection.style.display = 'none';
    adminSection.style.display = 'block';
    loadProducts();
  } catch (err) {
    showToast(err.message, 'danger');
  }
});

// Carrega produtos
async function loadProducts() {
  try {
    const res = await fetch(`${API_BASE}/products`);
    if (!res.ok) throw new Error('Erro ao carregar');
    const { data } = await res.json();

    productsList.innerHTML = '';
    if (data.length === 0) {
      productsList.innerHTML = '<p>Nenhum produto cadastrado.</p>';
      return;
    }

    data.forEach(p => {
      const div = document.createElement('div');
      div.className = 'd-flex justify-content-between align-items-center mb-2 p-2 bg-secondary rounded';
      div.innerHTML = `
        <div>${p.name} - R$ ${p.price.toFixed(2)} - Estoque: ${p.stock}</div>
        <div>
          <button class="btn btn-sm btn-warning edit-btn" data-id="${p.id}">Editar</button>
          <button class="btn btn-sm btn-danger delete-btn" data-id="${p.id}">Excluir</button>
        </div>
      `;
      productsList.appendChild(div);
    });
  } catch (err) {
    showToast(err.message, 'danger');
  }
}

// Adicionar produto (usando modal)
document.getElementById('productForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const name = document.getElementById('productName').value.trim();
  const desc = document.getElementById('productDesc').value.trim();
  const price = parseFloat(document.getElementById('productPrice').value);
  const stock = parseInt(document.getElementById('productStock').value);

  if (!name || isNaN(price) || price <= 0 || isNaN(stock) || stock < 0) {
    showToast('Preencha corretamente', 'danger');
    return;
  }

  try {
    await fetch(`${API_BASE}/products`, {
      method: 'POST',
      body: JSON.stringify({ name, description: desc, price, stock })
    });
    showToast('Produto adicionado!', 'success');
    productForm.reset();
    loadProducts();
  } catch (err) {
    showToast(err.message, 'danger');
  }
});