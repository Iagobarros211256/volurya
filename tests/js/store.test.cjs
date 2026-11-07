const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

const source = fs.readFileSync('views/static/js/store.js', 'utf8');

class Element {
  constructor(tag = 'div') {
    this.tagName = tag;
    this.children = [];
    this.dataset = {};
    const classes = new Set();
    this.classList = {
      add(...names) { names.forEach(name => classes.add(name)); },
      remove(...names) { names.forEach(name => classes.delete(name)); },
      contains(name) { return classes.has(name); }
    };
    this.listeners = {};
    this.style = {};
    this.textContent = '';
    this.disabled = false;
  }

  appendChild(child) {
    this.children.push(child);
    return child;
  }

  setAttribute(name, value) {
    this[name] = value;
  }

  addEventListener(event, callback) {
    this.listeners[event] = callback;
  }

  removeAttribute(name) {
    delete this[name];
  }

  set innerHTML(value) {
    this.children = [];
  }

  get innerHTML() {
    return '';
  }
}

function findAll(element, predicate) {
  const result = [];

  if (predicate(element)) result.push(element);

  for (const child of element.children) {
    result.push(...findAll(child, predicate));
  }

  return result;
}

async function renderStore({ mode, url, cartAvailable, stock = 5, product = null }) {
  const container = new Element();
  container.dataset = {
    salesMode: mode,
    salesWhatsappUrl: url,
    cartAvailable: String(cartAvailable)
  };

  const loading = new Element();
  const noProducts = new Element();

  let onReady;

  const document = {
    addEventListener(event, callback) {
      if (event === 'DOMContentLoaded') onReady = callback;
    },
    getElementById(id) {
      return {
        'store-items': container,
        loading,
        'no-products': noProducts
      }[id];
    },
    createElement(tag) {
      return new Element(tag);
    }
  };

  const context = {
    document,
    fetch: async () => ({
      ok: true,
      json: async () => ({
        data: [product ?? {
          id_product: 1,
          name: 'Camiseta Volurya',
          description: 'Produto oficial',
          price: 79.90,
          stock,
          image_url: ''
        }]
      })
    }),
    console,
    encodeURIComponent
  };

  vm.runInNewContext(source, context);
  assert.equal(typeof onReady, 'function');

  await onReady();

  return container;
}

test('WhatsApp válido mostra link de contato', async () => {
  const container = await renderStore({
    mode: 'whatsapp',
    url: 'https://wa.me/5585999999999',
    cartAvailable: false
  });

  const links = findAll(container, el => el.tagName === 'a');

  assert.equal(links.length, 1);
  assert.equal(links[0].textContent, 'Consultar pelo WhatsApp');
  assert.match(links[0].href, /^https:\/\/wa\.me\/5585999999999\?text=/);
  assert.match(
    decodeURIComponent(links[0].href),
    /Camiseta Volurya/
  );
});

test('WhatsApp inválido bloqueia vendas', async () => {
  const container = await renderStore({
    mode: 'whatsapp',
    url: '',
    cartAvailable: false
  });

  const buttons = findAll(
    container,
    el => el.tagName === 'button'
  );

  assert.equal(buttons.length, 1);
  assert.equal(buttons[0].disabled, true);
  assert.equal(
    buttons[0].textContent,
    'Vendas temporariamente indisponíveis'
  );
});

test('Pagamentos desabilitados não mostram carrinho', async () => {
  const container = await renderStore({
    mode: '',
    url: '',
    cartAvailable: false
  });

  const buttons = findAll(
    container,
    el => el.tagName === 'button'
  );

  assert.equal(buttons.length, 1);
  assert.equal(buttons[0].disabled, true);
  assert.equal(
    buttons[0].textContent,
    'Vendas temporariamente indisponíveis'
  );
});

test('Carrinho disponível mostra botão tradicional', async () => {
  const container = await renderStore({
    mode: '',
    url: '',
    cartAvailable: true
  });

  const buttons = findAll(
    container,
    el => el.tagName === 'button'
  );

  assert.equal(buttons.length, 1);
  assert.equal(buttons[0].textContent, 'Adicionar ao Carrinho');
  assert.equal(buttons[0].disabled, false);
});

test('Produto esgotado não oferece contato ativo', async () => {
  const container = await renderStore({
    mode: 'whatsapp',
    url: 'https://wa.me/5585999999999',
    cartAvailable: false,
    stock: 0
  });

  const links = findAll(container, el => el.tagName === 'a');

  assert.equal(links.length, 1);
  assert.equal(links[0].textContent, 'Esgotado');
  assert.equal(links[0].href, undefined);
});


function makePreorder(variants = [
  ['P', 5000], ['M', 5000], ['G', 5000],
  ['GG', 5000], ['XG', 5500], ['XGG', 5500]
].map(([label, price_cents], index) => ({
  id: index + 1,
  product_id: 1,
  label,
  price_cents,
  display_order: index + 1
}))) {
  return {
    id_product: 1,
    name: 'Camiseta Volurya — Pré-venda',
    description: 'Algodão premium, estampa frente e costas',
    price: 50,
    stock: 0,
    sales_type: 'preorder',
    image_url: '',
    variants
  };
}

test('Pré-venda exige seleção de tamanho e testa os seis preços', async () => {
  const container = await renderStore({
    mode: 'whatsapp',
    url: 'https://wa.me/558581842950',
    cartAvailable: false,
    product: makePreorder()
  });

  const [select] = findAll(container, el => el.tagName === 'select');
  const [link] = findAll(container, el => el.tagName === 'a');
  const [price] = findAll(
    container,
    el => el.tagName === 'p' &&
      el.className?.includes('text-success')
  );

  assert.ok(select);
  assert.ok(link);
  assert.ok(price);

  // Não permitir contato antes da escolha.
  assert.equal(link.href, undefined);
  assert.equal(link['aria-disabled'], 'true');
  assert.equal(link.classList.contains('disabled'), true);
  assert.match(price.textContent, /50,00/);

  const expected = [
    ['P', 5000],
    ['M', 5000],
    ['G', 5000],
    ['GG', 5000],
    ['XG', 5500],
    ['XGG', 5500]
  ];

  for (const [index, [size, cents]] of expected.entries()) {
    select.value = String(index + 1);
    select.listeners.change();

    assert.match(
      price.textContent,
      new RegExp(`${(cents / 100).toFixed(2).replace('.', ',')}`)
    );

    assert.equal(link.textContent, 'Consultar pelo WhatsApp');
    assert.equal(link.classList.contains('disabled'), false);
    assert.equal(link['aria-disabled'], undefined);

    const url = new URL(link.href);

    assert.equal(url.origin, 'https://wa.me');
    assert.equal(url.pathname, '/558581842950');

    const message = url.searchParams.get('text');

    assert.ok(message.includes('Camiseta Volurya'));
    assert.ok(message.includes('\n'));
    assert.equal(message.includes('\\n'), false);
    assert.ok(message.includes(`Tamanho escolhido: ${size}.`));
    assert.ok(
      message.includes(
        `Preço anunciado: R$ ${(cents / 100).toFixed(2).replace('.', ',')}`
      ) ||
      message.includes(
        `Preço anunciado: R$ ${ (cents / 100).toFixed(2).replace('.', ',')}`
      )
    );
  }

  // Remover seleção precisa bloquear novamente.
  select.value = '';
  select.listeners.change();

  assert.equal(link.href, undefined);
  assert.equal(link['aria-disabled'], 'true');
  assert.equal(link.classList.contains('disabled'), true);
});

test('Pré-venda sem variantes não libera WhatsApp', async () => {
  const container = await renderStore({
    mode: 'whatsapp',
    url: 'https://wa.me/558581842950',
    cartAvailable: false,
    product: makePreorder([])
  });

  const [select] = findAll(container, el => el.tagName === 'select');
  const [link] = findAll(container, el => el.tagName === 'a');

  assert.equal(select.disabled, true);
  assert.equal(link.href, undefined);
  assert.equal(link.textContent, 'Pré-venda indisponível');
});

test('Pré-venda não oferece botão ativo de carrinho', async () => {
  const container = await renderStore({
    mode: '',
    url: '',
    cartAvailable: true,
    product: makePreorder()
  });

  const [button] = findAll(container, el => el.tagName === 'button');

  assert.equal(button.disabled, true);
  assert.equal(button.textContent, 'Pré-venda indisponível neste modo');
});
