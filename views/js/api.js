// static/js/api.js

const API_BASE = "https://volurya.onrender.com"; // ou localhost se dev

export async function getProducts(limit = 10, cursor = null) {
    let url = `${API_BASE}/products?limit=${limit}`;
    if (cursor) url += `&cursor=${cursor}`;

    const res = await fetch(url, {
        headers: { "Content-Type": "application/json" }
    });
    return res.json();
}

export async function createProduct(product, token) {
    const res = await fetch(`${API_BASE}/products`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + token
        },
        body: JSON.stringify(product)
    });
    return res.json();
}

export async function updateProduct(id, product, token) {
    const res = await fetch(`${API_BASE}/products/${id}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + token
        },
        body: JSON.stringify(product)
    });
    return res.json();
}

export async function deleteProduct(id, token) {
    const res = await fetch(`${API_BASE}/products/${id}`, {
        method: "DELETE",
        headers: {
            "Authorization": "Bearer " + token
        }
    });
    return res;
}
