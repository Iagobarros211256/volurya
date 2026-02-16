// static/js/main.js
import { getProducts, createProduct, updateProduct, deleteProduct } from './api.js';

let nextCursor = null;
const token = "INSIRA_SEU_TOKEN_AQUI";

const productTable = document.getElementById('productTable');
const loadMoreBtn = document.getElementById('loadMoreBtn');

async function loadProducts() {
    const data = await getProducts(5, nextCursor);
    data.data.forEach(p => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${p.name}</td>
            <td>${p.price}</td>
            <td>${p.stock}</td>
            <td>
                <button class="btn btn-sm btn-warning editBtn" data-id="${p.id}">Edit</button>
                <button class="btn btn-sm btn-danger deleteBtn" data-id="${p.id}">Delete</button>
            </td>
        `;
        productTable.appendChild(tr);
    });

    nextCursor = data.pagination.has_more ? data.pagination.next_cursor : null;
    loadMoreBtn.disabled = !nextCursor;

    attachRowEvents();
}

function attachRowEvents() {
    document.querySelectorAll('.editBtn').forEach(btn => {
        btn.onclick = () => openModal(btn.dataset.id);
    });
    document.querySelectorAll('.deleteBtn').forEach(btn => {
        btn.onclick = async () => {
            await deleteProduct(btn.dataset.id, token);
            productTable.innerHTML = '';
            nextCursor = null;
            loadProducts();
        };
    });
}

function openModal(id = null) {
    const modal = new bootstrap.Modal(document.getElementById('productModal'));
    if (id) {
        // buscar dados existentes na tabela
        const row = document.querySelector(`[data-id="${id}"]`).closest('tr');
        document.getElementById('productId').value = id;
        document.getElementById('productName').value = row.cells[0].innerText;
        document.getElementById('productDescription').value = row.cells[1].innerText;
        document.getElementById('productPrice').value = row.cells[2].innerText;
        document.getElementById('productStock').value = row.cells[3].innerText;
    } else {
        document.getElementById('productId').value = '';
    }
    modal.show();
}

document.getElementById('addProductBtn').onclick = () => openModal();

document.getElementById('saveProductBtn').onclick = async () => {
    const id = document.getElementById('productId').value;
    const product = {
        name: document.getElementById('productName').value,
        description: document.getElementById('productDescription').value,
        price: parseFloat(document.getElementById('productPrice').value),
        stock: parseInt(document.getElementById('productStock').value)
    };

    if (id) await updateProduct(id, product, token);
    else await createProduct(product, token);

    productTable.innerHTML = '';
    nextCursor = null;
    loadProducts();

    bootstrap.Modal.getInstance(document.getElementById('productModal')).hide();
}

// inicializa
loadProducts();

loadMoreBtn.onclick = loadProducts;
