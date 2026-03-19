# Evidencias de Teste de Carga (k6)

Este diretorio concentra os artefatos de carga para comprovar, de forma reprodutivel, o comportamento da API sob concorrencia.

## 1. Objetivo do teste

Validar:

- estabilidade do endpoint `/telemetry`
- taxa de sucesso sob carga concorrente
- latencia em diferentes percentis
- throughput efetivo de requisicoes

## 2. Como executar

Na raiz do projeto:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js
```

Opcional para salvar a saida textual:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js > test-load/reports/run.log
```

## 3. Artefatos gerados

- `summary.json`: resumo estruturado com metricas agregadas
- `run.log` (opcional): saida textual da execucao

## 4. Resultado de referencia (execucao mais recente)

Fonte: `summary.json`

- `http_reqs.count`: 605
- `http_reqs.rate`: 15.02 req/s
- `checks`: 605 passes, 0 fails (100%)
- `http_req_failed.value`: 0 (0%)
- `http_req_duration.avg`: 3.77 ms
- `http_req_duration.p(90)`: 5.55 ms
- `http_req_duration.p(95)`: 7.49 ms
- `http_req_duration.max`: 66.95 ms

Interpretacao:

- O cenario executado manteve sucesso total no endpoint.
- A latencia media ficou baixa para o perfil de carga aplicado.
- O valor maximo representa pico isolado, sem comprometer disponibilidade.

## 5. Como interpretar os campos principais

- `http_reqs`: quantidade total e taxa de requisicoes por segundo
- `checks`: validacoes funcionais definidas no script (status 202)
- `http_req_failed`: percentual de falhas de rede/protocolo HTTP
- `http_req_duration`: tempo total da requisicao
- `p(90)` e `p(95)`: comportamento em cauda, mais relevante para UX/SLA

## 6. Cuidados para reproducao fiel

- Garantir que Docker Desktop esteja aberto antes de subir o ambiente.
- Validar `docker compose ps` com os servicos `back`, `middleware`, `rabbitmq` e `db` em estado `Up`.
- Se houver erro de conexao no inicio, reiniciar `back` e `middleware` apos banco e broker estarem prontos.

## 7. Melhorias futuras para os testes

- Exportar resultados em series historicas para comparacao entre versoes.
- Adicionar cenarios de stress, soak test e spike test.
- Simular payloads discretos e analogicos em proporcoes diferentes.
- Incluir thresholds no script (ex.: p95 maximo aceitavel).
- Automatizar execucao em CI para prevenir regressao de desempenho.
