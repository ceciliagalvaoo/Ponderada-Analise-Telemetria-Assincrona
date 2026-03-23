# Benchmark k6 com limite de CPU/RAM

Data: 2026-03-23 00:06:35
Rodadas por perfil: 1
Duracao por rodada: 10s
VUs por rodada: 10

| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 9.9266 | 9.9266 | 6.9764 | 0 |  |  |  |
| 5 | 9.9246 | 9.9246 | 13.0718 | 0 |  |  |  |
| 25 | 9.9299 | 9.9299 | 7.1396 | 0 |  |  |  |

Formula usada: req_s_estimado_1core = req_s_medido / cpu_perfil
