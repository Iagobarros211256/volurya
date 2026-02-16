// static/js/main.js
import { getProducts, createProduct, updateProduct, deleteProduct } from './api.js';

const productTable = document.getElementById("productTable"); // div ou table
const actionToastEl = document.getElementById("actionToast");
const actionToastMsg = document.getElementById("actionToastMsg");
const actionToast = new bootstrap.Toast(actionToastEl);

function showToast(message) {
    actionToastMsg.textContent = message;
    actionToast.show();
}

// Carregar produtos
let nextCursor = null;
async function loadProducts() {
    try {
        const data = await getProducts(10, nextCursor);
        data.products.forEach(product => {
            const row = document.createElement("tr");
            row.innerHTML = `
                <td>${product.id}</td>
                <td>${product.name}</td>
                <td>${product.description}</td>
                <td>${product.price}</td>
                <td>${product.stock}</td>
                <td>
                    <button class="btn btn-sm btn-warning editBtn" data-id="${product.id}">Edit</button>
                    <button class="btn btn-sm btn-danger deleteBtn" data-id="${product.id}">Delete</button>
                </td>
            `;
            productTable.appendChild(row);
        });

        nextCursor = data.nextCursor;
        attachEventListeners();
    } catch (err) {
        showToast(err.message);
    }
}

// CRUD handlers
function attachEventListeners() {
    document.querySelectorAll('.deleteBtn').forEach(btn => {
        btn.onclick = async () => {
            try {
                await deleteProduct(btn.dataset.id);
                showToast("Product deleted!");
                productTable.innerHTML = '';
                nextCursor = null;
                loadProducts();
            } catch (err) {
                showToast(err.message);
            }
        };
    });

    document.querySelectorAll('.editBtn').forEach(btn => {
        btn.onclick = async () => {
            const id = btn.dataset.id;
            const name = prompt("New name:");
            const description = prompt("New description:");
            const price = parseFloat(prompt("New price:"));
            const stock = parseInt(prompt("New stock:"));

            try {
                await updateProduct(id, { name, description, price, stock });
                showToast("Product updated!");
                productTable.innerHTML = '';
                nextCursor = null;
                loadProducts();
            } catch (err) {
                showToast(err.message);
            }
        };
    });
}

// Formulário de criação
const createForm = document.getElementById("createForm");
if (createForm) {
    createForm.onsubmit = async (e) => {
        e.preventDefault();
        const name = createForm.elements["name"].value;
        const description = createForm.elements["description"].value;
        const price = parseFloat(createForm.elements["price"].value);
        const stock = parseInt(createForm.elements["stock"].value);

        try {
            await createProduct({ name, description, price, stock });
            showToast("Product created!");
            createForm.reset();
            productTable.innerHTML = '';
            nextCursor = null;
            loadProducts();
        } catch (err) {
            showToast(err.message);
        }
    };
}

// Inicialização
loadProducts();
