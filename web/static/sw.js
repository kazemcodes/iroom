const CACHE_NAME = 'iroom-v1';
const STATIC_CACHE = 'iroom-static-v1';
const DYNAMIC_CACHE = 'iroom-dynamic-v1';
const API_CACHE = 'iroom-api-v1';

const STATIC_ASSETS = [
  '/',
  '/manifest.json',
  '/favicon.svg',
  '/robots.txt'
];

const STATIC_EXTENSIONS = /\.(css|js|woff2?|ttf|otf|eot|svg|png|jpg|jpeg|gif|ico|webp)$/;

// Install: pre-cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => cache.addAll(STATIC_ASSETS))
      .then(() => self.skipWaiting())
  );
});

// Activate: clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => key !== STATIC_CACHE && key !== DYNAMIC_CACHE && key !== API_CACHE)
            .map((key) => caches.delete(key))
        )
      )
      .then(() => self.clients.claim())
  );
});

// Fetch: route to appropriate strategy
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== 'GET') return;

  // API calls: network-first with cache fallback
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(networkFirst(request, API_CACHE));
    return;
  }

  // Static assets: cache-first with network fallback
  if (STATIC_EXTENSIONS.test(url.pathname) || url.pathname.startsWith('/_app/')) {
    event.respondWith(cacheFirst(request, STATIC_CACHE));
    return;
  }

  // HTML pages: network-first with cache fallback
  if (request.headers.get('accept')?.includes('text/html')) {
    event.respondWith(networkFirst(request, DYNAMIC_CACHE));
    return;
  }

  // Everything else: cache-first
  event.respondWith(cacheFirst(request, DYNAMIC_CACHE));
});

// Cache-first strategy
async function cacheFirst(request, cacheName) {
  const cached = await caches.match(request);
  if (cached) return cached;

  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(cacheName);
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    // Return offline fallback for navigation requests
    if (request.headers.get('accept')?.includes('text/html')) {
      return caches.match('/');
    }
    return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
  }
}

// Network-first strategy
async function networkFirst(request, cacheName) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(cacheName);
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    const cached = await caches.match(request);
    if (cached) return cached;
    return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
  }
}

// Push notification handler (placeholder)
self.addEventListener('push', (event) => {
  if (!event.data) return;

  let data = {};
  try {
    data = event.data.json();
  } catch {
    data = { title: event.data.text() };
  }

  const options = {
    body: data.body || 'New notification',
    icon: '/icon-192.png',
    badge: '/icon-192.png',
    dir: 'rtl',
    lang: 'fa',
    data: data.url || '/',
    actions: data.actions || []
  };

  event.waitUntil(
    self.registration.showNotification(data.title || 'آی‌روم', options)
  );
});

// Notification click handler
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  const targetUrl = event.notification.data || '/';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((windowClients) => {
        // Focus existing window if open
        for (const client of windowClients) {
          if (client.url.includes(targetUrl) && 'focus' in client) {
            return client.focus();
          }
        }
        // Open new window
        if (clients.openWindow) {
          return clients.openWindow(targetUrl);
        }
      })
  );
});

// Background sync handler (placeholder)
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-pending-actions') {
    event.waitUntil(syncPendingActions());
  }
});

async function syncPendingActions() {
  // Placeholder: implement offline action syncing
  // e.g., retry failed API calls stored in IndexedDB
  console.log('[SW] Background sync: syncing pending actions');
}

// Message handler for cache management
self.addEventListener('message', (event) => {
  if (event.data === 'skipWaiting') {
    self.skipWaiting();
  }
  if (event.data === 'clearCache') {
    caches.keys().then((keys) => Promise.all(keys.map((key) => caches.delete(key))));
  }
});
