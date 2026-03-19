# Sistema de Processamento de Telemetria com Arquitetura Assincrona

## 1. Contexto e Objetivo

Este projeto implementa um backend para ingestao de telemetria de dispositivos embarcados, com foco em escalabilidade, resiliencia e desacoplamento entre recepcao e persistencia.

Arquitetura adotada:

Cliente -> Backend (HTTP) -> RabbitMQ -> Middleware (consumidor) -> PostgreSQL

Stack utilizada:

- Go
- RabbitMQ
- PostgreSQL
- Docker / Docker Compose
- k6

## 2. Requisitos Atendidos

- Endpoint HTTP POST para recebimento de telemetria: atendido em `/telemetry`
- Desacoplamento com mensageria: atendido com `telemetry_queue`
- Processamento assincrono: atendido via middleware consumidor
- Persistencia relacional: atendido com PostgreSQL
- Conteinerizacao da infraestrutura: atendido com Docker Compose
- Testes de carga: atendido com k6 e export de sumario em JSON

## 3. Arquitetura do Sistema 

### 3.1 Topologia de implantacao

A arquitetura foi montada em quatro servicos executados por Docker Compose, todos na mesma rede virtual (`app-network`):

- `back`: API HTTP (porta 8080 exposta para host)
- `rabbitmq`: broker AMQP (porta 5672) e painel de gerenciamento (porta 15672)
- `middleware`: consumidor da fila e persistencia
- `db`: PostgreSQL (porta 5432)

Com essa topologia:

- o cliente acessa apenas o `back` via `localhost:8080`
- `back` publica no `rabbitmq` usando DNS interno do Compose (`rabbitmq`)
- `middleware` consome do `rabbitmq` e grava no `db` via DNS interno (`db`)

Essa montagem isola responsabilidades e permite evoluir cada servico de forma independente.

### 3.2 Fluxo arquitetural de ponta a ponta

O caminho da telemetria foi estruturado em duas etapas:

- etapa sincrona: recepcao HTTP e validacao no `back`
- etapa assincrona: enfileiramento, consumo e persistencia no `middleware`

Isso evita que a API fique bloqueada por operacoes de banco durante picos de requisicoes.

### 3.3 Backend sem persistencia direta

O backend recebe, valida e publica mensagens na fila sem escrever no banco.

Justificativa:

- Reduz latencia da API
- Evita acoplamento com I/O pesado no caminho sincrono
- Permite absorver picos de carga com o broker

### 3.4 RabbitMQ como buffer de resiliencia

A fila foi configurada como duravel e as mensagens sao publicadas como persistentes.

Justificativa:

- Menor risco de perda em reinicio do broker
- Suporte a bursts sem bloquear o endpoint

### 3.5 Consumo com ACK manual

O middleware confirma a mensagem apenas apos persistencia no banco.
Em caso de erro, aplica NACK com requeue.

Justificativa:

- Evita perda de mensagem por ack antecipado
- Mantem semantica de entrega mais confiavel no fluxo atual

### 3.6 Validacoes no payload

O backend valida:

- Metodo HTTP
- JSON valido
- `device_id` obrigatorio
- `timestamp` em RFC3339
- `sensor_type` obrigatorio
- `reading_type` em `analog` ou `discrete`

Justificativa:

- Rejeita entrada invalida cedo
- Evita enfileirar dados inconsistentes

### 3.7 Organizacao da infraestrutura

A organizacao dos arquivos de infraestrutura do projeto foi realizada de forma a manter proximidade com o componente responsavel pelo processamento dos dados, neste caso, o middleware.

Embora os servicos de RabbitMQ e PostgreSQL sejam definidos no `docker-compose.yml` como containers independentes, seus arquivos auxiliares (como scripts de inicializacao e documentacao) foram organizados em diretorios relacionados ao fluxo principal da aplicacao.

Essa decisao foi motivada pelos seguintes fatores:

- O middleware e o principal consumidor das mensagens e responsavel pela persistencia dos dados.
- Tanto o RabbitMQ quanto o PostgreSQL sao utilizados diretamente pelo middleware durante o processamento.
- A organizacao proxima ao middleware facilita a compreensao do fluxo de dados, especialmente em projetos de menor escala.
- Reduz a fragmentacao do projeto, evitando a criacao de multiplos diretorios de infraestrutura sem necessidade.

E importante destacar que, do ponto de vista arquitetural, RabbitMQ e PostgreSQL continuam sendo servicos independentes, executados em containers proprios, garantindo isolamento e desacoplamento entre os componentes.

## 4. Fluxo de Execucao do Sistema

### 4.1 Recepcao da requisicao

O cliente envia `POST /telemetry` com `device_id`, `timestamp`, `sensor_type`, `reading_type` e `value`.

### 4.2 Processamento no backend

Ao receber a requisicao, o backend:

- valida metodo HTTP
- desserializa JSON
- valida campos obrigatorios e formato
- serializa payload
- publica mensagem na fila

O backend nao grava em banco nesse momento, mantendo o caminho de resposta leve.

### 4.3 Enfileiramento e consumo

O RabbitMQ atua como buffer entre ingestao e persistencia. O middleware consome continuamente a fila e executa o processamento.

### 4.4 Persistencia e confirmacao

O middleware grava no PostgreSQL e so depois confirma a mensagem (ACK). Em falha de processamento/persistencia, aplica NACK com requeue para nova tentativa.

## 5. Modelo de Dados

Tabela principal:

```sql
CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    sensor_type VARCHAR(50) NOT NULL,
    reading_type VARCHAR(20) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Observacao:

- O campo `value` esta modelado como numerico continuo.
- Leituras discretas podem ser representadas por convencao numerica (ex.: 0/1).
- Uma extensao futura recomendada e separar valor analogico de valor discreto no esquema.

## 6. Como Executar

Pre-requisitos:

- Docker Desktop em execucao
- Docker Compose habilitado
- k6 instalado (para teste de carga)

### 6.1 Subir ambiente

```bash
docker compose up --build -d
docker compose ps
```

Servicos esperados:

- `back`
- `middleware`
- `rabbitmq`
- `db`

### 6.2 Testar endpoint manualmente

```powershell
$body = @{
  device_id = "dev-001"
  timestamp = "2026-03-17T15:00:00Z"
  sensor_type = "temperature"
  reading_type = "analog"
  value = 26.7
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/telemetry" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

Resposta esperada:

```text
mensagem enfileirada com sucesso
```

### 6.3 Validar persistencia

```bash
docker exec db psql -U postgres -d telemetrydb -c "SELECT * FROM telemetry ORDER BY id DESC LIMIT 5;"
```

### 6.4 Executar testes unitarios

```bash
cd back
go test
cd ../middleware
go test
```

### 6.5 Executar teste de carga

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js
```

Opcional:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js > test-load/reports/run.log
```

### 6.6 Encerrar ambiente

```bash
docker compose down
```

Reset completo:

```bash
docker compose down -v
```

## 7. Validacao Container a Container

Checklist pratico usado na validacao:

- `docker compose ps` confirma os 4 servicos ativos
- `back` responde no endpoint `/telemetry`
- `rabbitmqctl list_queues` mostra fila existente
- `middleware` registra mensagens persistidas com sucesso
- consulta SQL confirma gravacao da telemetria

Resultado da validacao:

- Fluxo ponta a ponta confirmado: back -> RabbitMQ -> middleware -> PostgreSQL

## 8. Testes Unitarios

Backend:

- requisicao valida retorna HTTP 202
- JSON invalido retorna HTTP 400
- metodo invalido retorna HTTP 405
- timestamp invalido retorna HTTP 400
- `reading_type` invalido retorna HTTP 400

Middleware:

- mensagem valida persiste no banco
- JSON invalido falha no parse
- timestamp invalido falha na conversao
- erro de banco propaga corretamente

## 9. Teste de Carga (k6)

### 9.1 Configuracao

- Script: `test-load/telemetry.js`
- Cenario: 3 estagios
- Duracao total ativa: 40s
- Pico: 30 usuarios virtuais
- Endpoint: `POST /telemetry`

### 9.2 Resultados mais recentes

Fonte: `test-load/reports/summary.json`

- total de requisicoes: 605
- checks aprovados: 605/605 (100%)
- falhas HTTP: 0%
- throughput: 15.02 req/s
- latencia media: 3.76 ms
- p90: 5.55 ms
- p95: 7.49 ms
- latencia maxima: 66.95 ms

### 9.3 Analise tecnica

- O endpoint manteve estabilidade sem erros durante o cenario proposto.
- A latencia media e baixa para o volume aplicado.
- Picos maximos pontuais sao esperados em sistema conteinerizado local, devido a agendamento de CPU e sincronizacao de I/O.
- A arquitetura assincrona reduz risco de degradacao da API sob carga, deslocando o custo de persistencia para o consumidor.

### 9.4 Evidencias no Repositorio

Arquivos de evidencia:

- `test-load/reports/summary.json`
- `test-load/reports/README.md`
- opcional: `test-load/reports/run.log`

## 10. Validacao do Sistema

A validacao pratica do ambiente foi realizada com:

- verificacao de containers ativos (`docker compose ps`)
- teste manual do endpoint `/telemetry`
- analise de fila no RabbitMQ
- observacao de logs do middleware
- consulta SQL de confirmacao no PostgreSQL
- testes unitarios de backend e middleware
- teste de carga com k6

Esse conjunto de evidencias confirma o funcionamento ponta a ponta do pipeline.

## 11. Melhorias Futuras

- Adicionar `healthcheck` e `depends_on.condition: service_healthy` no Compose para reduzir falhas de startup.
- Implementar retry com backoff exponencial em conexao com RabbitMQ e PostgreSQL.
- Criar dead-letter queue e politica de reprocessamento para erros permanentes.
- Evoluir schema para suportar leitura discreta de forma explicita (ex.: `value_numeric` e `value_discrete`).
- Adicionar autenticacao/autorizacao no endpoint de ingestao.
- Incluir observabilidade com metricas, tracing e dashboards (Prometheus/Grafana).
- Cobrir testes de integracao fim a fim com ambiente conteinerizado.
- Definir contratos de payload versionados para evolucao segura do protocolo.

## 12. Conclusao

A solucao atende ao objetivo da atividade ao implementar uma pipeline assincrona e desacoplada para telemetria em Go, com mensageria, persistencia relacional, containerizacao e validacao por testes unitarios e carga.

Com base nos resultados medidos e no funcionamento ponta a ponta validado, o projeto demonstra boa base para cenarios reais de monitoramento industrial: alta taxa de sucesso, baixa latencia media e comportamento estavel sob concorrencia.

As escolhas de desacoplamento com RabbitMQ, ACK manual no consumidor e persistencia relacional tornam a solucao tecnicamente consistente para evolucao em producao. As melhorias futuras propostas direcionam o proximo ciclo para maior robustez operacional, observabilidade e maturidade arquitetural.




