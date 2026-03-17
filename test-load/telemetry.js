import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '20s', target: 30 },
    { duration: '10s', target: 0 },
  ],
};

export default function () {
  const payload = JSON.stringify({
    device_id: `dev-${__VU}`,
    timestamp: new Date().toISOString(),
    sensor_type: 'temperature',
    reading_type: 'analog',
    value: Math.random() * 100,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post('http://localhost:8080/telemetry', payload, params);

  check(res, {
    'status is 202': (r) => r.status === 202,
  });

  sleep(1);
}