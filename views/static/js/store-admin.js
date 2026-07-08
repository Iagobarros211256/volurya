// static/js/store-admin.js

// Declaração explícita das variáveis globais
const loginSection = document.getElementById('login-section') || document.getElementById('loginSection');
const adminSection = document.getElementById('admin-section') || document.getElementById('adminSection');
const productsList = document.getElementById('products-list') || document.getElementById('productsList');
const productForm = document.getElementById('productForm');

// Verificação de role admin via /api/auth/me
async function isAdmin() {
  try {
    const res = await originalFetch(`${API_BASE}/auth/me`, {
      credentials: 'include'
    });
    if (!res.ok) return false;
    const data = await res.json();
    return data.role === 'admin';
  } catch {
    return false;
  }
}

;(async () => {
  const loggedIn = await isLoggedIn();
  const admin = loggedIn && await isAdmin();
  if (!admin) {
    if (loginSection) loginSection.style.display = 'block';
    if (adminSection) adminSection.style.display = 'none';
  } else {
    if (loginSection) loginSection.style.display = 'none';
    if (adminSection) adminSection.style.display = 'block';
    loadProducts();
  }
})();

// Login
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('email').value.trim();
  const password = document.getElementById('password').value; // sem trim em senha

  try {
    await login(email, password);

    if (!await isAdmin()) {
      showToast('Acesso restrito a administradores', 'danger');
      logout();
      return;
    }

    showToast('Login realizado!', 'success');
    if (loginSection) loginSection.style.display = 'none';
    if (adminSection) adminSection.style.display = 'block';
    loadProducts();
  } catch (err) {
    showToast(err.message || 'Erro ao fazer login', 'danger');
  }
});

// Carrega produtos
async function loadProducts() {
  if (!productsList) return;

  try {
    const res = await fetch(`${API_BASE}/products`);
    if (!res.ok) throw new Error('Erro ao carregar produtos');
    const { data } = await res.json();

    productsList.innerHTML = '';

    if (data.length === 0) {
      const p = document.createElement('p');
      p.textContent = 'Nenhum produto cadastrado.';
      productsList.appendChild(p);
      return;
    }

    data.forEach(p => {
      const div = document.createElement('div');
      div.className = 'd-flex justify-content-between align-items-center mb-2 p-2 bg-secondary rounded';

      // Info do produto — textContent para evitar XSS
      const info = document.createElement('div');
      info.textContent = `${p.name} - ${p.price.toLocaleString('pt-BR', {
        style: 'currency',
        currency: 'BRL'
      })} - Estoque: ${p.stock}`;

      // Botões
      const btnGroup = document.createElement('div');

      const editBtn = document.createElement('button');
      editBtn.className = 'btn btn-sm btn-warning me-1';
      editBtn.textContent = 'Editar';
      editBtn.dataset.id = p.id;
      editBtn.addEventListener('click', () => editProduct(p));

      const deleteBtn = document.createElement('button');
      deleteBtn.className = 'btn btn-sm btn-danger';
      deleteBtn.textContent = 'Excluir';
      deleteBtn.dataset.id = p.id;
      deleteBtn.addEventListener('click', () => deleteProduct(p.id, p.name));

      btnGroup.appendChild(editBtn);
      btnGroup.appendChild(deleteBtn);

      div.appendChild(info);
      div.appendChild(btnGroup);
      productsList.appendChild(div);
    });

  } catch {
    showToast('Erro ao carregar produtos', 'danger');
  }
}

// Adicionar produto
productForm?.addEventListener('submit', async (e) => {
  e.preventDefault();

  const name = document.getElementById('productName').value.trim();
  const desc = document.getElementById('productDesc').value.trim();
  const price = parseFloat(document.getElementById('productPrice').value);
  const stock = parseInt(document.getElementById('productStock').value);

  if (!name || isNaN(price) || price <= 0 || isNaN(stock) || stock < 0) {
    showToast('Preencha corretamente todos os campos', 'danger');
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/products`, {
      method: 'POST',
      body: JSON.stringify({ name, description: desc, price, stock })
    });

    if (!res.ok) {
      const data = await res.json();
      showToast(data.error || 'Erro ao criar produto', 'danger');
      return;
    }

    showToast('Produto adicionado!', 'success');
    productForm.reset();
    loadProducts();
  } catch {
    showToast('Erro de conexão', 'danger');
  }
});

// Editar produto
function editProduct(product) {
  document.getElementById('productName').value = product.name;
  document.getElementById('productDesc').value = product.description || '';
  document.getElementById('productPrice').value = product.price;
  document.getElementById('productStock').value = product.stock;

  const saveBtn = document.getElementById('saveProductBtn');
  if (saveBtn) {
    saveBtn.textContent = 'Atualizar';
    saveBtn.onclick = async () => {
      const name = document.getElementById('productName').value.trim();
      const desc = document.getElementById('productDesc').value.trim();
      const price = parseFloat(document.getElementById('productPrice').value);
      const stock = parseInt(document.getElementById('productStock').value);

      if (!name || isNaN(price) || price <= 0 || isNaN(stock) || stock < 0) {
        showToast('Preencha corretamente todos os campos', 'danger');
        return;
      }

      try {
        const res = await fetch(`${API_BASE}/products/${product.id}`, {
          method: 'PUT',
          body: JSON.stringify({ name, description: desc, price, stock })
        });

        if (!res.ok) {
          const data = await res.json();
          showToast(data.error || 'Erro ao atualizar', 'danger');
          return;
        }

        showToast('Produto atualizado!', 'success');
        productForm.reset();
        saveBtn.textContent = 'Salvar';
        saveBtn.onclick = null;
        loadProducts();

        // Fecha modal se existir
        const modal = bootstrap.Modal.getInstance(document.getElementById('productModal'));
        if (modal) modal.hide();

      } catch {
        showToast('Erro de conexão', 'danger');
      }
    };

    // Abre modal se existir
    const modal = new bootstrap.Modal(document.getElementById('productModal'));
    if (modal) modal.show();
  }
}

// Deletar produto com confirmação
async function deleteProduct(id, name) {
  if (!confirm(`Deletar "${name}"? Esta ação não pode ser desfeita.`)) return;

  try {
    const res = await fetch(`${API_BASE}/products/${id}`, {
      method: 'DELETE'
    });

    if (!res.ok) {
      const data = await res.json();
      showToast(data.error || 'Erro ao deletar', 'danger');
      return;
    }

    showToast('Produto deletado!', 'success');
    loadProducts();
  } catch {
    showToast('Erro de conexão', 'danger');
  }
}