/**
 * K6 Load Test — uninaquiz-backend: Quiz Routes
 *
 * Simulates authenticated users reading quiz history and fetching individual
 * quizzes.  Quiz generation (AI) is excluded from load testing as it is
 * rate-limited by the upstream Gemini API.
 *
 * Run:
 *   # Step 1 — seed a user and capture their JWT token:
 *   export JWT_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
 *     -H "Content-Type: application/json" \
 *     -d '{"username":"loadtest_user","password":"LoadTest@123"}' | jq -r .token)
 *
 *   # Step 2 — run the load test:
 *   k6 run tests/load/quiz_load_test.js \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e JWT_TOKEN=$JWT_TOKEN
 *
 * Thresholds:
 *   - p(95) of all HTTP request durations < 250ms
 *   - error rate < 1%
 *   - 99% of quiz history requests return 200
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// ─── Custom Metrics ────────────────────────────────────────────────────────────
const errorRate = new Rate("quiz_error_rate");
const historyDuration = new Trend("quiz_history_duration_ms", true);
const getQuizDuration = new Trend("get_quiz_duration_ms", true);

// ─── Options ───────────────────────────────────────────────────────────────────
export const options = {
  stages: [
    { duration: "20s", target: 5 },   // Ramp-up
    { duration: "1m", target: 30 },   // Sustain moderate load
    { duration: "30s", target: 80 },  // Spike
    { duration: "1m", target: 80 },   // Hold spike
    { duration: "20s", target: 0 },   // Ramp-down
  ],
  thresholds: {
    http_req_duration: ["p(95)<250"],
    quiz_error_rate: ["rate<0.01"],
    quiz_history_duration_ms: ["p(95)<200"],
    get_quiz_duration_ms: ["p(95)<200"],
  },
};

// ─── Config ────────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const JWT_TOKEN = __ENV.JWT_TOKEN || "";

// ─── Setup ─────────────────────────────────────────────────────────────────────
// setup() runs once before the load test begins.
// It creates a fresh user, seeds a quiz, and returns shared data to all VUs.
export function setup() {
  if (!JWT_TOKEN) {
    // Auto-register a test user for the session
    const user = {
      username: `k6_quiz_user_${Date.now()}`,
      password: "K6@LoadTest!123",
    };
    const headers = { "Content-Type": "application/json" };
    const res = http.post(
      `${BASE_URL}/api/auth/register`,
      JSON.stringify(user),
      { headers }
    );

    if (res.status !== 201) {
      console.error(`Setup failed: register returned ${res.status} — ${res.body}`);
      return { token: null, quizID: null };
    }

    const body = JSON.parse(res.body);
    return { token: body.token, userID: body.user.id, quizID: null };
  }

  return { token: JWT_TOKEN, quizID: null };
}

// ─── Default Scenario ──────────────────────────────────────────────────────────
export default function (data) {
  if (!data || !data.token) {
    console.warn("No token available — skipping VU iteration");
    return;
  }

  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${data.token}`,
  };

  // ── 1. GET /api/quiz/history ──────────────────────────────────────────────────
  const histStart = Date.now();
  const histRes = http.get(`${BASE_URL}/api/quiz/history`, {
    headers,
    tags: { name: "quiz_get_history" },
  });
  historyDuration.add(Date.now() - histStart);

  const histOk = check(histRes, {
    "quiz history: status 200": (r) => r.status === 200,
    "quiz history: body is array": (r) => {
      try {
        return Array.isArray(JSON.parse(r.body));
      } catch {
        return false;
      }
    },
  });

  errorRate.add(histOk ? 0 : 1);

  // ── 2. If history has items, GET a specific quiz ──────────────────────────────
  if (histOk && histRes.status === 200) {
    let quizzes = [];
    try {
      quizzes = JSON.parse(histRes.body);
    } catch {
      /* ignore parse error */
    }

    if (quizzes.length > 0) {
      const quizID = quizzes[0].id;
      const quizStart = Date.now();
      const quizRes = http.get(`${BASE_URL}/api/quiz/${quizID}`, {
        headers,
        tags: { name: "quiz_get_single" },
      });
      getQuizDuration.add(Date.now() - quizStart);

      const quizOk = check(quizRes, {
        "get quiz: status 200": (r) => r.status === 200,
        "get quiz: has id field": (r) => {
          try {
            return JSON.parse(r.body).id !== undefined;
          } catch {
            return false;
          }
        },
      });
      errorRate.add(quizOk ? 0 : 1);
    }
  }

  // ── 3. Attempt to access quiz with invalid ID → should return 404, not 500 ───
  const notFoundRes = http.get(`${BASE_URL}/api/quiz/non-existent-id`, {
    headers,
    tags: { name: "quiz_get_nonexistent" },
  });
  check(notFoundRes, {
    "nonexistent quiz: status is 404": (r) => r.status === 404,
  });

  sleep(0.5);
}

// ─── Teardown ──────────────────────────────────────────────────────────────────
// teardown() runs once after all VUs finish. Useful for cleanup reporting.
export function teardown(data) {
  console.log(
    `Load test complete. Token used: ${data && data.token ? "yes" : "no"}`
  );
}

