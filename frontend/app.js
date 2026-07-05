/* ============================================================
   Barber Desk — простой SPA-фронтенд для booking-микросервисов
   Работает через API Gateway (по умолчанию http://localhost:8080/api)
   ============================================================ */

const WEEKDAYS = [
  ["monday", "Понедельник"], ["tuesday", "Вторник"], ["wednesday", "Среда"],
  ["thursday", "Четверг"], ["friday", "Пятница"], ["saturday", "Суббота"], ["sunday", "Воскресенье"]
];
const STATUSES = ["created", "confirmed", "cancelled", "completed"];
const STATUS_LABEL = { created: "создана", confirmed: "подтверждена", cancelled: "отменена", completed: "завершена" };

const store = {
  get apiBase() { return localStorage.getItem("bd_api_base") || "http://localhost:8080/api"; },
  set apiBase(v) { localStorage.setItem("bd_api_base", v); },
  get token() { return localStorage.getItem("bd_token") || ""; },
  set token(v) { v ? localStorage.setItem("bd_token", v) : localStorage.removeItem("bd_token"); },
  get user() { try { return JSON.parse(localStorage.getItem("bd_user")); } catch { return null; } },
  set user(v) { v ? localStorage.setItem("bd_user", JSON.stringify(v)) : localStorage.removeItem("bd_user"); },
  clearSession() { this.token = ""; this.user = null; }
};

/* ---------------- API client ---------------- */
async function api(path, { method = "GET", body, auth = true } = {}) {
  const headers = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (auth && store.token) headers["Authorization"] = "Bearer " + store.token;

  let res;
  try {
    res = await fetch(store.apiBase.replace(/\/$/, "") + path, {
      method, headers, body: body !== undefined ? JSON.stringify(body) : undefined
    });
  } catch (e) {
    throw new Error("Не удалось связаться с сервером (" + store.apiBase + "). Проверьте адрес API и CORS на gateway.");
  }

  let data = null;
  const text = await res.text();
  if (text) { try { data = JSON.parse(text); } catch { data = text; } }

  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || res.statusText || ("HTTP " + res.status);
    throw new Error(typeof msg === "string" ? msg : JSON.stringify(msg));
  }
  return data;
}

/* ---------------- toast ---------------- */
let toastTimer = null;
function toast(msg, type = "ok") {
  const el = document.getElementById("toast");
  el.textContent = msg;
  el.className = "toast " + type;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.add("hidden"), 3200);
}

/* ---------------- small helpers ---------------- */
function weekdayLabel(w) { return (WEEKDAYS.find(([k]) => k === w) || [null, w])[1]; }
function fmtDateTime(v) {
  if (!v) return "—";
  const d = new Date(v);
  if (isNaN(d)) return String(v);
  return d.toLocaleString("ru-RU", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" });
}
function fmtMoney(v) { return (v ?? 0).toLocaleString("ru-RU") + " ₽"; }
function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}
function toLocalInputValue(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d)) return "";
  const pad = n => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
function fromLocalInputValue(v) { return v ? new Date(v).toISOString() : null; }

/* ---------------- auth screen wiring ---------------- */
const authScreen = document.getElementById("auth-screen");
const appShell = document.getElementById("app-shell");

document.querySelectorAll(".tab-btn").forEach(btn => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab-btn").forEach(b => b.classList.remove("active"));
    btn.classList.add("active");
    const tab = btn.dataset.tab;
    document.getElementById("login-form").classList.toggle("hidden", tab !== "login");
    document.getElementById("register-form").classList.toggle("hidden", tab !== "register");
  });
});

document.getElementById("api-base-input").value = store.apiBase;
document.getElementById("save-api-base").addEventListener("click", () => {
  const v = document.getElementById("api-base-input").value.trim();
  if (v) { store.apiBase = v; toast("Адрес API сохранён"); }
});

document.getElementById("login-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const errEl = document.querySelector('.form-error[data-for="login"]');
  errEl.textContent = "";
  const fd = new FormData(e.target);
  try {
    const { token } = await api("/auth/login", { method: "POST", auth: false, body: { email: fd.get("email"), password: fd.get("password") } });
    store.token = token;
    const me = await api("/auth/me");
    store.user = me.user;
    enterApp();
  } catch (err) { errEl.textContent = err.message; }
});

document.getElementById("register-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const errEl = document.querySelector('.form-error[data-for="register"]');
  errEl.textContent = "";
  const fd = new FormData(e.target);
  const body = { name: fd.get("name"), email: fd.get("email"), password: fd.get("password") };
  const adminCode = fd.get("admin_code");
  if (adminCode) body.admin_code = adminCode;
  try {
    await api("/auth/register", { method: "POST", auth: false, body });
    toast("Регистрация прошла успешно, теперь войдите");
    document.querySelector('.tab-btn[data-tab="login"]').click();
    e.target.reset();
  } catch (err) { errEl.textContent = err.message; }
});

document.getElementById("logout-btn").addEventListener("click", () => {
  store.clearSession();
  location.hash = "";
  authScreen.classList.remove("hidden");
  appShell.classList.add("hidden");
});

/* ---------------- navigation / router ---------------- */
const NAV = [
  { group: "Каталог", items: [
    { path: "services", label: "Услуги", roles: ["client", "admin"] },
    { path: "specialists", label: "Специалисты", roles: ["client", "admin"] },
  ]},
  { group: "Клиенту", items: [
    { path: "book", label: "Новая запись", roles: ["client"] },
    { path: "my-appointments", label: "Мои записи", roles: ["client"] },
  ]},
  { group: "Администрирование", items: [
    { path: "appointments", label: "Все записи", roles: ["admin"] },
    { path: "audit", label: "Журнал аудита", roles: ["admin"] },
  ]},
  { group: "", items: [
    { path: "notifications", label: "Уведомления", roles: ["client", "admin"] },
  ]},
];

function renderNav() {
  const role = store.user?.role;
  const nav = document.getElementById("nav");
  nav.innerHTML = "";
  NAV.forEach(section => {
    const items = section.items.filter(i => i.roles.includes(role));
    if (!items.length) return;
    if (section.group) {
      const label = document.createElement("div");
      label.className = "nav-group-label";
      label.textContent = section.group;
      nav.appendChild(label);
    }
    items.forEach(i => {
      const a = document.createElement("a");
      a.href = "#/" + i.path;
      a.textContent = i.label;
      a.dataset.path = i.path;
      nav.appendChild(a);
    });
  });
}

const ROUTES = [
  { re: /^services$/, view: viewServices },
  { re: /^specialists$/, view: viewSpecialists },
  { re: /^specialists\/(\d+)\/schedule$/, view: viewSchedule },
  { re: /^book$/, view: viewBook },
  { re: /^my-appointments$/, view: viewMyAppointments },
  { re: /^appointments$/, view: viewAllAppointments },
  { re: /^notifications$/, view: viewNotifications },
  { re: /^audit$/, view: viewAudit },
];

const TITLES = {
  services: "Услуги", specialists: "Специалисты", book: "Новая запись",
  "my-appointments": "Мои записи", appointments: "Все записи",
  notifications: "Уведомления", audit: "Журнал аудита"
};

async function router() {
  const hash = (location.hash || "#/services").replace(/^#\//, "");
  document.querySelectorAll("#nav a").forEach(a => a.classList.toggle("active", hash.startsWith(a.dataset.path)));
  const top = hash.split("/")[0];
  document.getElementById("page-title").textContent = TITLES[top] || "—";

  const viewEl = document.getElementById("view");
  const match = ROUTES.find(r => r.re.test(hash));
  if (!match) { viewEl.innerHTML = `<div class="empty">Страница не найдена</div>`; return; }
  const params = match.re.exec(hash).slice(1);
  viewEl.innerHTML = `<div class="loading">Загрузка…</div>`;
  try {
    await match.view(viewEl, ...params);
  } catch (err) {
    viewEl.innerHTML = `<div class="panel"><p class="form-error">${esc(err.message)}</p></div>`;
  }
}
window.addEventListener("hashchange", router);

function enterApp() {
  authScreen.classList.add("hidden");
  appShell.classList.remove("hidden");
  document.getElementById("me-name").textContent = store.user?.name || "—";
  document.getElementById("me-role").textContent = store.user?.role || "—";
  document.getElementById("api-base-label").textContent = store.apiBase;
  renderNav();
  router();
}

/* ============================================================
   VIEWS
   ============================================================ */

/* ---- Услуги ---- */
async function viewServices(root) {
  const isAdmin = store.user?.role === "admin";
  const services = await api("/services/");
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Список услуг</h3></div>
      <div id="services-list">${renderServiceCards(services, isAdmin)}</div>
    </div>
    ${isAdmin ? `
    <div class="panel">
      <div class="panel-title"><h3>Добавить услугу</h3></div>
      <form id="service-create-form">
        <div class="form-row">
          <label>Название <input name="title" required minlength="2" maxlength="120"></label>
          <label>Длительность, мин <input name="duration_minutes" type="number" min="1" max="1440" required></label>
          <label>Цена, ₽ <input name="price" type="number" min="0" required></label>
        </div>
        <div class="form-row">
          <label style="grid-column:1/-1">Описание <textarea name="description" required maxlength="2000"></textarea></label>
        </div>
        <label class="check-row"><input type="checkbox" name="is_active" checked> Активна</label>
        <div style="margin-top:14px"><button class="btn primary" type="submit">Создать услугу</button></div>
        <p class="form-error"></p>
      </form>
    </div>
    <div class="panel">
      <div class="panel-title"><h3>Привязать специалиста к услуге</h3></div>
      <form id="link-form">
        <div class="form-row">
          <label>Услуга
            <select name="service_id" required>${services.map(s => `<option value="${s.ID}">${esc(s.title)}</option>`).join("")}</select>
          </label>
          <label>Специалист <span id="link-specialists-slot">загрузка…</span></label>
        </div>
        <button class="btn primary" type="submit">Привязать</button>
        <p class="subtle">Чтобы отвязать — используйте ID связи (specialist_service_id), который приходит в ответе при создании.</p>
        <p class="form-error"></p>
      </form>
      <hr class="divider">
      <form id="unlink-form" style="display:flex;gap:10px;align-items:end">
        <label>ID связи для удаления <input name="id" type="number" min="1" required></label>
        <button class="btn danger" type="submit">Удалить связь</button>
      </form>
    </div>` : ""}
  `;

  if (isAdmin) {
    api("/specialists/").then(specs => {
      document.querySelector('select[name="specialist_id"]')?.remove();
      const slot = document.getElementById("link-specialists-slot");
      slot.outerHTML = `<select name="specialist_id" required>${specs.map(s => `<option value="${s.ID}">${esc(s.name)}</option>`).join("")}</select>`;
    }).catch(() => {});

    document.getElementById("service-create-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const errEl = e.target.querySelector(".form-error");
      const fd = new FormData(e.target);
      try {
        await api("/services/", { method: "POST", body: {
          title: fd.get("title"), description: fd.get("description"),
          duration_minutes: Number(fd.get("duration_minutes")), price: Number(fd.get("price")),
          is_active: fd.get("is_active") === "on"
        }});
        toast("Услуга создана"); router();
      } catch (err) { errEl.textContent = err.message; }
    });

    document.getElementById("link-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const errEl = e.target.querySelector(".form-error");
      const fd = new FormData(e.target);
      try {
        const res = await api("/services/services-specialist", { method: "POST", body: {
          service_id: Number(fd.get("service_id")), specialist_id: Number(fd.get("specialist_id"))
        }});
        toast("Связь создана, её ID: " + (res.ID ?? res.id ?? "?"));
      } catch (err) { errEl.textContent = err.message; }
    });

    document.getElementById("unlink-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("/services/services-specialist/" + fd.get("id"), { method: "DELETE" });
        toast("Связь удалена");
      } catch (err) { toast(err.message, "error"); }
    });

    root.querySelectorAll("[data-edit-service]").forEach(btn => btn.addEventListener("click", () => openServiceEdit(btn.dataset.editService, services)));
    root.querySelectorAll("[data-del-service]").forEach(btn => btn.addEventListener("click", async () => {
      if (!confirm("Удалить услугу?")) return;
      try { await api("/services/" + btn.dataset.delService, { method: "DELETE" }); toast("Услуга удалена"); router(); }
      catch (err) { toast(err.message, "error"); }
    }));
  }
}

function renderServiceCards(services, isAdmin) {
  if (!services?.length) return `<div class="empty">Пока нет услуг</div>`;
  return `<div class="grid-cards">${services.map(s => `
    <div class="card">
      <h4>${esc(s.title)}</h4>
      <p class="desc">${esc(s.description)}</p>
      <div class="meta"><span>${s.duration_minutes} мин</span><span class="price">${fmtMoney(s.price)}</span></div>
      <span class="badge ${s.is_active ? "active" : "inactive"}">${s.is_active ? "активна" : "неактивна"}</span>
      ${isAdmin ? `<div style="display:flex;gap:6px;margin-top:6px">
        <button class="btn ghost small" data-edit-service="${s.ID}">Изменить</button>
        <button class="btn danger small" data-del-service="${s.ID}">Удалить</button>
      </div>` : ""}
    </div>`).join("")}</div>`;
}

function openServiceEdit(id, services) {
  const s = services.find(x => String(x.ID) === String(id));
  if (!s) return;
  const wrap = document.createElement("div");
  wrap.className = "panel";
  wrap.innerHTML = `
    <div class="panel-title"><h3>Изменить услугу «${esc(s.title)}»</h3></div>
    <form id="edit-service-form">
      <div class="form-row">
        <label>Название <input name="title" value="${esc(s.title)}"></label>
        <label>Длительность, мин <input name="duration_minutes" type="number" value="${s.duration_minutes}"></label>
        <label>Цена, ₽ <input name="price" type="number" value="${s.price}"></label>
      </div>
      <div class="form-row"><label style="grid-column:1/-1">Описание <textarea name="description">${esc(s.description)}</textarea></label></div>
      <label class="check-row"><input type="checkbox" name="is_active" ${s.is_active ? "checked" : ""}> Активна</label>
      <div style="margin-top:14px;display:flex;gap:8px">
        <button class="btn primary" type="submit">Сохранить</button>
        <button class="btn ghost" type="button" id="cancel-edit">Отмена</button>
      </div>
      <p class="form-error"></p>
    </form>`;
  document.getElementById("view").prepend(wrap);
  wrap.querySelector("#cancel-edit").addEventListener("click", () => wrap.remove());
  wrap.querySelector("form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = e.target.querySelector(".form-error");
    const fd = new FormData(e.target);
    try {
      await api("/services/" + s.ID, { method: "PATCH", body: {
        title: fd.get("title"), description: fd.get("description"),
        duration_minutes: Number(fd.get("duration_minutes")), price: Number(fd.get("price")),
        is_active: fd.get("is_active") === "on"
      }});
      toast("Услуга обновлена"); router();
    } catch (err) { errEl.textContent = err.message; }
  });
}

/* ---- Специалисты ---- */
async function viewSpecialists(root) {
  const isAdmin = store.user?.role === "admin";
  const specs = await api("/specialists/");
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Специалисты</h3></div>
      <div class="grid-cards">${!specs?.length ? `<div class="empty">Пока нет специалистов</div>` : specs.map(s => `
        <div class="card">
          <h4>${esc(s.name)}</h4>
          <p class="desc">${esc(s.description)}</p>
          <span class="badge ${s.is_active ? "active" : "inactive"}">${s.is_active ? "активен" : "неактивен"}</span>
          <div style="display:flex;gap:6px;margin-top:6px;flex-wrap:wrap">
            <a class="btn ghost small" href="#/specialists/${s.ID}/schedule">Расписание</a>
            ${isAdmin ? `
              <button class="btn ghost small" data-edit-spec="${s.ID}">Изменить</button>
              <button class="btn danger small" data-del-spec="${s.ID}">Удалить</button>` : ""}
          </div>
        </div>`).join("")}</div>
    </div>
    ${isAdmin ? `
    <div class="panel">
      <div class="panel-title"><h3>Добавить специалиста</h3></div>
      <form id="spec-create-form">
        <div class="form-row">
          <label>Имя <input name="name" required minlength="2" maxlength="100"></label>
        </div>
        <div class="form-row"><label style="grid-column:1/-1">Описание <textarea name="description" required maxlength="1000"></textarea></label></div>
        <label class="check-row"><input type="checkbox" name="is_active" checked> Активен</label>
        <div style="margin-top:14px"><button class="btn primary" type="submit">Создать</button></div>
        <p class="form-error"></p>
      </form>
    </div>` : ""}
  `;

  if (isAdmin) {
    document.getElementById("spec-create-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const errEl = e.target.querySelector(".form-error");
      const fd = new FormData(e.target);
      try {
        await api("/specialists/", { method: "POST", body: {
          name: fd.get("name"), description: fd.get("description"), is_active: fd.get("is_active") === "on"
        }});
        toast("Специалист создан"); router();
      } catch (err) { errEl.textContent = err.message; }
    });

    root.querySelectorAll("[data-del-spec]").forEach(btn => btn.addEventListener("click", async () => {
      if (!confirm("Удалить специалиста?")) return;
      try { await api("/specialists/" + btn.dataset.delSpec, { method: "DELETE" }); toast("Специалист удалён"); router(); }
      catch (err) { toast(err.message, "error"); }
    }));
    root.querySelectorAll("[data-edit-spec]").forEach(btn => btn.addEventListener("click", () => openSpecEdit(btn.dataset.editSpec, specs)));
  }
}

function openSpecEdit(id, specs) {
  const s = specs.find(x => String(x.ID) === String(id));
  if (!s) return;
  const wrap = document.createElement("div");
  wrap.className = "panel";
  wrap.innerHTML = `
    <div class="panel-title"><h3>Изменить специалиста «${esc(s.name)}»</h3></div>
    <form id="edit-spec-form">
      <div class="form-row"><label>Имя <input name="name" value="${esc(s.name)}"></label></div>
      <div class="form-row"><label style="grid-column:1/-1">Описание <textarea name="description">${esc(s.description)}</textarea></label></div>
      <label class="check-row"><input type="checkbox" name="is_active" ${s.is_active ? "checked" : ""}> Активен</label>
      <div style="margin-top:14px;display:flex;gap:8px">
        <button class="btn primary" type="submit">Сохранить</button>
        <button class="btn ghost" type="button" id="cancel-edit">Отмена</button>
      </div>
      <p class="form-error"></p>
    </form>`;
  document.getElementById("view").prepend(wrap);
  wrap.querySelector("#cancel-edit").addEventListener("click", () => wrap.remove());
  wrap.querySelector("form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = e.target.querySelector(".form-error");
    const fd = new FormData(e.target);
    try {
      await api("/specialists/" + s.ID, { method: "PATCH", body: {
        name: fd.get("name"), description: fd.get("description"), is_active: fd.get("is_active") === "on"
      }});
      toast("Специалист обновлён"); router();
    } catch (err) { errEl.textContent = err.message; }
  });
}

/* ---- Расписание специалиста ---- */
async function viewSchedule(root, specId) {
  const isAdmin = store.user?.role === "admin";
  let schedule;
  try { schedule = await api("/specialists/" + specId + "/schedule"); }
  catch { schedule = null; }
  const rows = Array.isArray(schedule) ? schedule : (schedule ? [schedule] : []);

  root.innerHTML = `
    <a href="#/specialists" class="subtle">← ко всем специалистам</a>
    <div class="panel" style="margin-top:14px">
      <div class="panel-title"><h3>Расписание специалиста #${specId}</h3></div>
      ${rows.length ? `<table>
        <thead><tr><th>День</th><th>Начало</th><th>Конец</th>${isAdmin ? "<th></th>" : ""}</tr></thead>
        <tbody>${rows.map(r => `
          <tr>
            <td>${weekdayLabel(r.Weekday || r.weekday)}</td>
            <td>${fmtDateTime(r.StartTime || r.start_time)}</td>
            <td>${fmtDateTime(r.EndTime || r.end_time)}</td>
            ${isAdmin ? `<td class="actions">
              <button class="btn danger small" data-del-sched="${r.ID || r.id}">Удалить</button>
            </td>` : ""}
          </tr>`).join("")}</tbody>
      </table>` : `<div class="empty">Расписание пока не задано</div>`}
    </div>
    ${isAdmin ? `
    <div class="panel">
      <div class="panel-title"><h3>Добавить слот расписания</h3></div>
      <form id="sched-form">
        <div class="form-row">
          <label>День недели
            <select name="weekday" required>${WEEKDAYS.map(([k, l]) => `<option value="${k}">${l}</option>`).join("")}</select>
          </label>
          <label>Начало <input name="start_time" type="datetime-local" required></label>
          <label>Конец <input name="end_time" type="datetime-local" required></label>
        </div>
        <button class="btn primary" type="submit">Добавить</button>
        <p class="form-error"></p>
      </form>
    </div>` : ""}
  `;

  if (isAdmin) {
    document.getElementById("sched-form").addEventListener("submit", async (e) => {
      e.preventDefault();
      const errEl = e.target.querySelector(".form-error");
      const fd = new FormData(e.target);
      try {
        await api("/specialists/" + specId + "/schedule", { method: "POST", body: {
          weekday: fd.get("weekday"),
          start_time: fromLocalInputValue(fd.get("start_time")),
          end_time: fromLocalInputValue(fd.get("end_time"))
        }});
        toast("Слот добавлен"); router();
      } catch (err) { errEl.textContent = err.message; }
    });
    root.querySelectorAll("[data-del-sched]").forEach(btn => btn.addEventListener("click", async () => {
      if (!confirm("Удалить слот расписания?")) return;
      try { await api("/specialists/" + specId + "/schedule", { method: "DELETE" }); toast("Слот удалён"); router(); }
      catch (err) { toast(err.message, "error"); }
    }));
  }
}

/* ---- Новая запись (клиент) ---- */
async function viewBook(root) {
  const [services, specs] = await Promise.all([api("/services/"), api("/specialists/")]);
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Записаться на услугу</h3></div>
      <form id="book-form">
        <div class="form-row">
          <label>Специалист
            <select name="specialist_id" required>${specs.map(s => `<option value="${s.ID}">${esc(s.name)}</option>`).join("")}</select>
          </label>
          <label>Услуга
            <select name="service_id" required>${services.map(s => `<option value="${s.ID}">${esc(s.title)} — ${fmtMoney(s.price)}</option>`).join("")}</select>
          </label>
          <label>День недели
            <select name="weekday" required>${WEEKDAYS.map(([k, l]) => `<option value="${k}">${l}</option>`).join("")}</select>
          </label>
        </div>
        <div class="form-row">
          <label>Время начала <input name="start_time" type="time" required></label>
        </div>
        <p class="subtle">Время окончания рассчитается автоматически по длительности выбранной услуги.</p>
        <button class="btn primary" type="submit">Создать запись</button>
        <p class="form-error"></p>
      </form>
    </div>
  `;
  if (!services.length || !specs.length) {
    document.querySelector('#book-form button').disabled = true;
  }
  document.getElementById("book-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = e.target.querySelector(".form-error");
    const fd = new FormData(e.target);
    try {
      await api("/appointments/", { method: "POST", body: {
        specialist_id: Number(fd.get("specialist_id")),
        service_id: Number(fd.get("service_id")),
        weekday: fd.get("weekday"),
        start_time: fd.get("start_time"),
        status: "created"
      }});
      toast("Запись создана"); e.target.reset(); location.hash = "#/my-appointments";
    } catch (err) { errEl.textContent = err.message; }
  });
}

/* ---- Мои записи (клиент) ---- */
async function viewMyAppointments(root) {
  const list = await api("/appointments/my");
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Мои записи</h3></div>
      ${renderAppointmentsTable(list, { canCancel: true })}
    </div>
  `;
  wireAppointmentActions(root);
}

/* ---- Все записи (админ) ---- */
async function viewAllAppointments(root) {
  const list = await api("/appointments/all");
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Все записи</h3></div>
      ${renderAppointmentsTable(list, { canChangeStatus: true })}
    </div>
  `;
  wireAppointmentActions(root);
}

function renderAppointmentsTable(list, { canCancel, canChangeStatus } = {}) {
  if (!list?.length) return `<div class="empty">Записей нет</div>`;
  return `<table>
    <thead><tr>
      <th>№</th><th>Клиент</th><th>Специалист</th><th>Услуга</th><th>День</th><th>Начало</th><th>Конец</th><th>Статус</th><th></th>
    </tr></thead>
    <tbody>${list.map(a => `
      <tr>
        <td>${a.ID}</td>
        <td>${a.client_id}</td>
        <td>${a.specialist_id}</td>
        <td>${a.service_id}</td>
        <td>${weekdayLabel(a.weekday)}</td>
        <td>${fmtDateTime(a.start_time)}</td>
        <td>${fmtDateTime(a.end_time)}</td>
        <td><span class="badge ${a.status}">${STATUS_LABEL[a.status] || a.status}</span></td>
        <td class="actions">
          ${canChangeStatus ? `<select data-status-for="${a.ID}">${STATUSES.map(s => `<option value="${s}" ${s === a.status ? "selected" : ""}>${STATUS_LABEL[s]}</option>`).join("")}</select>
            <button class="btn ghost small" data-save-status="${a.ID}">Сохранить</button>` : ""}
          ${canCancel ? `<button class="btn danger small" data-del-appt="${a.ID}">Отменить</button>` : ""}
        </td>
      </tr>`).join("")}</tbody>
  </table>`;
}

function wireAppointmentActions(root) {
  root.querySelectorAll("[data-save-status]").forEach(btn => btn.addEventListener("click", async () => {
    const id = btn.dataset.saveStatus;
    const status = root.querySelector(`[data-status-for="${id}"]`).value;
    try { await api("/appointments/" + id + "/status", { method: "PATCH", body: { status } }); toast("Статус обновлён"); router(); }
    catch (err) { toast(err.message, "error"); }
  }));
  root.querySelectorAll("[data-del-appt]").forEach(btn => btn.addEventListener("click", async () => {
    if (!confirm("Отменить/удалить запись?")) return;
    try { await api("/appointments/" + btn.dataset.delAppt, { method: "DELETE" }); toast("Запись удалена"); router(); }
    catch (err) { toast(err.message, "error"); }
  }));
}

/* ---- Уведомления ---- */
async function viewNotifications(root) {
  const res = await api("/notifications/my");
  const list = res?.data || res || [];
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Уведомления</h3></div>
      ${!list.length ? `<div class="empty">Уведомлений нет</div>` : `
      <table>
        <thead><tr><th>Тип</th><th>Заголовок</th><th>Сообщение</th><th>Дата</th><th>Статус</th><th></th></tr></thead>
        <tbody>${list.map(n => `
          <tr>
            <td>${esc(n.type)}</td>
            <td>${esc(n.title)}</td>
            <td>${esc(n.message)}</td>
            <td>${fmtDateTime(n.CreatedAt || n.created_at)}</td>
            <td><span class="badge ${n.is_read ? "completed" : "created"}">${n.is_read ? "прочитано" : "новое"}</span></td>
            <td>${!n.is_read ? `<button class="btn ghost small" data-read="${n.ID || n.id}">Отметить прочитанным</button>` : ""}</td>
          </tr>`).join("")}</tbody>
      </table>`}
    </div>
  `;
  root.querySelectorAll("[data-read]").forEach(btn => btn.addEventListener("click", async () => {
    try { await api("/notifications/" + btn.dataset.read + "/read", { method: "PATCH" }); toast("Отмечено прочитанным"); router(); }
    catch (err) { toast(err.message, "error"); }
  }));
}

/* ---- Журнал аудита (админ) ---- */
async function viewAudit(root) {
  const res = await api("/audit/events");
  const list = res?.data || res || [];
  root.innerHTML = `
    <div class="panel">
      <div class="panel-title"><h3>Журнал аудита</h3></div>
      ${!list.length ? `<div class="empty">Событий нет</div>` : `
      <table>
        <thead><tr><th>№</th><th>Событие</th><th>Сервис</th><th>Сущность</th><th>Actor</th><th>Дата</th><th></th></tr></thead>
        <tbody>${list.map(a => `
          <tr>
            <td>${a.ID || a.id}</td>
            <td>${esc(a.event_type)}</td>
            <td>${esc(a.source_service)}</td>
            <td>${esc(a.entity_type)} #${a.entity_id}</td>
            <td>${a.actor_id}</td>
            <td>${fmtDateTime(a.CreatedAt || a.created_at)}</td>
            <td><button class="btn ghost small" data-audit-details='${esc(JSON.stringify(a))}'>Детали</button></td>
          </tr>`).join("")}</tbody>
      </table>`}
    </div>
    <div id="audit-detail-slot"></div>
  `;
  root.querySelectorAll("[data-audit-details]").forEach(btn => btn.addEventListener("click", () => {
    const a = JSON.parse(btn.dataset.auditDetails);
    document.getElementById("audit-detail-slot").innerHTML = `
      <div class="panel"><div class="panel-title"><h3>Событие #${a.ID || a.id}</h3></div>
      <pre style="white-space:pre-wrap;font-size:12.5px;background:var(--cream);padding:12px;border-radius:8px;">${esc(a.payload || "—")}</pre>
      </div>`;
  }));
}

/* ---------------- boot ---------------- */
(async function boot() {
  if (store.token && store.user) {
    try {
      const me = await api("/auth/me");
      store.user = me.user;
      enterApp();
      return;
    } catch { store.clearSession(); }
  }
  authScreen.classList.remove("hidden");
  appShell.classList.add("hidden");
})();
