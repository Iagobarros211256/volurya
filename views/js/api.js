// static/js/api.js
const API_BASE = "https://volurya.onrender.com"; // backend real

function getToken() {
    return localStorage.getItem("token");
}

// GET products
export async function getProducts(limit = 10, cursor = null) {
    const token = getToken();
    let url = `${API_BASE}/products?limit=${limit}`;
    if (cursor) url += `&cursor=${cursor}`;

    const res = await fetch(url, {
        headers: { "Authorization": `Bearer ${token}` }
    });
    if (!res.ok) throw new Error("Failed to fetch products");
    return res.json();
}

// CREATE product
export async function createProduct(product) {
    const token = getToken();
    const res = await fetch(`${API_BASE}/products`, {
        method: "POST",
        headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(product)
    });
    if (!res.ok) throw new Error("Failed to create product");
    return res.json();
}

// UPDATE product
export async function updateProduct(id, product) {
    const token = getToken();
    const res = await fetch(`${API_BASE}/products/${id}`, {
        method: "PUT",
        headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(product)
    });
    if (!res.ok) throw new Error("Failed to update product");
    return res.json();
}

// DELETE product
export async function deleteProduct(id) {
    const token = getToken();
    const res = await fetch(`${API_BASE}/products/${id}`, {
        method: "DELETE",
        headers: { "Authorization": `Bearer ${token}` }
    });
    if (!res.ok) throw new Error("Failed to delete product");
    return res.json();
}
