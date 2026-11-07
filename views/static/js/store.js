document.addEventListener('DOMContentLoaded', async () => {
  const container = document.getElementById('store-items');
  const salesMode = (container.dataset.salesMode || '').toLowerCase();
  const salesWhatsAppURL = container.dataset.salesWhatsappUrl || '';
  const cartAvailable = container.dataset.cartAvailable === 'true';
  const whatsappSalesEnabled =
    salesMode === 'whatsapp' &&
    /^https:\/\/wa\.me\/55[0-9]{10,11}$/.test(salesWhatsAppURL);
  const loading = document.getElementById('loading');
  const noProducts = document.getElementById('no-products');

  // skeletons já visíveis por padrão — não precisa forçar display
  container.innerHTML = '';
  noProducts.classList.add('d-none');

  try {
    const res = await fetch('/api/products?limit=12');
    if (!res.ok) throw new Error('Erro ao carregar produtos');

    const { data } = await res.json();

    if (data.length === 1) {
      container.classList.add('store-single-product');
    }

    if (data.length === 0) {
      noProducts.classList.remove('d-none');
      return;
    }

    data.forEach(product => {
      const col = document.createElement('div');
      col.className = 'col';

      // Card
      const card = document.createElement('div');
      card.className = 'card bg-dark text-light border-danger h-100 shadow-lg store-product-card';

      // Imagem — usa image_url real se disponível e segura
      const img = document.createElement('img');
      img.src = isValidImageUrl(product.image_url)
        ? product.image_url
        : '/static/imagens/camiseta-feminina-preta.jpg';
      img.className = 'card-img-top store-product-image';
      img.alt = product.name; // propriedade — não interpreta HTML

      // Card body
      const cardBody = document.createElement('div');
      cardBody.className = 'card-body d-flex flex-column text-start';

      // Nome
      const title = document.createElement('h5');
      title.className = 'card-title text-danger';
      title.textContent = product.name; // seguro

      // Descrição
      const desc = document.createElement('p');
      desc.className = 'card-text mb-2';
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
      stockP.className = 'card-text text-light opacity-75 small';
      const outOfStock = product.stock === 0;
      const isPreorder = product.sales_type === 'preorder';
      stockP.textContent = isPreorder
        ? 'Pré-venda — disponibilidade a confirmar'
        : (outOfStock ? 'Esgotado' : `Estoque: ${product.stock}`);

      // Venda assistida: tamanho e preço vêm da API, nunca do nome do produto.
      if (whatsappSalesEnabled) {
        const link = document.createElement('a');
        link.className = 'btn btn-success w-100';
        link.target = '_blank';
        link.rel = 'noopener noreferrer';

        if (isPreorder) {
          const variants = Array.isArray(product.variants)
            ? product.variants.filter(v =>
                Number.isInteger(v.id) &&
                typeof v.label === 'string' &&
                Number.isSafeInteger(v.price_cents) &&
                v.price_cents >= 0
              )
            : [];

          const formatPrice = cents =>
            (cents / 100).toLocaleString('pt-BR', {
              style: 'currency',
              currency: 'BRL'
            });

          const label = document.createElement('label');
          label.className = 'form-label';
          label.textContent = 'Tamanho';

          const select = document.createElement('select');
          select.id = `product-size-${product.id_product}`;
          select.className = 'form-select bg-dark text-white mb-3';
          label.htmlFor = select.id;

          const placeholder = document.createElement('option');
          placeholder.value = '';
          placeholder.textContent = 'Escolha seu tamanho';
          select.appendChild(placeholder);

          variants.forEach(variant => {
            const option = document.createElement('option');
            option.value = String(variant.id);
            option.textContent =
              `${variant.label} — ${formatPrice(variant.price_cents)}`;
            select.appendChild(option);
          });

          const initialPrice = variants.length
            ? `A partir de ${formatPrice(
                Math.min(...variants.map(v => v.price_cents))
              )}`
            : 'Preço indisponível';

          priceP.textContent = initialPrice;
          link.textContent = variants.length
            ? 'Selecione um tamanho'
            : 'Pré-venda indisponível';
          link.classList.add('disabled');
          link.setAttribute('aria-disabled', 'true');
          select.disabled = variants.length === 0;

          select.addEventListener('change', () => {
            const selected = variants.find(
              variant => String(variant.id) === select.value
            );

            if (!selected) {
              priceP.textContent = initialPrice;
              link.textContent = 'Selecione um tamanho';
              link.removeAttribute('href');
              link.classList.add('disabled');
              link.setAttribute('aria-disabled', 'true');
              return;
            }

            const price = formatPrice(selected.price_cents);
            priceP.textContent = price;

            const message = [
              'Olá, Heitor! Vim pelo site da Volurya.',
              `Tenho interesse no produto: ${product.name}.`,
              `Tamanho escolhido: ${selected.label}.`,
              `Preço anunciado: ${price}.`,
              'Gostaria de confirmar disponibilidade, pagamento e entrega.'
            ].join('\n');

            link.href =
              `${salesWhatsAppURL}?text=${encodeURIComponent(message)}`;
            link.textContent = 'Consultar pelo WhatsApp';
            link.classList.remove('disabled');
            link.removeAttribute('aria-disabled');
          });

          priceWrapper.appendChild(priceP);
          priceWrapper.appendChild(stockP);
          priceWrapper.appendChild(label);
          priceWrapper.appendChild(select);
        } else {
          const message = [
            'Olá, Heitor! Vim pelo site da Volurya.',
            `Tenho interesse no produto: ${product.name}.`,
            `Preço anunciado: ${priceP.textContent}.`,
            'Gostaria de confirmar disponibilidade, pagamento e entrega.'
          ].join('\n');

          link.textContent = outOfStock
            ? 'Esgotado'
            : 'Consultar pelo WhatsApp';

          if (outOfStock) {
            link.classList.add('disabled');
            link.setAttribute('aria-disabled', 'true');
          } else {
            link.href =
              `${salesWhatsAppURL}?text=${encodeURIComponent(message)}`;
          }

          priceWrapper.appendChild(priceP);
          priceWrapper.appendChild(stockP);
        }

        priceWrapper.appendChild(link);
        cardBody.appendChild(title);
        cardBody.appendChild(desc);
        cardBody.appendChild(priceWrapper);
        card.appendChild(img);
        card.appendChild(cardBody);
        col.appendChild(card);
        container.appendChild(col);
        return;
      }

      // Configuração incompleta: não oferecer checkout indisponível.
      if (!cartAvailable) {
        const unavailable = document.createElement('button');
        unavailable.type = 'button';
        unavailable.className = 'btn btn-outline-secondary w-100';
        unavailable.textContent = 'Vendas temporariamente indisponíveis';
        unavailable.disabled = true;

        priceWrapper.appendChild(priceP);
        priceWrapper.appendChild(stockP);
        priceWrapper.appendChild(unavailable);

        cardBody.appendChild(title);
        cardBody.appendChild(desc);
        cardBody.appendChild(priceWrapper);

        card.appendChild(img);
        card.appendChild(cardBody);
        col.appendChild(card);
        container.appendChild(col);
        return;
      }

      // Modo original: carrinho da loja.
      const btn = document.createElement('button');
      btn.className = 'btn btn-outline-danger w-100';
      btn.textContent = isPreorder
        ? 'Pré-venda indisponível neste modo'
        : (outOfStock ? 'Esgotado' : 'Adicionar ao Carrinho');
      btn.disabled = isPreorder || outOfStock;

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
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ product_id: product.id_product, quantity: 1 })
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