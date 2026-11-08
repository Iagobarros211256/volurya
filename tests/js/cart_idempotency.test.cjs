const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

const source = fs.readFileSync('views/static/js/cart.js', 'utf8');

function setup() {
  const values = new Map();
  let generated = 0;

  const sessionStorage = {
    getItem: key => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value)
  };

  const window = {
    sessionStorage,
    crypto: {
      randomUUID() {
        generated++;
        return `${String(generated).padStart(8, '0')}-0000-4000-8000-000000000000`;
      }
    }
  };

  const context = {
    window,
    sessionStorage,
    document: { addEventListener() {} }
  };

  vm.runInNewContext(source, context);

  return {
    key: context.checkoutIdempotencyKey,
    window,
    generated: () => generated
  };
}

test('mesmo carrinho reutiliza chave', () => {
  const env = setup();
  const items = [
    { product_id: 2, quantity: 1 },
    { product_id: 1, quantity: 3 }
  ];

  const first = env.key(items, 'CLIENTE@EXAMPLE.INVALID');
  const second = env.key([...items].reverse(), 'cliente@example.invalid');

  assert.equal(first, second);
  assert.equal(env.generated(), 1);
});

test('alteração do carrinho gera nova chave', () => {
  const env = setup();
  const first = env.key([{ product_id: 1, quantity: 1 }], 'a@example.invalid');
  const second = env.key([{ product_id: 1, quantity: 2 }], 'a@example.invalid');

  assert.notEqual(first, second);
  assert.equal(env.generated(), 2);
});

test('alteração do email gera nova chave', () => {
  const env = setup();
  const items = [{ product_id: 1, quantity: 1 }];

  const first = env.key(items, 'a@example.invalid');
  const second = env.key(items, 'b@example.invalid');

  assert.notEqual(first, second);
});

test('sem gerador seguro, falha sem criar chave', () => {
  const env = setup();
  env.window.crypto = undefined;

  assert.throws(
    () => env.key([{ product_id: 1, quantity: 1 }], 'a@example.invalid'),
    /Gerador seguro/
  );
});

test('checkout inclui chave idempotente no corpo', () => {
  assert.match(source, /idempotency_key:\s*checkoutKey/);
});
