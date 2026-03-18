# Sistema de Processamento de Telemetria com Arquitetura Assíncrona

## 1. Introdução

Este projeto consiste no desenvolvimento de um sistema distribuído para ingestão e processamento de dados de telemetria, utilizando uma arquitetura baseada em mensageria. O objetivo principal foi construir um pipeline completo de dados, desde a recepção via API até a persistência em banco de dados, garantindo desacoplamento entre os componentes, escalabilidade e robustez.

A solução foi implementada utilizando a linguagem Go, com suporte de ferramentas amplamente utilizadas no mercado, como RabbitMQ para mensageria, PostgreSQL para persistência, Docker para containerização e k6 para testes de carga.

## 2. Objetivos

Os principais objetivos do projeto foram:

* Implementar uma API capaz de receber dados estruturados de telemetria
* Garantir o desacoplamento entre ingestão e processamento utilizando filas
* Persistir os dados de forma confiável em banco relacional
* Validar a aplicação por meio de testes unitários
* Avaliar o comportamento do sistema sob carga


## 3. Arquitetura do Sistema

A arquitetura adotada segue o padrão de processamento assíncrono orientado a eventos.

Fluxo geral:

Cliente → Backend → RabbitMQ → Middleware → PostgreSQL

### 3.1 Componentes

* **Backend (Go)**: responsável por expor um endpoint HTTP e publicar mensagens na fila
* **RabbitMQ**: responsável por intermediar a comunicação entre os serviços
* **Middleware (Go)**: responsável por consumir mensagens da fila e realizar a persistência
* **PostgreSQL**: responsável pelo armazenamento dos dados
* **Docker**: responsável por orquestrar os containers

## 4. Fluxo de Execução do Sistema

O funcionamento do sistema ocorre em etapas bem definidas, conforme descrito a seguir.

### 4.1 Recepção da requisição

O cliente realiza uma requisição HTTP POST para o endpoint `/telemetry`, enviando um payload JSON contendo os dados de telemetria, incluindo identificador do dispositivo, timestamp, tipo de sensor e valor medido.

### 4.2 Processamento no backend

Ao receber a requisição, o backend realiza:

* Validação do método HTTP
* Desserialização do JSON recebido
* Serialização do payload em formato apropriado para envio
* Publicação da mensagem na fila do RabbitMQ

Nesse momento, o backend não realiza nenhuma operação de persistência, garantindo resposta rápida ao cliente.

### 4.3 Enfileiramento da mensagem

A mensagem é armazenada na fila `telemetry_queue` no RabbitMQ. Esse componente atua como buffer, permitindo que o sistema suporte picos de requisições sem sobrecarregar o banco de dados.

### 4.4 Consumo pelo middleware

O middleware mantém uma conexão ativa com o RabbitMQ e consome mensagens continuamente. Para cada mensagem recebida:

* O payload é desserializado
* O timestamp é convertido para formato compatível com o banco
* Os dados são validados

### 4.5 Persistência no banco de dados

Após validação, o middleware executa um comando SQL de inserção na tabela `telemetry`, armazenando os dados de forma estruturada.

## 5. Modelagem do Banco de Dados

A tabela utilizada para armazenamento foi definida da seguinte forma:

```sql
CREATE TABLE telemetry (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(50),
    timestamp TIMESTAMP,
    sensor_type VARCHAR(50),
    reading_type VARCHAR(50),
    value FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 6. Containerização e Orquestração

O sistema foi containerizado utilizando Docker Compose, permitindo a execução simultânea dos serviços.

Os containers definidos foram:

* backend
* middleware
* rabbitmq
* postgres

Essa abordagem garante:

* isolamento entre os serviços
* reprodutibilidade do ambiente
* facilidade de execução em diferentes máquinas

## 6.1 Decisão de Organização da Infraestrutura

A organização dos arquivos de infraestrutura do projeto foi realizada de forma a manter proximidade com o componente responsável pelo processamento dos dados, neste caso, o middleware.

Embora os serviços de RabbitMQ e PostgreSQL sejam definidos no `docker-compose.yml` como containers independentes, seus arquivos auxiliares (como scripts de inicialização e documentação) foram organizados em diretórios relacionados ao middleware.

Essa decisão foi motivada pelos seguintes fatores:

* O middleware é o principal consumidor das mensagens e responsável pela persistência dos dados
* Tanto o RabbitMQ quanto o PostgreSQL são utilizados diretamente pelo middleware durante o processamento
* A organização próxima ao middleware facilita a compreensão do fluxo de dados, especialmente em projetos de menor escala
* Reduz a fragmentação do projeto, evitando a criação de múltiplos diretórios de infraestrutura sem necessidade

É importante destacar que, do ponto de vista arquitetural, RabbitMQ e PostgreSQL continuam sendo serviços independentes, executados em containers próprios, garantindo isolamento e desacoplamento entre os componentes.

Essa organização foi adotada visando simplicidade, clareza e facilidade de manutenção, sem comprometer os princípios da arquitetura distribuída.

## 7. Como Executar o Projeto

Para executar o sistema localmente, é necessário possuir o Docker instalado e em execução na máquina.

### 7.1 Subindo o ambiente

Na raiz do projeto, execute:

```bash
docker compose up --build
```

Esse comando irá:

* Construir as imagens
* Subir todos os containers
* Inicializar o banco de dados
* Iniciar backend e middleware

### 7.2 Testando o endpoint manualmente

Exemplo via PowerShell:

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

### 7.3 Verificando a persistência no banco

```bash
docker exec -it db psql -U postgres -d telemetrydb
```

```sql
SELECT * FROM telemetry;
```

### 7.4 Executando os testes unitários

#### Backend

```bash
cd back
go test
```

#### Middleware

```bash
cd middleware
go test
```

### 7.5 Executando o teste de carga

```bash
k6 run test-load/telemetry.js
```

### 7.6 Encerrando o ambiente

```bash
docker compose down
```

Reset completo:

```bash
docker compose down -v
```

## 8. Testes Unitários

### 8.1 Testes do Backend

Foi utilizada uma abordagem baseada em mocks para desacoplar a dependência do RabbitMQ.

Foi criada uma interface de publicação (`Publisher`), permitindo a substituição por uma implementação mock durante os testes.

Cenários testados:

* Requisição válida (HTTP 202)
* JSON inválido (HTTP 400)
* Método inválido (HTTP 405)

### 8.2 Testes do Middleware

A lógica de processamento foi isolada, permitindo testes independentes da infraestrutura.

Cenários testados:

* Mensagem válida
* JSON inválido
* Timestamp inválido
* Erro no banco

## 9. Teste de Carga

### 9.1 Configuração

* Até 30 usuários simultâneos
* Duração: 40 segundos
* Endpoint: `/telemetry`

### 9.2 Resultados

* Total de requisições: 605
* Taxa de sucesso: 100%
* Falhas: 0%

### 9.3 Métricas

* Latência média: 3.01 ms
* p95: 5.96 ms
* Latência máxima: 10.54 ms

### 9.4 Análise

Os resultados demonstram que o sistema apresenta:

* Alta capacidade de resposta
* Baixa latência
* Estabilidade sob carga
* Nenhuma perda de requisição

O uso de mensageria contribuiu significativamente para esse desempenho, pois o backend atua apenas como produtor de mensagens, delegando o processamento e a persistência ao middleware. Dessa forma, operações de I/O não impactam diretamente o tempo de resposta da API.

## 10. Validação do Sistema

A validação foi realizada por meio de:

* Verificação com `docker ps`
* Análise de logs
* Testes manuais via API
* Consulta ao banco
* Testes unitários
* Teste de carga

Isso garantiu o funcionamento completo do pipeline de ponta a ponta.

## 11. Conclusão

O sistema desenvolvido demonstra, na prática, a aplicação de uma arquitetura assíncrona baseada em mensageria para o processamento eficiente de dados de telemetria. A separação entre ingestão e processamento permitiu reduzir o acoplamento entre os componentes, garantindo maior resiliência e capacidade de resposta mesmo sob cenários de carga concorrente. Os testes unitários realizados no backend e no middleware apresentaram sucesso em todos os cenários avaliados, validando a corretude das regras de negócio e o tratamento adequado de erros. Adicionalmente, o teste de carga evidenciou o bom desempenho da aplicação, com 605 requisições processadas, 100% de taxa de sucesso e latência média de aproximadamente 3 ms, mantendo estabilidade ao longo da execução. Esses resultados confirmam que a solução atende aos objetivos propostos, apresentando comportamento consistente, baixo tempo de resposta e capacidade de escalabilidade, sendo adequada para cenários reais de monitoramento com alto volume de dados.




