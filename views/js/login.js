
// static/js/login.js
const API_BASE = "https://volurya.onrender.com"; // backend
const loginBtn = document.getElementById("loginBtn");
const toastEl = document.getElementById("loginToast");
const toastMsg = document.getElementById("toastMessage");
const bsToast = new bootstrap.Toast(toastEl);

loginBtn.onclick = async () => {
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    try {
        const res = await fetch(`${API_BASE}/auth/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, password })
        });

        if (!res.ok) throw new Error("Invalid credentials");

        const data = await res.json();
        localStorage.setItem("token", data.access_token);
        toastMsg.textContent = "Login successful!";
        bsToast.show();

        setTimeout(() => {
            window.location.href = "/index.html"; // redireciona para o painel
        }, 1000);
    } catch (err) {
        toastMsg.textContent = err.message;
        bsToast.show();
    }
};
