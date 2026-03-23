import http from 'k6/http';
import { check, sleep } from 'k6';

function parseIntEnv(name, fallbackValue) {
  const rawValue = __ENV[name];
  if (!rawValue) {
    return fallbackValue;
  }

  const parsed = parseInt(rawValue, 10);
  return Number.isNaN(parsed) ? fallbackValue : parsed;
}

const targetUrl = __ENV.TARGET_URL || 'http://localhost:8080/telemetry';
const testProfile = __ENV.TEST_PROFILE || 'default';
const duration = __ENV.K6_DURATION;
const vus = parseIntEnv('K6_VUS', 0);

export const options = {
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<100'],
  },
  tags: {
    profile: testProfile,
  },
  ...(duration && vus > 0
    ? { vus, duration }
    : {
        stages: [
          { duration: '10s', target: 10 },
          { duration: '20s', target: 30 },
          { duration: '10s', target: 0 },
        ],
      }),
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

  const res = http.post(targetUrl, payload, params);

  check(res, {
    'status is 202': (r) => r.status === 202,
  });

  sleep(1);
}