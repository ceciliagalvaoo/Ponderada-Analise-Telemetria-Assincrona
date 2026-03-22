# Evidências de Teste de Carga com k6

Este diretório concentra os artefatos gerados pelos testes de carga da API. O objetivo é registrar, de forma reprodutível, o comportamento do endpoint `/telemetry` sob concorrência e manter evidências alinhadas com o relatório principal do projeto.

## 1. Objetivo do teste

O teste foi estruturado para validar principalmente:

- estabilidade funcional do endpoint `/telemetry`
- taxa de sucesso sob carga concorrente
- latência em diferentes percentis
- throughput efetivo de requisições

## 2. Como executar

Na raiz do projeto:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js
```

Opcionalmente, para salvar também a saída textual:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js > test-load/reports/run.log
```

## 3. Artefatos gerados

- `summary.json`: resumo estruturado com métricas agregadas
- `run.log`: saída textual da execução, quando redirecionada

## 4. Resultado de referência

Fonte: `summary.json`

| Métrica | Valor |
| --- | --- |
| `http_reqs.count` | 605 |
| `http_reqs.rate` | 15.02 req/s |
| `checks` | 605 passes, 0 fails |
| `http_req_failed.value` | 0 |
| `http_req_duration.avg` | 3.09 ms |
| `http_req_duration.med` | 2.37 ms |
| `http_req_duration.p(90)` | 4.99 ms |
| `http_req_duration.p(95)` | 6.28 ms |
| `http_req_duration.max` | 121.10 ms |

Interpretação:

- o cenário executado manteve sucesso total no endpoint
- a latência média permaneceu baixa para o perfil de carga aplicado
- o pico máximo foi pontual e não comprometeu a estabilidade geral do teste

## 5. Como interpretar os campos principais

- `http_reqs`: quantidade total e taxa de requisições por segundo
- `checks`: validações funcionais definidas no script, neste caso o status HTTP esperado
- `http_req_failed`: percentual de falhas de rede ou protocolo HTTP
- `http_req_duration`: tempo total de duração da requisição
- `p(90)` e `p(95)`: percentis que ajudam a observar o comportamento da cauda de latência

## 6. Cuidados para reprodução fiel

- garantir que Docker Desktop esteja aberto antes de subir o ambiente
- validar `docker compose ps` com os serviços `back`, `middleware`, `rabbitmq` e `db` em estado ativo
- se houver erro de conexão no início, reiniciar `back` e `middleware` após banco e broker estarem prontos

## 7. Melhorias futuras para os testes

- exportar resultados em séries históricas para comparação entre versões
- adicionar cenários de stress, soak test e spike test
- simular payloads discretos e analógicos em proporções diferentes
- incluir thresholds no script, como limite máximo aceitável para `p95`
- automatizar a execução em CI para prevenir regressões de desempenho
