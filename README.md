# Relatório Técnico

## Sistema de Processamento de Telemetria com Arquitetura Assíncrona

> **Vídeo da solução em funcionamento:** [clique aqui](https://drive.google.com/file/d/1uArArnO_worznHrNW3EcYiT0d8e3jvqY/view?usp=sharing)

**Stack utilizada:** Go, RabbitMQ, PostgreSQL, Docker Compose e k6

## 1. Contexto e Objetivo

Este projeto implementa um sistema de ingestão de telemetria para dispositivos embarcados, com foco em escalabilidade, resiliência e separação clara de responsabilidades entre recepção, transporte e persistência dos dados. A necessidade central da arquitetura era permitir que a API HTTP recebesse dados com baixa latência, mesmo em cenários com aumento de carga, sem depender diretamente da velocidade de escrita no banco de dados.

Para atender esse objetivo, a solução foi estruturada como uma pipeline assíncrona. O backend atua como porta de entrada, validando os dados recebidos e publicando as mensagens em uma fila RabbitMQ. Em seguida, um serviço separado, chamado `middleware`, consome essas mensagens e realiza a persistência no PostgreSQL. Com isso, o caminho síncrono da requisição fica mais leve e a responsabilidade de gravação é deslocada para um fluxo desacoplado.

O fluxo principal do sistema pode ser resumido da seguinte forma:

```text
Cliente -> Backend HTTP -> RabbitMQ -> Middleware -> PostgreSQL
```

Essa abordagem traz três ganhos principais. O primeiro é a redução da latência percebida pelo cliente, já que a API não precisa aguardar a gravação no banco para responder. O segundo é a maior resiliência em momentos de pico, pois a fila passa a funcionar como buffer entre a entrada e o processamento. O terceiro é a evolução mais simples da arquitetura, uma vez que cada componente possui uma responsabilidade mais bem definida.

## 2. Requisitos Atendidos

Os requisitos propostos pela atividade foram atendidos de forma completa. O endpoint HTTP `POST /telemetry` foi implementado no serviço `back`, que recebe a telemetria, valida o payload e publica a mensagem na fila. O desacoplamento entre recepção e persistência foi garantido com o uso da fila `telemetry_queue`, tornando o processamento assíncrono e mais robusto frente a picos de carga.

O consumo das mensagens é realizado por um serviço dedicado, o `middleware`, que lê os eventos da fila e persiste os dados no PostgreSQL. Toda a infraestrutura foi conteinerizada com Docker Compose, o que permite reprodução simples do ambiente e isolamento entre os componentes. Além disso, a validação de desempenho foi complementada com testes de carga em k6, incluindo exportação dos resultados em JSON para rastreabilidade e análise posterior.

| Requisito | Atendimento no projeto |
| --- | --- |
| Endpoint HTTP `POST /telemetry` | Implementado no serviço `back` |
| Desacoplamento com mensageria | Realizado com a fila `telemetry_queue` |
| Processamento assíncrono | Executado pelo serviço `middleware` |
| Persistência relacional | Realizada no PostgreSQL |
| Conteinerização | Estruturada com Docker Compose |
| Teste de carga com exportação de resultados | Executado com k6 e `summary.json` |

## 3. Arquitetura do Sistema

### 3.1 Topologia de implantação

O sistema foi organizado em quatro serviços principais executados na mesma rede virtual do Docker Compose:

| Serviço | Exposição | Papel no sistema |
| --- | --- | --- |
| `back` | `8080` | API HTTP, validação e publicação na fila |
| `rabbitmq` | `5672` e `15672` | Broker AMQP e painel de gerenciamento |
| `middleware` | Interno | Consumo da fila e persistência |
| `db` | `5432` | Banco de dados PostgreSQL |

Na prática, o cliente externo interage apenas com a API HTTP. A partir desse ponto, toda a comunicação entre os serviços ocorre por DNS interno do Compose. Isso significa que o backend publica mensagens usando o hostname `rabbitmq`, enquanto o middleware persiste os dados no banco usando o hostname `db`. Essa estratégia simplifica a configuração e reforça o isolamento entre os componentes.

### 3.2 Fluxo de execução ponta a ponta

O fluxo de dados foi dividido em duas etapas complementares: uma etapa síncrona, voltada para a recepção e validação da telemetria, e uma etapa assíncrona, voltada para o processamento efetivo e a persistência.

Na etapa síncrona:

- o cliente envia uma requisição `POST /telemetry` com o payload da leitura
- o backend verifica se o método HTTP está correto
- o JSON recebido é desserializado e validado
- a estrutura é serializada novamente para publicação na fila
- a API devolve ao cliente a confirmação de enfileiramento

Na etapa assíncrona:

- o RabbitMQ atua como buffer entre entrada e processamento
- o `middleware` consome continuamente a `telemetry_queue`
- a mensagem é interpretada e convertida para persistência
- a leitura é gravada no PostgreSQL
- o consumidor envia `ACK` somente após a persistência
- em caso de falha, a mensagem recebe `NACK` com `requeue`

Essa separação é importante porque evita que a API fique bloqueada por operações de I/O de banco de dados durante o atendimento da requisição. Em outras palavras, o endpoint permanece responsivo mesmo quando a etapa de persistência é mais lenta ou está sob maior pressão.

### 3.3 Principais decisões arquiteturais

Uma decisão central do projeto foi impedir que o backend escrevesse diretamente no banco. Essa escolha reduz o acoplamento entre API e persistência e preserva a responsividade do endpoint. Em vez de transformar o backend em um componente que faz tudo ao mesmo tempo, ele foi mantido enxuto: validar, serializar e publicar.

Outra decisão relevante foi o uso de fila durável e publicação de mensagens persistentes. Isso reduz o risco de perda de dados em reinícios do broker e fortalece o comportamento do sistema em cenários mais próximos de uso real.

O uso de `ACK` manual também foi uma escolha deliberada. Confirmar a mensagem apenas depois da persistência em banco evita o problema clássico de reconhecer cedo demais uma mensagem que ainda não foi efetivamente processada. Embora esse desenho não elimine todos os desafios de sistemas distribuídos, ele melhora bastante a confiabilidade do fluxo atual.

Por fim, a validação precoce no backend reduz a chance de dados inconsistentes chegarem ao broker e ao banco. Isso melhora a qualidade dos dados armazenados e também evita desperdício de processamento assíncrono com mensagens sabidamente inválidas.

## 4. Validações do Payload

Antes de publicar qualquer mensagem na fila, o backend aplica um conjunto de validações para garantir consistência mínima dos dados recebidos. O método HTTP deve ser compatível com a operação esperada, o corpo precisa conter JSON válido e os campos essenciais devem estar presentes.

Os principais campos validados são:

- `device_id`, obrigatório para identificar a origem da leitura
- `timestamp`, obrigatório e esperado em formato RFC3339
- `sensor_type`, obrigatório para caracterizar o tipo de sensor
- `reading_type`, restrito aos valores `analog` ou `discrete`

Essa validação na borda do sistema é importante por dois motivos. Primeiro, impede que mensagens estruturalmente inválidas sejam propagadas para a fila. Segundo, simplifica a responsabilidade do consumidor, que passa a lidar majoritariamente com mensagens já filtradas pelo backend.

## 5. Modelo de Dados

Os dados persistidos são armazenados na tabela `telemetry`, definida da seguinte forma:

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

O modelo foi pensado para armazenar cada leitura de forma simples e objetiva. O campo `device_id` identifica o dispositivo que originou o dado, enquanto `timestamp` registra o instante associado à leitura. Os campos `sensor_type` e `reading_type` ajudam a interpretar semanticamente o evento, e o campo `value` representa o valor medido.

Do ponto de vista de modelagem, essa estrutura atende bem ao escopo atual do projeto, especialmente para sensores analógicos. Para leituras discretas, a representação numérica por convenção, como `0` e `1`, é suficiente neste contexto. Ainda assim, uma melhoria arquitetural futura seria separar explicitamente os valores analógicos e discretos em colunas diferentes, tornando o esquema mais expressivo e reduzindo ambiguidades de interpretação.

Os campos possuem os seguintes papéis:

- `id`: identificador único de cada leitura persistida
- `device_id`: identificação do dispositivo emissor
- `timestamp`: instante associado ao evento recebido
- `sensor_type`: categoria do sensor, como temperatura ou pressão
- `reading_type`: natureza da leitura, analógica ou discreta
- `value`: valor numérico armazenado
- `created_at`: momento em que o registro foi inserido no banco

## 6. Como Executar

### 6.1 Pré-requisitos

Para executar o projeto, é necessário ter:

- Docker Desktop em execução
- Docker Compose habilitado
- k6 instalado, caso o objetivo seja reproduzir os testes de carga

### 6.2 Subir o ambiente

Com os pré-requisitos atendidos, o ambiente pode ser iniciado com:

```bash
docker compose up --build -d
docker compose ps
```

A expectativa é que os serviços `back`, `middleware`, `rabbitmq` e `db` estejam em execução após a inicialização.

### 6.3 Testar o endpoint manualmente

Uma forma simples de validar o fluxo é enviar uma leitura manualmente para a API:

```powershell
$body = @{
  device_id    = "dev-001"
  timestamp    = "2026-03-17T15:00:00Z"
  sensor_type  = "temperature"
  reading_type = "analog"
  value        = 26.7
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/telemetry" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

Se o fluxo estiver correto, a resposta esperada será:

```text
mensagem enfileirada com sucesso
```

### 6.4 Validar persistência

Depois de publicar a mensagem, a persistência pode ser confirmada diretamente no banco:

```bash
docker exec db psql -U postgres -d telemetrydb -c "SELECT * FROM telemetry ORDER BY id DESC LIMIT 5;"
```

### 6.5 Executar testes unitários

Os testes unitários podem ser executados separadamente para backend e middleware:

```bash
cd back
go test
cd ../middleware
go test
```

### 6.6 Executar teste de carga

O cenário de carga pode ser executado com:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js
```

Caso seja necessário registrar também a saída textual, o comando abaixo gera o arquivo de log:

```bash
k6 run --summary-export test-load/reports/summary.json test-load/telemetry.js > test-load/reports/run.log
```

### 6.7 Executar benchmark reproduzivel (CPU/RAM limitados)

Para manter o ambiente de desenvolvimento intacto, o benchmark usa o arquivo `docker-compose.bench.yml` como override de recursos e o script `test-load/run-benchmark.ps1` para automatizar as rodadas.

Executar benchmark com os perfis padrao (0.25, 0.5 e 1.0 core), 3 rodadas por perfil:

```bash
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
./test-load/run-benchmark.ps1
```

Executar benchmark com parametros customizados:

```bash
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
./test-load/run-benchmark.ps1 -Runs 5 -Duration 60s -Vus 40
```

O script gera artefatos em `test-load/reports/benchmarks/<timestamp>/` e um ponteiro em `test-load/reports/benchmarks/latest.txt`.

### 6.8 Encerrar o ambiente

Para interromper os containers:

```bash
docker compose down
```

Para remover também os volumes e realizar um reset completo:

```bash
docker compose down -v
```

## 7. Validação do Sistema

A validação prática do ambiente foi conduzida com uma combinação de verificações operacionais e funcionais. Primeiro, `docker compose ps` foi utilizado para confirmar que os quatro serviços estavam ativos. Em seguida, o endpoint `/telemetry` foi acionado manualmente para garantir o recebimento da requisição pelo backend.

Depois disso, a confirmação do fluxo interno foi feita observando a fila no RabbitMQ, os logs do middleware e a presença dos registros no PostgreSQL. Esse conjunto de evidências mostrou que a mensagem percorre corretamente todas as etapas do pipeline, desde a entrada HTTP até a persistência final.

Em termos de resultado, o comportamento observado confirma que o fluxo ponta a ponta está operacional: o `back` recebe a telemetria, o RabbitMQ desacopla as etapas, o `middleware` consome a mensagem e o PostgreSQL armazena o dado com sucesso.

Checklist de validação utilizado:

- verificar os containers com `docker compose ps`
- enviar uma requisição manual para `/telemetry`
- observar a fila no RabbitMQ
- acompanhar os logs do `middleware`
- consultar os registros persistidos no PostgreSQL

## 8. Testes Unitários

Os testes unitários foram distribuídos entre os dois principais componentes de aplicação.

No backend, os cenários exercitados cobrem tanto o caminho feliz quanto falhas de validação:

- aceitação de requisições válidas com retorno HTTP `202`
- rejeição de JSON inválido com HTTP `400`
- bloqueio de métodos HTTP indevidos com HTTP `405`
- identificação de `timestamp` inválido
- validação restritiva do campo `reading_type`

No middleware, os testes verificam o comportamento do consumidor diante de diferentes tipos de mensagem e falha:

- mensagem válida com persistência bem-sucedida
- erros de parse de JSON
- falhas na conversão de `timestamp`
- propagação correta de falhas de banco

Esses testes são importantes porque cobrem a etapa que, do ponto de vista arquitetural, concentra a confiabilidade do processamento assíncrono.

## 9. Teste de Carga com k6

### 9.1 Configuração do cenário

O teste de carga foi executado a partir do script `test-load/telemetry.js`, com três estágios progressivos e duração total ativa de 40 segundos. O pico do cenário chegou a 30 usuários virtuais simultâneos, todos exercitando o endpoint `POST /telemetry`.

Esse desenho foi suficiente para avaliar o comportamento da API em um contexto local conteinerizado, observando tanto estabilidade de resposta quanto a capacidade de absorver requisições consecutivas sem erro.

### 9.2 Resultados obtidos

Com base no arquivo `test-load/reports/summary.json`, os resultados mais recentes foram:

| Métrica | Resultado |
| --- | --- |
| Requisições realizadas | 605 |
| Checks aprovados | 605 |
| Taxa de sucesso | 100% |
| Falhas HTTP | 0% |
| Throughput | 15.02 req/s |
| Latência média | 3.09 ms |
| Mediana | 2.37 ms |
| p90 | 4.99 ms |
| p95 | 6.28 ms |
| Latência máxima | 121.10 ms |

### 9.3 Análise técnica dos resultados

Os resultados indicam que o endpoint permaneceu estável ao longo de todo o cenário executado. A ausência de falhas HTTP e de checks quebrados mostra que o sistema conseguiu manter consistência funcional mesmo sob carga concorrente.

Do ponto de vista de desempenho, a latência média de 3.09 ms é baixa para o ambiente adotado, o que sugere que o caminho síncrono da API está enxuto e bem alinhado com a proposta arquitetural. A mediana também se manteve reduzida, indicando comportamento consistente na maior parte das requisições.

O valor máximo de latência, de 121.10 ms, precisa ser interpretado com cuidado. Em ambiente local conteinerizado, picos pontuais como esse podem ocorrer por disputa de CPU, sincronização de I/O, aquecimento de processo e variações naturais do host. Como esses picos não vieram acompanhados de falhas ou degradação geral das métricas, a leitura mais razoável é que se tratam de outliers operacionais, e não de um gargalo sistêmico do endpoint.

Mais importante do que o pico isolado é o comportamento agregado: a API manteve taxa de sucesso total enquanto a persistência permaneceu desacoplada da resposta ao cliente. Isso reforça que a decisão de usar mensageria entre backend e banco foi tecnicamente acertada para o escopo da solução.

### 9.4 Evidências no repositório

As evidências relacionadas ao teste de carga podem ser consultadas nos seguintes arquivos:

- `test-load/reports/summary.json`
- `test-load/reports/run.log`
- `test-load/reports/README.md`

### 9.5 Benchmark com limite de CPU e extrapolacao para 1 core

Foi adicionado um fluxo de benchmark para reproduzir resultados com limite de CPU e memoria em todos os containers (`back`, `middleware`, `rabbitmq`, `db`).

Perfis executados:

- `0.25` core
- `0.5` core
- `1.0` core

Metodologia:

- 3 rodadas por perfil
- consolidacao automatica de `req/s`, `p95` e `fail_rate`
- comparacao da estimativa por regra de 3 contra o valor medido real de 1 core

Formula aplicada na extrapolacao:

```text
req_s_estimado_1core = req_s_medido / cpu_perfil
erro_percentual = ((estimado - real_1core) / real_1core) * 100
```

Arquivos gerados por execucao:

- `benchmark-summary.md`: tabela pronta para relatorio
- `benchmark-summary.json`: consolidacao estruturada
- `benchmark-summary.csv`: consolidacao tabular
- `raw-runs.json`: resultados de cada rodada

### 9.6 Resultado do benchmark final (execucao 20260323-000757)

Fonte: `test-load/reports/benchmarks/20260323-000757/benchmark-summary.md`

| CPU (core) | req/s medio | req/s mediana | p95 medio (ms) | fail_rate medio | estimado 1 core | 1 core real | erro extrapolacao (%) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0.25 | 29.7870 | 29.7892 | 9.6314 | 0 | 119.1482 | 29.7845 | 300.03 |
| 0.5 | 29.7872 | 29.8160 | 10.8853 | 0 | 59.5745 | 29.7845 | 100.02 |
| 1.0 | 29.7845 | 29.7933 | 11.6490 | 0 | - | - | - |

Leitura tecnica do resultado:

- o endpoint ficou estavel em todas as rodadas (fail_rate medio 0 em todos os perfis)
- a extrapolacao por regra de 3 superestimou bastante o 1 core real neste experimento
- isso indica que o throughput ficou limitado por fatores do proprio cenario de carga (como `sleep(1)` por iteracao no script), e nao apenas por CPU

### 9.7 Analise quantitativa do benchmark final

Com base na tabela consolidada:

- a diferenca entre o maior e o menor throughput medio foi de apenas 0.0027 req/s (29.7872 - 29.7845), indicando variacao muito pequena entre os perfis testados
- a latencia p95 variou de 9.6314 ms a 11.6490 ms, oscilacao esperada em ambiente local conteinerizado, sem impacto em estabilidade funcional
- a extrapolacao linear apresentou erro alto versus 1 core real: 300.03% (perfil 0.25) e 100.02% (perfil 0.5)

Conclusao tecnica desta bateria:

- o teste validou estabilidade e reproducibilidade operacional da stack sob limites de recurso
- este cenario especifico nao foi adequado para inferir escalabilidade linear por CPU, pois o desenho da carga impôs gargalo adicional no throughput

## 10. Melhorias Futuras

O projeto já atende bem ao objetivo da atividade, mas há um conjunto claro de melhorias que aumentariam sua maturidade técnica.

- adicionar `healthcheck` e `depends_on.condition: service_healthy` para reduzir falhas de inicialização entre serviços
- implementar retry com backoff exponencial nas conexões com RabbitMQ e PostgreSQL
- criar uma dead-letter queue para separar mensagens com erro permanente
- incluir autenticação e autorização no endpoint de ingestão
- adicionar observabilidade com métricas, tracing e dashboards
- separar leituras analógicas e discretas de forma mais explícita no schema
- ampliar a cobertura com testes de integração fim a fim
- versionar contratos de payload para evoluir o protocolo com mais segurança
- executar nova bateria de benchmark removendo gargalo de think time no k6 (reduzir/remover `sleep(1)`) para avaliar escalabilidade por CPU com maior fidelidade

## 11. Conclusão

O sistema desenvolvido atende ao objetivo proposto ao implementar uma pipeline assíncrona e desacoplada para processamento de telemetria. A combinação entre backend HTTP, RabbitMQ, middleware consumidor e PostgreSQL forma uma arquitetura coerente com os requisitos de escalabilidade básica, resiliência e organização de responsabilidades.

Os testes funcionais e os testes unitários demonstram que o fluxo principal está correto. Além disso, o benchmark reproduzível com limites de CPU/RAM em todos os containers confirmou estabilidade total em 3 rodadas por perfil (0.25, 0.5 e 1.0 core), mantendo `fail_rate` médio igual a 0 e throughput próximo de 29.78 req/s no cenário adotado.

Na análise de extrapolação para 1 core, a regra de 3 superestimou o valor real medido de 1.0 core (erros de 300.03% para 0.25 core e 100.02% para 0.5 core). Esse resultado evidencia que, neste experimento, o throughput foi fortemente influenciado pelo próprio desenho do teste (como `sleep(1)` por iteração) e não apenas por CPU disponível. Ainda assim, a execução validou o objetivo metodológico da atividade: comparar perfis limitados, medir com rastreabilidade e confrontar a extrapolação com dado real.

Em síntese, o projeto não apenas atende aos requisitos da atividade, mas também estabelece uma fundação consistente para evoluções futuras em confiabilidade, observabilidade, segurança e maturidade operacional.
