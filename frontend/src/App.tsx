import { useState, useEffect } from "react";
import { Link2, BarChart3, Copy, Check, ExternalLink, Activity } from "lucide-react";

interface Link {
  id: number;
  code: string;
  url: string;
  clicks: number;
  created_at: string;
}

interface Stats {
  id: number;
  code: string;
  url: string;
  clicks: number;
  clicks_by_day: { date: string; count: number }[] | null;
  top_referers: { key: string; count: number }[] | null;
  top_countries: { key: string; count: number }[] | null;
}

interface KVStat { key: string; count: number }
interface DayStat { date: string; count: number }
interface EventStats {
  total: number;
  events_by_day: DayStat[] | null;
  top_events: KVStat[] | null;
  top_projects: KVStat[] | null;
  top_pages: KVStat[] | null;
  top_countries: KVStat[] | null;
}

const API = "/linkpulse/api";

export default function App() {
  const [tab, setTab] = useState<"links" | "events">("links");
  const [url, setUrl] = useState("");
  const [links, setLinks] = useState<Link[]>([]);
  const [stats, setStats] = useState<Stats | null>(null);
  const [shortUrl, setShortUrl] = useState("");
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);

  async function loadLinks() {
    const resp = await fetch(`${API}/links`);
    setLinks(await resp.json());
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!url.trim()) return;
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
    } finally {
      setLoading(false);
    }
  }

  async function handleCopy() {
    await navigator.clipboard.writeText(shortUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  async function showStats(code: string) {
    const resp = await fetch(`${API}/links/${code}/stats`);
    setStats(await resp.json());
  }

  // Load links on first render
  if (links.length === 0 && !loading) {
    loadLinks();
  }

  return (
    <div className="min-h-screen">
      <header className="bg-pulse-primary text-white px-6 py-4">
        <div className="max-w-3xl mx-auto flex items-center gap-3">
          <Link2 className="w-6 h-6" />
          <div>
            <h1 className="text-xl font-bold">LinkPulse</h1>
            <p className="text-xs text-indigo-200 font-mono">URL Shortener + Analytics</p>
          </div>
        </div>
      </header>

      <main className="max-w-3xl mx-auto p-6 space-y-6">
        {/* Tabs */}
        <div className="flex gap-2 border-b border-gray-200">
          <button
            onClick={() => setTab("links")}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition ${
              tab === "links"
                ? "border-pulse-primary text-pulse-primary"
                : "border-transparent text-pulse-muted hover:text-pulse-text"
            }`}
          >
            <Link2 className="w-4 h-4 inline mr-1.5 -mt-0.5" /> Links
          </button>
          <button
            onClick={() => setTab("events")}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition ${
              tab === "events"
                ? "border-pulse-primary text-pulse-primary"
                : "border-transparent text-pulse-muted hover:text-pulse-text"
            }`}
          >
            <Activity className="w-4 h-4 inline mr-1.5 -mt-0.5" /> Events
          </button>
        </div>

        {tab === "events" && <EventsTab />}
        {tab === "links" && <>

        {/* Create form */}
        <form onSubmit={handleSubmit} className="flex gap-3">
          <input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/very-long-url..."
            className="flex-1 px-4 py-3 rounded-lg border border-gray-300 text-sm focus:outline-none focus:border-pulse-primary focus:ring-1 focus:ring-pulse-primary"
          />
          <button
            disabled={loading || !url.trim()}
            className="px-6 py-3 bg-pulse-primary text-white rounded-lg font-medium text-sm hover:bg-pulse-dark disabled:opacity-40 transition"
          >
            {loading ? "..." : "Сократить"}
          </button>
        </form>

        {/* Result */}
        {shortUrl && (
          <div className="flex items-center gap-3 bg-pulse-light border border-indigo-200 rounded-lg px-4 py-3">
            <a href={shortUrl} target="_blank" className="flex-1 font-mono text-sm text-pulse-primary truncate">
              {shortUrl}
            </a>
            <button onClick={handleCopy} className="p-2 hover:bg-indigo-100 rounded transition">
              {copied ? <Check className="w-4 h-4 text-pulse-success" /> : <Copy className="w-4 h-4 text-pulse-muted" />}
            </button>
          </div>
        )}

        {/* Stats modal */}
        {stats && (
          <div className="bg-white border border-gray-200 rounded-lg p-5 space-y-4">
            <div className="flex justify-between items-start">
              <div>
                <p className="font-mono text-sm text-pulse-primary">{stats.code}</p>
                <p className="text-xs text-pulse-muted truncate max-w-md">{stats.url}</p>
              </div>
              <button onClick={() => setStats(null)} className="text-pulse-muted hover:text-pulse-text text-sm">✕</button>
            </div>
            <div className="text-3xl font-bold">{stats.clicks} <span className="text-sm font-normal text-pulse-muted">кликов</span></div>

            {stats.clicks_by_day && stats.clicks_by_day.length > 0 && (
              <div>
                <p className="text-xs font-medium text-pulse-muted mb-2 uppercase tracking-wider">Клики по дням</p>
                <div className="flex items-end gap-1 h-20">
                  {stats.clicks_by_day.map((d) => {
                    const max = Math.max(...stats.clicks_by_day!.map((x) => x.count));
                    return (
                      <div key={d.date} className="flex-1 bg-pulse-primary/20 hover:bg-pulse-primary/40 rounded-t transition" style={{ height: `${(d.count / max) * 100}%` }} title={`${d.date}: ${d.count}`} />
                    );
                  })}
                </div>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              {stats.top_referers && stats.top_referers.length > 0 && (
                <div>
                  <p className="text-xs font-medium text-pulse-muted mb-1 uppercase tracking-wider">Источники</p>
                  {stats.top_referers.map((r) => (
                    <div key={r.key} className="flex justify-between text-xs py-0.5">
                      <span className="truncate">{r.key}</span>
                      <span className="font-mono text-pulse-muted ml-2">{r.count}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Links list */}
        <div className="space-y-2">
          {links.map((link) => (
            <div key={link.id} className="flex items-center gap-3 bg-white border border-gray-200 rounded-lg px-4 py-3">
              <div className="flex-1 min-w-0">
                <p className="font-mono text-sm text-pulse-primary">{link.code}</p>
                <p className="text-xs text-pulse-muted truncate">{link.url}</p>
              </div>
              <span className="font-mono text-sm text-pulse-muted">{link.clicks}</span>
              <button onClick={() => showStats(link.code)} className="p-1.5 hover:bg-gray-100 rounded transition" title="Статистика">
                <BarChart3 className="w-4 h-4 text-pulse-muted" />
              </button>
              <a href={`/${link.code}`} target="_blank" className="p-1.5 hover:bg-gray-100 rounded transition" title="Открыть">
                <ExternalLink className="w-4 h-4 text-pulse-muted" />
              </a>
            </div>
          ))}
        </div>
        </>}
      </main>
    </div>
  );
}

function EventsTab() {
  const [project, setProject] = useState("");
  const [since, setSince] = useState(30);
  const [stats, setStats] = useState<EventStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    const url = `${API}/events/stats?since=${since}${project ? `&project=${encodeURIComponent(project)}` : ""}`;
    fetch(url)
      .then((r) => r.json())
      .then((d) => setStats(d))
      .finally(() => setLoading(false));
  }, [project, since]);

  const max = stats?.events_by_day?.length
    ? Math.max(...stats.events_by_day.map((d) => d.count))
    : 0;

  return (
    <div className="space-y-5">
      {/* Filters */}
      <div className="flex gap-3 items-center">
        <input
          value={project}
          onChange={(e) => setProject(e.target.value)}
          placeholder="project (empty = all)"
          className="px-3 py-2 rounded-lg border border-gray-300 text-sm focus:outline-none focus:border-pulse-primary"
        />
        <select
          value={since}
          onChange={(e) => setSince(parseInt(e.target.value, 10))}
          className="px-3 py-2 rounded-lg border border-gray-300 text-sm focus:outline-none focus:border-pulse-primary"
        >
          <option value={1}>last 24h</option>
          <option value={7}>last 7 days</option>
          <option value={30}>last 30 days</option>
          <option value={90}>last 90 days</option>
        </select>
        {loading && <span className="text-xs text-pulse-muted">loading…</span>}
      </div>

      {!stats && !loading && (
        <div className="text-sm text-pulse-muted">No data yet. Trigger events from your apps via <code>cheslav.track(...)</code>.</div>
      )}

      {stats && (
        <>
          <div className="text-3xl font-bold">
            {stats.total} <span className="text-sm font-normal text-pulse-muted">events</span>
          </div>

          {stats.events_by_day && stats.events_by_day.length > 0 && (
            <div>
              <p className="text-xs font-medium text-pulse-muted mb-2 uppercase tracking-wider">Events per day</p>
              <div className="flex items-end gap-1 h-20">
                {stats.events_by_day.map((d) => (
                  <div
                    key={d.date}
                    className="flex-1 bg-pulse-primary/20 hover:bg-pulse-primary/50 rounded-t transition"
                    style={{ height: max > 0 ? `${(d.count / max) * 100}%` : "0" }}
                    title={`${d.date}: ${d.count}`}
                  />
                ))}
              </div>
            </div>
          )}

          <div className="grid grid-cols-2 gap-5">
            <KVList title="Top events" items={stats.top_events} />
            {stats.top_projects && <KVList title="Top projects" items={stats.top_projects} />}
            <KVList title="Top pages" items={stats.top_pages} />
            <KVList title="Top countries" items={stats.top_countries} />
          </div>
        </>
      )}
    </div>
  );
}

function KVList({ title, items }: { title: string; items: KVStat[] | null }) {
  if (!items || items.length === 0) return null;
  return (
    <div>
      <p className="text-xs font-medium text-pulse-muted mb-1 uppercase tracking-wider">{title}</p>
      {items.map((r) => (
        <div key={r.key} className="flex justify-between text-xs py-0.5">
          <span className="truncate">{r.key}</span>
          <span className="font-mono text-pulse-muted ml-2">{r.count}</span>
        </div>
      ))}
    </div>
  );
}