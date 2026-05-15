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
