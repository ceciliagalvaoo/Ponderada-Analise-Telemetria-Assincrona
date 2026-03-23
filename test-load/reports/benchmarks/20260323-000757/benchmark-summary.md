# Benchmark k6 com limite de CPU/RAM

Data: 2026-03-23 00:17:35
Rodadas por perfil: 3
Duracao por rodada: 45s
VUs por rodada: 30

| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0.25 | 29.787 | 29.7892 | 9.6314 | 0 | 119.1482 | 29.7845 | 300.03 |
| 0.5 | 29.7872 | 29.816 | 10.8853 | 0 | 59.5745 | 29.7845 | 100.02 |
| 1 | 29.7845 | 29.7933 | 11.649 | 0 |  |  |  |

Formula usada: req_s_estimado_1core = req_s_medido / cpu_perfil
