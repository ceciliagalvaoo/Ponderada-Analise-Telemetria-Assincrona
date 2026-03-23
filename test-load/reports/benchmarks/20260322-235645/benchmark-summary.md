# Benchmark k6 com limite de CPU/RAM

Data: 2026-03-23 00:01:05
Rodadas por perfil: 1
Duracao por rodada: 15s
VUs por rodada: 10

| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 9.9514 | 9.9514 | 5.706 | 0 |  |  |  |
| 5 | 9.8549 | 9.8549 | 64.6819 | 0 |  |  |  |
| 25 | 9.9154 | 9.9154 | 8.0536 | 0 |  |  |  |

Formula usada: req_s_estimado_1core = req_s_medido / cpu_perfil
