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

Essa abordagem garante isolamento entre os serviços e facilidade de execução em diferentes ambientes.

## 7. Testes Unitários

### 7.1 Testes do Backend

Para o backend, foi utilizada uma abordagem baseada em mocks para desacoplar a dependência do RabbitMQ.

Foi criada uma interface de publicação (`Publisher`), permitindo a substituição por uma implementação mock durante os testes.

Os seguintes cenários foram testados:

* Requisição válida: valida se o endpoint retorna status HTTP 202
* JSON inválido: valida retorno de erro HTTP 400
* Método inválido: valida retorno de erro HTTP 405

Esses testes garantem que a API responde corretamente a diferentes tipos de entrada, além de validar o comportamento esperado do endpoint.

### 7.2 Testes do Middleware

Para possibilitar testes unitários no middleware, a lógica de processamento foi isolada em uma função específica responsável por tratar cada mensagem recebida.

Foi criada uma interface para o repositório de banco de dados, permitindo a utilização de um mock durante os testes.

Os cenários testados foram:

* Mensagem válida: valida se os dados são corretamente processados e enviados ao repositório
* JSON inválido: garante que erros de parsing são tratados corretamente
* Timestamp inválido: valida o tratamento de erro na conversão de data
* Erro no banco: simula falha na persistência e valida o comportamento do sistema

Essa abordagem permite testar a lógica de negócio sem depender de RabbitMQ ou PostgreSQL reais.

## 8. Teste de Carga

O teste de carga foi realizado utilizando a ferramenta k6, com o objetivo de avaliar o comportamento do sistema sob múltiplas requisições simultâneas.

### 8.1 Configuração do teste

* Usuários virtuais (VUs): até 30 simultâneos
* Duração: 40 segundos
* Tipo de requisição: POST para `/telemetry`

### 8.2 Resultados obtidos

* Total de requisições: 605
* Taxa de sucesso: 100%
* Falhas: 0%

### 8.3 Métricas de desempenho

* Latência média: 3.01 ms
* Percentil 95 (p95): 5.96 ms
* Latência máxima: 10.54 ms

### 8.4 Análise dos resultados

Os resultados indicam que o sistema apresenta:

* Alta capacidade de resposta
* Baixa latência
* Estabilidade sob carga
* Nenhuma perda de requisição

O uso de mensageria contribuiu significativamente para esse desempenho, pois o backend apenas enfileira mensagens, evitando bloqueios relacionados à persistência.

## 9. Validação do Sistema

A validação do sistema foi realizada por meio de múltiplas estratégias:

* Verificação do estado dos containers com `docker ps`
* Análise dos logs dos serviços
* Execução de requisições manuais via API
* Consulta direta ao banco de dados
* Execução de testes automatizados (unitários e de carga)

Essa abordagem garantiu que todo o pipeline estava funcionando corretamente de ponta a ponta.

## 10. Conclusão

O sistema desenvolvido demonstra uma arquitetura moderna baseada em comunicação assíncrona, capaz de lidar com múltiplas requisições de forma eficiente e escalável.

A separação entre ingestão e processamento permite maior flexibilidade e resiliência, enquanto a utilização de testes unitários e de carga assegura a qualidade e confiabilidade da solução.

Os resultados obtidos indicam que o sistema atende plenamente aos objetivos propostos, apresentando desempenho consistente e comportamento estável.


