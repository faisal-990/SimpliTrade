import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 1000, // number of concurrent users
  duration: "5s", // run for 5 seconds
};

export default function () {
  const random = Math.floor(Math.random() * 1e9);

  const payload = JSON.stringify({
    name: "Rahul",
    email: `testting_${random}@gmail.com`,
    password: "pass123345",
  });

  const headers = {
    "Content-Type": "application/json",
  };

  const res = http.post("http://localhost:8080/api/auth/signup", payload, {
    headers,
  });

  check(res, {
    "status is 200 or 409": (r) => r.status == 200 || r.status == 409,
  });

  sleep(0.1);
}
