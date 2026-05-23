document.addEventListener('DOMContentLoaded', async () => {
  const container = document.getElementById('store-items');
  const loading = document.getElementById('loading');
  const noProducts = document.getElementById('no-products');

  loading.style.display = 'block';
  container.innerHTML = '';
  noProducts.classList.add('d-none');

  try {
    const res = await fetch('/api/products?limit=12');

    if (!res.ok) throw new Error('Erro ao carregar produtos');

    const { data } = await res.json();

    if (data.length === 0) {
      noProducts.classList.remove('d-none');
      return;
    }

    data.forEach(product => {
      const col = document.createElement('div');
      col.className = 'col';
      col.innerHTML = `
        <div class="card bg-dark border-danger h-100 shadow-lg">
          <img src="/static/imagens/camiseta-feminina-preta.jpg" class="card-img-top" alt="${product.name}">
          <div class="card-body d-flex flex-column">
            <h5 class="card-title text-danger">${product.name}</h5>
            <p class="card-text flex-grow-1">${product.description || 'Produto oficial VOLURYA'}</p>
            <div class="mt-auto">
              <p class="card-text fw-bold fs-4 text-success">R$ ${product.price.toFixed(2)}</p>
              <p class="card-text text-muted small">Estoque: ${product.stock}</p>
              <button class="btn btn-outline-danger w-100 buy-btn" data-id="${product.id}">
                Comprar
              </button>
            </div>
          </div>
        </div>
      `;
      container.appendChild(col);

      // listener adicionado direto no botão após criação
      col.querySelector('.buy-btn').addEventListener('click', async () => {
        if (!isLoggedIn()) {
          goToLogin();
          return;
        }

        try {
          const res = await fetch('/api/orders', {
            method: 'POST',
            body: JSON.stringify({ product_id: parseInt(product.id), quantity: 1 })
          });

          const data = await res.json();
          if (res.ok) {
            updateCartBadge();
            window.location.href = data.payment_url;
          } else {
            showToast(data.error, 'danger');
          }
        } catch (err) {
          showToast(err.message, 'danger');
        }
      });
    });

  } catch (err) {
    container.innerHTML = `<p class="col-12 text-center text-danger fs-4">Erro ao carregar produtos: ${err.message}</p>`;
  } finally {
    loading.style.display = 'none';
  }
});


/*

XSS via innerHTML com dados da API
javascriptcol.innerHTML = `
    <h5 class="card-title text-danger">${product.name}</h5>
    <p class="card-text flex-grow-1">${product.description || '...'}</p>
product.name e product.description não sanitizados — mesmo problema do cart.js e store-admin.js. Use textContent para dados dinâmicos.

🔴 err.message renderizado como HTML
javascriptcontainer.innerHTML = `<p ...>Erro ao carregar produtos: ${err.message}</p>`
Mesmo padrão inseguro dos outros arquivos JS.

🔴 Imagem hardcoded — ignora product.image_url
javascript<img src="/static/imagens/camiseta-feminina-preta.jpg" alt="${product.name}">
O backend tem toda a infraestrutura de upload e processamento de imagens — R2 storage, worker pool, migration — mas a loja sempre mostra a mesma imagem estática. Todo esse trabalho de backend não está sendo usado:
javascriptconst imgSrc = product.image_url || '/static/imagens/camiseta-feminina-preta.jpg'

🔴 Compra sem adicionar ao carrinho
javascriptconst res = await fetch('/api/orders', {
    method: 'POST',
    body: JSON.stringify({ product_id: parseInt(product.id), quantity: 1 })
})
window.location.href = data.payment_url
O botão "Comprar" cria uma ordem diretamente — bypassa o carrinho completamente. O usuário não pode escolher quantidade, não vê o total, não pode revisar antes de pagar. O fluxo correto seria adicionar ao carrinho:
javascriptawait fetch('/api/cart/items', {
    method: 'POST',
    body: JSON.stringify({ product_id: product.id, quantity: 1 })
})
window.location.href = '/cart'

🟡 Sem paginação — carrega apenas os primeiros 12 produtos
javascriptfetch('/api/products?limit=12')
Carrega só 12 e não há botão de "carregar mais" ou paginação. O cursor-based pagination implementado no backend não é usado.

🟡 parseInt(product.id) desnecessário
javascriptproduct_id: parseInt(product.id)
product.id já vem como número do JSON — parseInt é redundante e pode mascarar problemas se o tipo mudar.

🟡 alt do img não sanitizado
javascript<img src="..." alt="${product.name}">
Dentro de innerHTML, alt com aspas no nome pode quebrar o HTML: um produto chamado "Test" onload="..." injeta atributos. Com textContent isso não ocorre — mais um motivo para evitar innerHTML.

🟡 Sem tratamento de estoque esgotado
javascript<button class="btn btn-outline-danger w-100 buy-btn" data-id="${product.id}">
    Comprar
</button>
Produtos com stock === 0 mostram o botão ativo. Deveria desabilitar:
javascriptconst outOfStock = product.stock === 0
// button disabled + texto "Esgotado"




*/