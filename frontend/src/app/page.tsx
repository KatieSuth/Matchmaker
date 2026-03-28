"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api";
import styles from "./page.module.css";

interface HealthResponse {
  status: string;
  message: string;
  timestamp: string;
}

interface Item {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export default function Home() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<"idle" | "checking" | "ok" | "error">("idle");

  const checkHealth = async () => {
    setStatus("checking");
    const { data, error } = await api.get<HealthResponse>("/health");
    if (error) { setStatus("error"); setError(error); }
    else { setStatus("ok"); setHealth(data); setError(null); }
  };

  useEffect(() => { checkHealth(); }, []);

  return (
    <main className={styles.main}>
      <div className={styles.grid_bg} aria-hidden />

      <header className={styles.header}>
        <div className={styles.logo}>
          <span className={styles.logo_bracket}>[</span>
          <span>NEXT</span>
          <span className={styles.logo_sep}>×</span>
          <span className={styles.logo_accent}>GIN</span>
          <span className={styles.logo_bracket}>]</span>
        </div>
        <p className={styles.tagline}>Dockerized Full-Stack Boilerplate</p>
      </header>

      <div className={styles.content}>
        {/* Health Check */}
        <section className={styles.card}>
          <div className={styles.card_header}>
            <span className={styles.card_label}>SYSTEM</span>
            <h2 className={styles.card_title}>API Health Check</h2>
          </div>
          <div className={styles.health_row}>
            <div className={`${styles.status_dot} ${styles[`status_${status}`]}`} />
            <span className={styles.status_text}>
              {status === "idle" && "—"}
              {status === "checking" && "Connecting..."}
              {status === "ok" && `${health?.message} · ${health?.status}`}
              {status === "error" && `Connection failed`}
            </span>
            <button className={styles.btn_ghost} onClick={checkHealth}>
              Ping
            </button>
          </div>
          {health && (
            <div className={styles.mono_row}>
              <span className={styles.muted}>timestamp</span>
              <span className={styles.mono}>{health.timestamp}</span>
            </div>
          )}
          {error && <p className={styles.error}>{error}</p>}
        </section>

        {/* Stack Info */}
        <section className={styles.card_inline}>
          {[
            { label: "Frontend", value: "Next.js 14", sub: "App Router · TypeScript" },
            { label: "Backend", value: "Golang Gin", sub: "REST API · CORS enabled" },
            { label: "Runtime", value: "Docker", sub: "Compose · Networked" },
          ].map(item => (
            <div key={item.label} className={styles.stat}>
              <span className={styles.stat_label}>{item.label}</span>
              <span className={styles.stat_value}>{item.value}</span>
              <span className={styles.stat_sub}>{item.sub}</span>
            </div>
          ))}
        </section>
      </div>
    </main>
  );

  /*
  const [items, setItems] = useState<Item[]>([]);
  const [newItem, setNewItem] = useState({ name: "", description: "" });


  const fetchItems = async () => {
    const { data, error } = await api.get<Item[]>("/api/v1/items");
    if (!error && data) setItems(data);
  };

  const createItem = async () => {
    if (!newItem.name.trim()) return;
    setLoading(true);
    const { data, error } = await api.post<Item>("/api/v1/items", newItem);
    if (!error && data) {
      setItems(prev => [...prev, data]);
      setNewItem({ name: "", description: "" });
    } else {
      setError(error);
    }
    setLoading(false);
  };

  const deleteItem = async (id: number) => {
    const { error } = await api.delete(`/api/v1/items/${id}`);
    if (!error) setItems(prev => prev.filter(i => i.id !== id));
  };

  useEffect(() => { checkHealth(); fetchItems(); }, []);

  return (
    <main className={styles.main}>
      <div className={styles.grid_bg} aria-hidden />

      <header className={styles.header}>
        <div className={styles.logo}>
          <span className={styles.logo_bracket}>[</span>
          <span>NEXT</span>
          <span className={styles.logo_sep}>×</span>
          <span className={styles.logo_accent}>GIN</span>
          <span className={styles.logo_bracket}>]</span>
        </div>
        <p className={styles.tagline}>Dockerized Full-Stack Boilerplate</p>
      </header>

      <div className={styles.content}>
        {/* Health Check }
        <section className={styles.card}>
          <div className={styles.card_header}>
            <span className={styles.card_label}>SYSTEM</span>
            <h2 className={styles.card_title}>API Health Check</h2>
          </div>
          <div className={styles.health_row}>
            <div className={`${styles.status_dot} ${styles[`status_${status}`]}`} />
            <span className={styles.status_text}>
              {status === "idle" && "—"}
              {status === "checking" && "Connecting..."}
              {status === "ok" && `${health?.message} · ${health?.status}`}
              {status === "error" && `Connection failed`}
            </span>
            <button className={styles.btn_ghost} onClick={checkHealth}>
              Ping
            </button>
          </div>
          {health && (
            <div className={styles.mono_row}>
              <span className={styles.muted}>timestamp</span>
              <span className={styles.mono}>{health.timestamp}</span>
            </div>
          )}
          {error && <p className={styles.error}>{error}</p>}
        </section>

        {/* Create Item }
        <section className={styles.card}>
          <div className={styles.card_header}>
            <span className={styles.card_label}>POST /api/v1/items</span>
            <h2 className={styles.card_title}>Create Item</h2>
          </div>
          <div className={styles.form}>
            <input
              className={styles.input}
              placeholder="Item name"
              value={newItem.name}
              onChange={e => setNewItem(p => ({ ...p, name: e.target.value }))}
              onKeyDown={e => e.key === "Enter" && createItem()}
            />
            <input
              className={styles.input}
              placeholder="Description (optional)"
              value={newItem.description}
              onChange={e => setNewItem(p => ({ ...p, description: e.target.value }))}
              onKeyDown={e => e.key === "Enter" && createItem()}
            />
            <button
              className={styles.btn_primary}
              onClick={createItem}
              disabled={loading || !newItem.name.trim()}
            >
              {loading ? "Creating..." : "Create →"}
            </button>
          </div>
        </section>

        {/* Items List }
        <section className={styles.card}>
          <div className={styles.card_header}>
            <span className={styles.card_label}>GET /api/v1/items</span>
            <h2 className={styles.card_title}>Items <span className={styles.count}>{items.length}</span></h2>
          </div>
          {items.length === 0 ? (
            <p className={styles.empty}>No items yet. Create one above.</p>
          ) : (
            <ul className={styles.item_list}>
              {items.map(item => (
                <li key={item.id} className={styles.item}>
                  <div className={styles.item_body}>
                    <span className={styles.item_id}>#{item.id}</span>
                    <div>
                      <p className={styles.item_name}>{item.name}</p>
                      {item.description && <p className={styles.item_desc}>{item.description}</p>}
                    </div>
                  </div>
                  <button className={styles.btn_delete} onClick={() => deleteItem(item.id)}>×</button>
                </li>
              ))}
            </ul>
          )}
        </section>

        {/* Stack Info }
        <section className={styles.card_inline}>
          {[
            { label: "Frontend", value: "Next.js 14", sub: "App Router · TypeScript" },
            { label: "Backend", value: "Golang Gin", sub: "REST API · CORS enabled" },
            { label: "Runtime", value: "Docker", sub: "Compose · Networked" },
          ].map(item => (
            <div key={item.label} className={styles.stat}>
              <span className={styles.stat_label}>{item.label}</span>
              <span className={styles.stat_value}>{item.value}</span>
              <span className={styles.stat_sub}>{item.sub}</span>
            </div>
          ))}
        </section>
      </div>
    </main>
  );
  */
}

