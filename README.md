# Esta documentação fornece um guia passo a passo para configurar e rodar o projeto go-tracing-span em um ambiente de desenvolvimento.

## Pré-requisitos
Antes de iniciar, certifique-se de ter as seguintes ferramentas instaladas:
- Docker
- Docker Compose
- Git

## Passo 1: Clonar o Repositório. Clone o repositório go-tracing-span do GitHub.
- git clone https://github.com/natansa/go-tracing-span.git
- cd go-tracing-span

## Passo 2: Estrutura do Projeto
- serviceA/ e serviceB/: Contêm o código e Dockerfiles dos serviços A e B.
- otel-collector-config.yaml: Configuração do OpenTelemetry Collector.
- prometheus.yaml: Configuração do Prometheus.
- docker-compose.yml: Configuração do Docker Compose para orquestrar os serviços.

## Passo 3: Construir e Rodar os Serviços. Execute os seguintes comandos para construir e iniciar os serviços usando Docker Compose.
- docker-compose up --build

Esse comando construirá as imagens Docker para serviceA e serviceB, e iniciará todos os serviços definidos no docker-compose.yml.
Você pode verificar os logs dos serviços para garantir que estão rodando corretamente.
- docker-compose logs servicea
- docker-compose logs serviceb
- docker-compose logs otel-collector
- docker-compose logs prometheus
- docker-compose logs zipkin

## Passo 4: Testar o Serviço. Para testar a configuração, você pode fazer uma requisição HTTP POST para serviceA.
- curl -X POST http://localhost:8080/weather -d '{"cep": "29902555"}' -H "Content-Type: application/json"

## Passo 5: Acessar Zipkin
- Acesse http://localhost:9411 para visualizar os traces.
