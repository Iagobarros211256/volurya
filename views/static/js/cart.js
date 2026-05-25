document.addEventListener('DOMContentLoaded', async () => {
  if (!isLoggedIn()) {
    goToLogin();
    return;
  }
  await loadCart();
});

async function loadCart() {
  const container = document.getElementById('cart-items');
  const emptyCart = document.getElementById('empty-cart');
  const cartSummary = document.getElementById('cart-summary');
  const loading = document.getElementById('loading');

  loading.style.display = 'block';
  container.innerHTML = '';

  try {
    const res = await fetch(`${API_BASE}/cart`);
    const cart = await res.json();

    loading.style.display = 'none';

    if (!cart.items || cart.items.length === 0) {
      emptyCart.classList.remove('d-none');
      cartSummary.classList.add('d-none');
      return;
    }

    emptyCart.classList.add('d-none');
    cartSummary.classList.remove('d-none');

    let total = 0;

    cart.items.forEach(item => {
      const product = item.product;
      const subtotal = product.price * item.quantity;
      total += subtotal;

      const row = document.createElement('div');
      row.className = 'row align-items-center mb-3 p-3 bg-secondary rounded';
      row.id = `item-${item.id}`;

      // Coluna da imagem
      const imgCol = document.createElement('div');
      imgCol.className = 'col-3 col-md-2';
      const img = document.createElement('img');
      const imgSrc = isValidImageUrl(product.image_url)
        ? product.image_url
        : '/static/imagens/camiseta-feminina-preta.jpg';
      img.src = imgSrc;
      img.className = 'img-fluid rounded';
      img.alt = product.name; // textContent seguro via propriedade
      imgCol.appendChild(img);

      // Coluna do nome e preço
      const infoCol = document.createElement('div');
      infoCol.className = 'col-9 col-md-4 text-start';
      const name = document.createElement('h6');
      name.className = 'text-white mb-1';
      name.textContent = product.name; // seguro
      const price = document.createElement('p');
      price.className = 'text-danger mb-0 fw-bold';
      price.textContent = `R$ ${product.price.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`;
      infoCol.appendChild(name);
      infoCol.appendChild(price);

      // Coluna de quantidade
      const qtyCol = document.createElement('div');
      qtyCol.className = 'col-6 col-md-3 mt-2 mt-md-0';
      const inputGroup = document.createElement('div');
      inputGroup.className = 'input-group input-group-sm';

      const btnMinus = document.createElement('button');
      btnMinus.className = 'btn btn-outline-danger';
      btnMinus.textContent = '-';
      btnMinus.addEventListener('click', () => updateQuantity(item.id, item.quantity - 1));

      const qtyInput = document.createElement('input');
      qtyInput.type = 'number';
      qtyInput.className = 'form-control bg-dark text-white text-center border-danger';
      qtyInput.value = item.quantity;
      qtyInput.min = 1;
      qtyInput.id = `qty-${item.id}`;
      qtyInput.addEventListener('change', () => {
        const val = parseInt(qtyInput.value);
        if (!isNaN(val) && val > 0) {
          updateQuantity(item.id, val);
        }
      });

      const btnPlus = document.createElement('button');
      btnPlus.className = 'btn btn-outline-danger';
      btnPlus.textContent = '+';
      btnPlus.addEventListener('click', () => updateQuantity(item.id, item.quantity + 1));

      inputGroup.appendChild(btnMinus);
      inputGroup.appendChild(qtyInput);
      inputGroup.appendChild(btnPlus);
      qtyCol.appendChild(inputGroup);

      // Coluna do subtotal
      const subtotalCol = document.createElement('div');
      subtotalCol.className = 'col-4 col-md-2 mt-2 mt-md-0 text-end';
      const subtotalP = document.createElement('p');
      subtotalP.className = 'text-white mb-0 fw-bold';
      subtotalP.textContent = `R$ ${subtotal.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`;
      subtotalCol.appendChild(subtotalP);

      // Coluna do botão remover
      const removeCol = document.createElement('div');
      removeCol.className = 'col-2 col-md-1 mt-2 mt-md-0 text-end';
      const removeBtn = document.createElement('button');
      removeBtn.className = 'btn btn-sm btn-outline-danger';
      removeBtn.innerHTML = '<i class="fa fa-trash"></i>'; // ícone estático, sem dados da API
      removeBtn.addEventListener('click', () => removeItem(item.id));
      removeCol.appendChild(removeBtn);

      row.appendChild(imgCol);
      row.appendChild(infoCol);
      row.appendChild(qtyCol);
      row.appendChild(subtotalCol);
      row.appendChild(removeCol);
      container.appendChild(row);
    });

    document.getElementById('cart-total').textContent =
      `R$ ${total.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`;

  } catch (err) {
    loading.style.display = 'none';
    const errP = document.createElement('p');
    errP.className = 'text-danger';
    errP.textContent = 'Erro ao carregar carrinho. Tente novamente.';
    container.appendChild(errP);
  }
}

// Valida se a URL da imagem é segura
function isValidImageUrl(url) {
  if (!url) return false;
  return url.startsWith('https://') || url.startsWith('/');
}

async function updateQuantity(itemId, quantity) {
  if (quantity <= 0) {
    await removeItem(itemId);
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/cart/items/${itemId}`, {
      method: 'PUT',
      body: JSON.stringify({ quantity })
    });

    if (res.ok) {
      await loadCart();
      updateCartBadge();
    } else {
      const data = await res.json();
      showToast(data.error || 'Erro ao atualizar', 'danger');
    }
  } catch {
    showToast('Erro ao atualizar item', 'danger');
  }
}

async function removeItem(itemId) {
  try {
    const res = await fetch(`${API_BASE}/cart/items/${itemId}`, {
      method: 'DELETE'
    });

    if (res.ok) {
      await loadCart();
      updateCartBadge();
    } else {
      showToast('Erro ao remover item', 'danger');
    }
  } catch {
    showToast('Erro ao remover item', 'danger');
  }
}

async function checkout() {
  const btn = document.getElementById('checkout-btn');
  btn.disabled = true;
  btn.textContent = 'Processando...';

  try {
    const res = await fetch(`${API_BASE}/cart/checkout`, {
      method: 'POST'
    });

    const data = await res.json();

    if (res.ok && data.payment_urls && data.payment_urls.length > 0) {
      window.location.href = data.payment_urls[0];
    } else {
      showToast(data.error || 'Erro no checkout', 'danger');
      btn.disabled = false;
      btn.textContent = 'FINALIZAR COMPRA';
    }
  } catch {
    showToast('Erro de conexão', 'danger');
    btn.disabled = false;
    btn.textContent = 'FINALIZAR COMPRA';
  }
}