/**
 * K6 Load Test — uninaquiz-backend: Auth Routes
 *
 * Simulates concurrent users registering and logging in.
 *
 * Run:
 *   k6 run tests/load/auth_load_test.js \
 *     -e BASE_URL=http://localhost:8080
 *
 * Thresholds:
 *   - 99% of requests must respond HTTP 2xx or 4xx (no 5xx)
 *   - p(95) latency < 250ms
 *   - error rate < 1%
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";
import { randomString } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

// ─── Custom Metrics ────────────────────────────────────────────────────────────
const registrationErrors = new Counter("registration_errors");
const loginErrors = new Counter("login_errors");
const errorRate = new Rate("error_rate");
const loginDuration = new Trend("login_duration_ms", true);

// ─── Test Options ──────────────────────────────────────────────────────────────
export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp-up: 0 → 10 VUs
    { duration: "1m", target: 50 }, // Sustain: hold 50 VUs for 1 minute
    { duration: "30s", target: 100 }, // Stress spike: ramp to 100 VUs
    { duration: "1m", target: 100 }, // Sustain peak load
    { duration: "30s", target: 0 }, // Ramp-down: graceful shutdown
  ],
  thresholds: {
    // p(95) of all HTTP request durations must be below 250ms
    http_req_duration: ["p(95)<250"],
    // No more than 1% of iterations should fail
    error_rate: ["rate<0.01"],
    // At least 99% of login requests complete in < 300ms
    login_duration_ms: ["p(99)<300"],
  },
};

// ─── Helpers ───────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

function makeUser() {
  return {
    username: `loadtest_${randomString(10)}`,
    password: `Load@Test#${randomString(8)}`,
  };
}

// ─── Default Scenario ──────────────────────────────────────────────────────────
export default function () {
  const user = makeUser();
  const headers = { "Content-Type": "application/json" };

  // ── 1. Register ──────────────────────────────────────────────────────────────
  const registerRes = http.post(
    `${BASE_URL}/api/auth/register`,
    JSON.stringify(user),
    { headers, tags: { name: "auth_register" } }
  );

  const registerOk = check(registerRes, {
    "register: status is 201": (r) => r.status === 201,
    "register: token present": (r) => {
      try {
        return JSON.parse(r.body).token !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (!registerOk) {
    registrationErrors.add(1);
    errorRate.add(1);
    sleep(1);
    return;
  }

  errorRate.add(0);

  // ── 2. Login ─────────────────────────────────────────────────────────────────
  const loginStart = Date.now();
  const loginRes = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify(user),
    { headers, tags: { name: "auth_login" } }
  );
  loginDuration.add(Date.now() - loginStart);

  const loginOk = check(loginRes, {
    "login: status is 200": (r) => r.status === 200,
    "login: token present": (r) => {
      try {
        return JSON.parse(r.body).token !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (!loginOk) {
    loginErrors.add(1);
    errorRate.add(1);
  } else {
    errorRate.add(0);
  }

  // ── 3. Login with wrong password (expect 401, not a 5xx) ─────────────────────
  const badLoginRes = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify({ username: user.username, password: "totally_wrong" }),
    { headers, tags: { name: "auth_login_bad_credentials" } }
  );

  check(badLoginRes, {
    "bad login: status is 401": (r) => r.status === 401,
  });

  // ── 4. Logout ─────────────────────────────────────────────────────────────────
  const logoutRes = http.post(`${BASE_URL}/api/auth/logout`, null, {
    tags: { name: "auth_logout" },
  });
  check(logoutRes, {
    "logout: status is 200": (r) => r.status === 200,
  });

  sleep(1); // Think time between VU iterations
}

