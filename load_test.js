import http from "k6/http";
import { check } from "k6";

export const options = {
  duration: "10s",
  vu: 50,
  thresholds: {
    http_req_duration: ["p(95)<2000"], // Let’s allow some slack here
    http_req_failed: ["rate<0.05"], // Acceptable failure threshold
    checks: ["rate>0.95"], // Must pass status checks
  },
  summaryTrendStats: ["avg", "min", "med", "p(90)", "p(95)", "max"],
};

export default function () {
  const res = http.get("http://localhost:8080/api/health");
  check(res, { "status is 200": (r) => r.status === 200 });
  // ❌ No sleep — we fire as fast as CPU/IO allows
}
