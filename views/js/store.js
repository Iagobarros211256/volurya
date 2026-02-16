// js/store.js

async function loadProducts() {
  try {
    const response = await fetch('https://volurya.onrender.com/products?limit=10');
    const data = await response.json();

    const container = document.getElementById('store-items');
    container.innerHTML = ''; // limpa antes de adicionar

    data.data.forEach(product => {
      const card = document.createElement('div');
      card.className = 'card m-2';
      card.style.width = '200px';
      card.innerHTML = `
        <img src="/static/imagens/camiseta-feminina-preta.jpg" class="card-img-top">
        <div class="card-body">
          <h5 class="card-title">${product.name}</h5>
          <p class="card-text">${product.description}</p>
          <p>R$ ${product.price.toFixed(2)}</p>
          <p>Stock: ${product.stock}</p>
        </div>
      `;
      container.appendChild(card);
    });

  } catch (error) {
    console.error('Erro ao carregar produtos:', error);
  }
}

// Carregar produtos quando a página abrir
document.addEventListener('DOMContentLoaded', loadProducts);


// Atualizar produto
async function updateProduct(id, name, description, price, stock) {
  await authFetch(`https://volurya.onrender.com/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, description, price, stock })
  });
}

// Deletar produto
async function deleteProduct(id) {
  await authFetch(`https://volurya.onrender.com/products/${id}`, {
    method: 'DELETE'
  });
  loadProducts(); // Atualiza lista
}
