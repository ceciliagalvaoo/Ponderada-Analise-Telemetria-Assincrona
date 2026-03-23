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

## 3. Benchmark reproduzivel com limite de CPU/RAM

O benchmark reproduzivel usa limites de recursos no arquivo `docker-compose.bench.yml` e automacao no script `test-load/run-benchmark.ps1`.

Perfis padrao executados:

- `0.25` core
- `0.5` core
- `1.0` core

Rodar benchmark completo (3 rodadas por perfil):

```bash
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
./test-load/run-benchmark.ps1
```

Rodar benchmark com parametros personalizados:

```bash
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
./test-load/run-benchmark.ps1 -Runs 5 -Duration 60s -Vus 40
```

Verificar limites aplicados em runtime (exemplo para `back`):

```bash
docker inspect back --format "CPU={{.HostConfig.NanoCpus}} Memory={{.HostConfig.Memory}}"
```

## 4. Artefatos gerados

Uma execucao do benchmark completa cria uma pasta em `test-load/reports/benchmarks/<timestamp>/` com:

- `raw-runs.json`: metrica de cada rodada por perfil
- `benchmark-summary.json`: consolidacao por perfil
- `benchmark-summary.csv`: consolidacao em tabela
- `benchmark-summary.md`: resumo pronto para anexar no relatorio
- `cpu-0_25/run-*-summary.json`, `cpu-0_5/run-*-summary.json`, `cpu-1_0/run-*-summary.json`
- `cpu-*/run-*.log`: saida textual completa de cada rodada

O arquivo `test-load/reports/benchmarks/latest.txt` guarda o ultimo timestamp executado.

## 5. Artefatos do teste simples

- `summary.json`: resumo estruturado com métricas agregadas
- `run.log`: saída textual da execução, quando redirecionada

## 6. Resultado de referência

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

## 7. Como interpretar os campos principais

- `http_reqs`: quantidade total e taxa de requisições por segundo
- `checks`: validações funcionais definidas no script, neste caso o status HTTP esperado
- `http_req_failed`: percentual de falhas de rede ou protocolo HTTP
- `http_req_duration`: tempo total de duração da requisição
- `p(90)` e `p(95)`: percentis que ajudam a observar o comportamento da cauda de latência

## 8. Cuidados para reprodução fiel

- garantir que Docker Desktop esteja aberto antes de subir o ambiente
- validar `docker compose ps` com os serviços `back`, `middleware`, `rabbitmq` e `db` em estado ativo
- se houver erro de conexão no início, reiniciar `back` e `middleware` após banco e broker estarem prontos

Para benchmark comparativo:

- manter mesma quantidade de rodadas por perfil
- manter os mesmos parametros de VUs e duracao em todos os perfis
- evitar outros processos pesados no host durante o teste

## 9. Extrapolacao para 1 core (regra de 3)

O resumo automatico calcula a extrapolacao para perfis sub-core com:

```text
req_s_estimado_1core = req_s_medido / cpu_perfil
```

E compara com a medicao real em `1.0` core:

```text
erro_percentual = ((estimado - real_1core) / real_1core) * 100
```

Esse comparativo mostra o quanto a escala foi linear ou nao no ambiente medido.

## 10. Resultado de referencia do benchmark final

Fonte: `test-load/reports/benchmarks/20260323-000757/benchmark-summary.md`

| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0.25 | 29.7870 | 29.7892 | 9.6314 | 0 | 119.1482 | 29.7845 | 300.03 |
| 0.5 | 29.7872 | 29.8160 | 10.8853 | 0 | 59.5745 | 29.7845 | 100.02 |
| 1.0 | 29.7845 | 29.7933 | 11.6490 | 0 | - | - | - |

Interpretacao:

- a taxa de requisicoes ficou praticamente igual entre perfis
- a extrapolacao linear nao representou bem o 1 core real para este cenario
- como o script usa `sleep(1)`, o limite de throughput pode vir do cenario de carga e nao apenas de CPU

Leitura quantitativa:

- diferenca de throughput medio entre perfis: 0.0027 req/s (max 29.7872, min 29.7845)
- p95 medio ficou entre 9.6314 ms e 11.6490 ms
- erro da extrapolacao para 1 core: 300.03% (0.25 core) e 100.02% (0.5 core)

Implicacao para o experimento:

- os resultados sao excelentes para evidenciar estabilidade
- para evidenciar escalabilidade de CPU, recomenda-se um cenario sem gargalo de think time

## 11. Melhorias futuras para os testes

- exportar resultados em séries históricas para comparação entre versões
- adicionar cenários de stress, soak test e spike test
- simular payloads discretos e analógicos em proporções diferentes
- incluir thresholds no script, como limite máximo aceitável para `p95`
- automatizar a execução em CI para prevenir regressões de desempenho
- repetir o benchmark com redução/remoção de `sleep(1)` no script k6 para medir escalabilidade por CPU sem gargalo de think time
