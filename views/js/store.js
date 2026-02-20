// static/js/store.js (atualizado)

document.addEventListener('DOMContentLoaded', async () => {
  const container = document.getElementById('store-items');
  const loading = document.getElementById('loading');

  try {
    loading.style.display = 'block';

    const res = await fetch(`${API_BASE}/products?limit=12`);
    if (!res.ok) throw new Error('Erro ao carregar produtos');

    const { data } = await res.json();

    container.innerHTML = '';

    if (data.length === 0) {
      container.innerHTML = '<p class="col-12 text-center text-muted fs-4">Nenhum produto disponível no momento. Volte em breve!</p>';
    } else {
      data.forEach(product => {
        const col = document.createElement('div');
        col.className = 'col';
        col.innerHTML = `
          <div class="card bg-dark border-danger h-100 shadow-lg hover-glow">
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
      });
    }
  } catch (err) {
    container.innerHTML = `<p class="col-12 text-center text-danger fs-4">Erro ao carregar produtos: ${err.message}</p>`;
  } finally {
    loading.style.display = 'none';
  }
});