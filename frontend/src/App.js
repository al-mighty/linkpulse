import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from "react";
import { Link2, BarChart3, Copy, Check, ExternalLink } from "lucide-react";
const API = "/api";
export default function App() {
    const [url, setUrl] = useState("");
    const [links, setLinks] = useState([]);
    const [stats, setStats] = useState(null);
    const [shortUrl, setShortUrl] = useState("");
    const [copied, setCopied] = useState(false);
    const [loading, setLoading] = useState(false);
    async function loadLinks() {
        const resp = await fetch(`${API}/links`);
        setLinks(await resp.json());
    }
    async function handleSubmit(e) {
        e.preventDefault();
        if (!url.trim())
            return;
        setLoading(true);
        try {
            const resp = await fetch(`${API}/links`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ url }),
            });
            const data = await resp.json();
            setShortUrl(data.short_url);
            setUrl("");
            loadLinks();
        }
        finally {
            setLoading(false);
        }
    }
    async function handleCopy() {
        await navigator.clipboard.writeText(shortUrl);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    }
    async function showStats(code) {
        const resp = await fetch(`${API}/links/${code}/stats`);
        setStats(await resp.json());
    }
    // Load links on first render
    if (links.length === 0 && !loading) {
        loadLinks();
    }
    return (_jsxs("div", { className: "min-h-screen", children: [_jsx("header", { className: "bg-pulse-primary text-white px-6 py-4", children: _jsxs("div", { className: "max-w-3xl mx-auto flex items-center gap-3", children: [_jsx(Link2, { className: "w-6 h-6" }), _jsxs("div", { children: [_jsx("h1", { className: "text-xl font-bold", children: "LinkPulse" }), _jsx("p", { className: "text-xs text-indigo-200 font-mono", children: "URL Shortener + Analytics" })] })] }) }), _jsxs("main", { className: "max-w-3xl mx-auto p-6 space-y-6", children: [_jsxs("form", { onSubmit: handleSubmit, className: "flex gap-3", children: [_jsx("input", { value: url, onChange: (e) => setUrl(e.target.value), placeholder: "https://example.com/very-long-url...", className: "flex-1 px-4 py-3 rounded-lg border border-gray-300 text-sm focus:outline-none focus:border-pulse-primary focus:ring-1 focus:ring-pulse-primary" }), _jsx("button", { disabled: loading || !url.trim(), className: "px-6 py-3 bg-pulse-primary text-white rounded-lg font-medium text-sm hover:bg-pulse-dark disabled:opacity-40 transition", children: loading ? "..." : "Сократить" })] }), shortUrl && (_jsxs("div", { className: "flex items-center gap-3 bg-pulse-light border border-indigo-200 rounded-lg px-4 py-3", children: [_jsx("a", { href: shortUrl, target: "_blank", className: "flex-1 font-mono text-sm text-pulse-primary truncate", children: shortUrl }), _jsx("button", { onClick: handleCopy, className: "p-2 hover:bg-indigo-100 rounded transition", children: copied ? _jsx(Check, { className: "w-4 h-4 text-pulse-success" }) : _jsx(Copy, { className: "w-4 h-4 text-pulse-muted" }) })] })), stats && (_jsxs("div", { className: "bg-white border border-gray-200 rounded-lg p-5 space-y-4", children: [_jsxs("div", { className: "flex justify-between items-start", children: [_jsxs("div", { children: [_jsx("p", { className: "font-mono text-sm text-pulse-primary", children: stats.code }), _jsx("p", { className: "text-xs text-pulse-muted truncate max-w-md", children: stats.url })] }), _jsx("button", { onClick: () => setStats(null), className: "text-pulse-muted hover:text-pulse-text text-sm", children: "\u2715" })] }), _jsxs("div", { className: "text-3xl font-bold", children: [stats.clicks, " ", _jsx("span", { className: "text-sm font-normal text-pulse-muted", children: "\u043A\u043B\u0438\u043A\u043E\u0432" })] }), stats.clicks_by_day && stats.clicks_by_day.length > 0 && (_jsxs("div", { children: [_jsx("p", { className: "text-xs font-medium text-pulse-muted mb-2 uppercase tracking-wider", children: "\u041A\u043B\u0438\u043A\u0438 \u043F\u043E \u0434\u043D\u044F\u043C" }), _jsx("div", { className: "flex items-end gap-1 h-20", children: stats.clicks_by_day.map((d) => {
                                            const max = Math.max(...stats.clicks_by_day.map((x) => x.count));
                                            return (_jsx("div", { className: "flex-1 bg-pulse-primary/20 hover:bg-pulse-primary/40 rounded-t transition", style: { height: `${(d.count / max) * 100}%` }, title: `${d.date}: ${d.count}` }, d.date));
                                        }) })] })), _jsx("div", { className: "grid grid-cols-2 gap-4", children: stats.top_referers && stats.top_referers.length > 0 && (_jsxs("div", { children: [_jsx("p", { className: "text-xs font-medium text-pulse-muted mb-1 uppercase tracking-wider", children: "\u0418\u0441\u0442\u043E\u0447\u043D\u0438\u043A\u0438" }), stats.top_referers.map((r) => (_jsxs("div", { className: "flex justify-between text-xs py-0.5", children: [_jsx("span", { className: "truncate", children: r.key }), _jsx("span", { className: "font-mono text-pulse-muted ml-2", children: r.count })] }, r.key)))] })) })] })), _jsx("div", { className: "space-y-2", children: links.map((link) => (_jsxs("div", { className: "flex items-center gap-3 bg-white border border-gray-200 rounded-lg px-4 py-3", children: [_jsxs("div", { className: "flex-1 min-w-0", children: [_jsx("p", { className: "font-mono text-sm text-pulse-primary", children: link.code }), _jsx("p", { className: "text-xs text-pulse-muted truncate", children: link.url })] }), _jsx("span", { className: "font-mono text-sm text-pulse-muted", children: link.clicks }), _jsx("button", { onClick: () => showStats(link.code), className: "p-1.5 hover:bg-gray-100 rounded transition", title: "\u0421\u0442\u0430\u0442\u0438\u0441\u0442\u0438\u043A\u0430", children: _jsx(BarChart3, { className: "w-4 h-4 text-pulse-muted" }) }), _jsx("a", { href: `/${link.code}`, target: "_blank", className: "p-1.5 hover:bg-gray-100 rounded transition", title: "\u041E\u0442\u043A\u0440\u044B\u0442\u044C", children: _jsx(ExternalLink, { className: "w-4 h-4 text-pulse-muted" }) })] }, link.id))) })] })] }));
}
