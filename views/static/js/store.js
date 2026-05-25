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

      // Card
      const card = document.createElement('div');
      card.className = 'card bg-dark border-danger h-100 shadow-lg';

      // Imagem — usa image_url real se disponível e segura
      const img = document.createElement('img');
      img.src = isValidImageUrl(product.image_url)
        ? product.image_url
        : '/static/imagens/camiseta-feminina-preta.jpg';
      img.className = 'card-img-top';
      img.alt = product.name; // propriedade — não interpreta HTML

      // Card body
      const cardBody = document.createElement('div');
      cardBody.className = 'card-body d-flex flex-column';

      // Nome
      const title = document.createElement('h5');
      title.className = 'card-title text-danger';
      title.textContent = product.name; // seguro

      // Descrição
      const desc = document.createElement('p');
      desc.className = 'card-text flex-grow-1';
      desc.textContent = product.description || 'Produto oficial VOLURYA'; // seguro

      // Preço
      const priceWrapper = document.createElement('div');
      priceWrapper.className = 'mt-auto';

      const priceP = document.createElement('p');
      priceP.className = 'card-text fw-bold fs-4 text-success';
      priceP.textContent = product.price.toLocaleString('pt-BR', {
        style: 'currency',
        currency: 'BRL'
      });

      // Estoque
      const stockP = document.createElement('p');
      stockP.className = 'card-text text-muted small';
      const outOfStock = product.stock === 0;
      stockP.textContent = outOfStock ? 'Esgotado' : `Estoque: ${product.stock}`;

      // Botão comprar
      const btn = document.createElement('button');
      btn.className = 'btn btn-outline-danger w-100';
      btn.textContent = outOfStock ? 'Esgotado' : 'Adicionar ao Carrinho';
      btn.disabled = outOfStock;

      btn.addEventListener('click', async () => {
        if (!isLoggedIn()) {
          goToLogin();
          return;
        }

        btn.disabled = true;
        btn.textContent = 'Adicionando...';

        try {
          const res = await fetch('/api/cart/items', {
            method: 'POST',
            body: JSON.stringify({ product_id: product.id, quantity: 1 })
          });

          const resData = await res.json();

          if (res.ok) {
            updateCartBadge();
            showToast('Produto adicionado ao carrinho!', 'success');
            btn.textContent = 'Adicionado ✓';
            setTimeout(() => {
              btn.disabled = outOfStock;
              btn.textContent = outOfStock ? 'Esgotado' : 'Adicionar ao Carrinho';
            }, 2000);
          } else {
            showToast(resData.error || 'Erro ao adicionar', 'danger');
            btn.disabled = outOfStock;
            btn.textContent = outOfStock ? 'Esgotado' : 'Adicionar ao Carrinho';
          }
        } catch {
          showToast('Erro de conexão', 'danger');
          btn.disabled = outOfStock;
          btn.textContent = outOfStock ? 'Esgotado' : 'Adicionar ao Carrinho';
        }
      });

      priceWrapper.appendChild(priceP);
      priceWrapper.appendChild(stockP);
      priceWrapper.appendChild(btn);

      cardBody.appendChild(title);
      cardBody.appendChild(desc);
      cardBody.appendChild(priceWrapper);

      card.appendChild(img);
      card.appendChild(cardBody);
      col.appendChild(card);
      container.appendChild(col);
    });

  } catch {
    const errP = document.createElement('p');
    errP.className = 'col-12 text-center text-danger fs-4';
    errP.textContent = 'Erro ao carregar produtos. Tente novamente.';
    container.appendChild(errP);
  } finally {
    loading.style.display = 'none';
  }
});

function isValidImageUrl(url) {
  if (!url) return false;
  return url.startsWith('https://') || url.startsWith('/');
}