const API_BASE = window.location.hostname.includes('localhost')
  ? 'http://localhost:8080/api'
  : 'https://volurya.onrender.com/api';

document.addEventListener('DOMContentLoaded', async () => {
  // Redireciona se não estiver logado
  if (!isLoggedIn()) {
    window.location.href = '/login';
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
      row.innerHTML = `
        <div class="col-3 col-md-2">
          <img src="${product.image_url || '/static/imagens/camiseta-feminina-preta.jpg'}"
               class="img-fluid rounded" alt="${product.name}">
        </div>
        <div class="col-9 col-md-4 text-start">
          <h6 class="text-white mb-1">${product.name}</h6>
          <p class="text-danger mb-0 fw-bold">R$ ${product.price.toFixed(2)}</p>
        </div>
        <div class="col-6 col-md-3 mt-2 mt-md-0">
          <div class="input-group input-group-sm">
            <button class="btn btn-outline-danger" onclick="updateQuantity(${item.id}, ${item.quantity - 1})">-</button>
            <input type="number" class="form-control bg-dark text-white text-center border-danger"
                   value="${item.quantity}" min="1"
                   onchange="updateQuantity(${item.id}, parseInt(this.value))"
                   id="qty-${item.id}">
            <button class="btn btn-outline-danger" onclick="updateQuantity(${item.id}, ${item.quantity + 1})">+</button>
          </div>
        </div>
        <div class="col-4 col-md-2 mt-2 mt-md-0 text-end">
          <p class="text-white mb-0 fw-bold">R$ ${subtotal.toFixed(2)}</p>
        </div>
        <div class="col-2 col-md-1 mt-2 mt-md-0 text-end">
          <button class="btn btn-sm btn-outline-danger" onclick="removeItem(${item.id})">
            <i class="fa fa-trash"></i>
          </button>
        </div>
      `;
      container.appendChild(row);
    });

    document.getElementById('cart-total').textContent = `R$ ${total.toFixed(2)}`;

  } catch (err) {
    loading.style.display = 'none';
    container.innerHTML = `<p class="text-danger">Erro ao carregar carrinho: ${err.message}</p>`;
  }
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
    } else {
      const data = await res.json();
      showToast(data.error || 'Erro ao atualizar', 'danger');
    }
  } catch (err) {
    showToast(err.message, 'danger');
  }
}

async function removeItem(itemId) {
  try {
    const res = await fetch(`${API_BASE}/cart/items/${itemId}`, {
      method: 'DELETE'
    });

    if (res.ok) {
      document.getElementById(`item-${itemId}`)?.remove();
      await loadCart();
    } else {
      showToast('Erro ao remover item', 'danger');
    }
  } catch (err) {
    showToast(err.message, 'danger');
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
  } catch (err) {
    showToast(err.message, 'danger');
    btn.disabled = false;
    btn.textContent = 'FINALIZAR COMPRA';
  }
}