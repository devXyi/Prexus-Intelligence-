/**
 * js/store.js
 * Prexus Intelligence — Global State Store
 * THE GREAT FILE · Phase 1
 *
 * Simple event-based store. No framework dependency.
 * All modules import from here. No module holds its own state.
 *
 * Usage:
 *   import { store, subscribe } from './store.js';
 *   store.set('user', userData);
 *   subscribe('user', (user) => console.log(user));
 */

const _state = {
  // ── Session ─────────────────────────────────────────────
  token:    null,
  user:     null,    // { id, email, full_name, org_name, role }
  org:      null,    // { orgName, orgType, country, domain, ... }

  // ── Navigation ──────────────────────────────────────────
  page:     'auth',  // 'auth' | 'org' | 'hub' | 'meteorium'
  module:   'dashboard',

  // ── Assets ──────────────────────────────────────────────
  assets:   [],
  selectedAsset: null,

  // ── Loading states ───────────────────────────────────────
  // Set of string keys e.g. 'auth', 'assets', 'risk:MUM-001'
  loading:  new Set(),

  // ── Errors ───────────────────────────────────────────────
  // Map of key → error message
  errors:   new Map(),
};

// Subscribers: Map of key → Set of callbacks
const _subscribers = new Map();

/**
 * The public store API.
 */
export const store = {

  /** Get a value by key. */
  get(key) {
    return _state[key];
  },

  /** Set a value and notify subscribers. */
  set(key, value) {
    _state[key] = value;
    _notify(key, value);
  },

  /** Set multiple values at once. */
  setMany(updates) {
    for (const [key, value] of Object.entries(updates)) {
      _state[key] = value;
      _notify(key, value);
    }
  },

  /** Check if a loading key is active. */
  isLoading(key) {
    return _state.loading.has(key);
  },

  /** Start loading state. */
  startLoading(key) {
    _state.loading.add(key);
    _notify('loading', _state.loading);
  },

  /** End loading state. */
  stopLoading(key) {
    _state.loading.delete(key);
    _notify('loading', _state.loading);
  },

  /** Set error for a key. */
  setError(key, message) {
    _state.errors.set(key, message);
    _notify('errors', _state.errors);
  },

  /** Clear error for a key. */
  clearError(key) {
    _state.errors.delete(key);
    _notify('errors', _state.errors);
  },

  /** Get error for a key. */
  getError(key) {
    return _state.errors.get(key) || null;
  },
};

/**
 * Subscribe to state changes for a specific key.
 * Returns an unsubscribe function.
 */
export function subscribe(key, callback) {
  if (!_subscribers.has(key)) {
    _subscribers.set(key, new Set());
  }
  _subscribers.get(key).add(callback);

  // Return unsubscribe
  return () => {
    _subscribers.get(key)?.delete(callback);
  };
}

/** Internal: notify all subscribers for a key. */
function _notify(key, value) {
  _subscribers.get(key)?.forEach(cb => {
    try {
      cb(value);
    } catch (e) {
      console.error(`[store] subscriber error for key "${key}":`, e);
    }
  });
}

/**
 * Persist session to localStorage.
 * Called after successful login/register.
 */
export function persistSession(token, user) {
  try {
    localStorage.setItem('prx_token', token);
    localStorage.setItem('prx_user', JSON.stringify(user));
  } catch (e) {
    console.warn('[store] localStorage write failed:', e);
  }
}

/**
 * Restore session from localStorage.
 * Returns true if a valid session was found.
 */
export function restoreSession() {
  try {
    const token = localStorage.getItem('prx_token');
    const userRaw = localStorage.getItem('prx_user');
    if (!token || !userRaw) return false;

    const user = JSON.parse(userRaw);
    if (!user || !user.email) return false;

    store.setMany({ token, user });

    // Restore org if saved
    const orgRaw = localStorage.getItem(`prexus_org_${user.email}`);
    if (orgRaw) {
      store.set('org', JSON.parse(orgRaw));
    }

    return true;
  } catch (e) {
    console.warn('[store] session restore failed:', e);
    return false;
  }
}

/**
 * Persist org data to localStorage.
 */
export function persistOrg(orgData) {
  try {
    const email = store.get('user')?.email;
    if (!email) return;
    localStorage.setItem(`prexus_org_${email}`, JSON.stringify(orgData));
    store.set('org', orgData);
  } catch (e) {
    console.warn('[store] org persist failed:', e);
  }
}

/**
 * Clear all session data (logout).
 */
export function clearSession() {
  try {
    localStorage.removeItem('prx_token');
    localStorage.removeItem('prx_user');
  } catch (e) {
    console.warn('[store] session clear failed:', e);
  }

  store.setMany({
    token:         null,
    user:          null,
    org:           null,
    assets:        [],
    selectedAsset: null,
  });
  store.get('loading').clear();
  store.get('errors').clear();
}

