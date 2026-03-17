# RabbitMQ

Este módulo é responsável pela mensageria da aplicação.

O RabbitMQ é utilizado como broker de mensagens, permitindo o desacoplamento entre o backend e o middleware.

## Função no sistema

- Receber mensagens do backend
- Armazenar mensagens em filas
- Disponibilizar mensagens para consumo pelo middleware

## Configuração

O serviço é executado via Docker utilizando a imagem oficial:

rabbitmq:3-management

Portas expostas:
- 5672 (AMQP)
- 15672 (interface web)