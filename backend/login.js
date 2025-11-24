import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 50, // 50 concurrent users
  duration: "10s", // run for 10 seconds
};

export default function () {
  const payload = JSON.stringify({
    email: "sawezfaisals123@gmail.com",
    password: "pass12345",
  });

  const headers = {
    "Content-Type": "application/json",
  };

  const res = http.post("http://localhost:8080/api/auth/login", payload, {
    headers,
  });

  check(res, {
    "status is 200 or 401": (r) => r.status === 200 || r.status === 401,
  });

  sleep(0.1);
}
