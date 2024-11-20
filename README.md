# fullcycle-rate-limiter

Este projeto é um **Rate Limiter** desenvolvido em **Go**, projetado para controlar o tráfego de requisições para um serviço web com base em dois critérios: 
- **Endereço IP**
- **Token de acesso**

O objetivo principal é limitar o número de requisições permitidas em um determinado intervalo de tempo, protegendo a aplicação contra abusos.

## Recursos

- **Limitação por IP**: Permite configurar o número máximo de requisições permitidas por segundo para cada endereço IP.
- **Limitação por Token**: Baseia-se no cabeçalho `API_KEY` para diferenciar os limites por token. Os limites por token têm prioridade sobre os limites por IP.
- **Middleware**: O rate limiter é implementado como um middleware que pode ser facilmente integrado ao servidor web.
- **Persistência com Redis**: Utiliza o Redis para armazenar dados do rate limiter.
- **Configuração flexível**: Parâmetros podem ser ajustados via variáveis de ambiente ou arquivo `.env`.
- **Substituição de persistência**: Implementação baseada em "strategy", permitindo fácil substituição do Redis por outro mecanismo.

---

## Como Funciona o Rate Limiter

### Fluxo de Funcionamento

1. **Requisição recebida**:
   - O middleware identifica o cliente por meio do endereço IP ou do token presente no cabeçalho `API_KEY`.

2. **Validação do Limite**:
   - Verifica no Redis se o cliente já atingiu o limite configurado para aquele critério (IP ou token).
   - Se o limite for excedido, retorna o código `429 - Too Many Requests` com a mensagem:
     ```
     you have reached the maximum number of requests or actions allowed within a certain time frame
     ```

3. **Contador de Requisições**:
   - Caso o limite não tenha sido excedido, o contador é incrementado e a requisição é processada normalmente.

4. **Bloqueio Temporário**:
   - Quando o cliente ultrapassa o limite, ele é bloqueado por um período configurável (`BAN_DURATION`).

### Configuração do Middleware

As principais variáveis que controlam o comportamento do rate limiter são:

| Variável             | Descrição                                                                 | Exemplo         |
|----------------------|---------------------------------------------------------------------------|-----------------|
| `SERVER_PORT`        | Porta em que o servidor será iniciado.                                   | `8080`          |
| `REDIS_ADDR`         | Endereço do Redis.                                                       | `localhost:6379`|
| `RATE_LIMIT_IP`      | Número máximo de requisições permitidas por segundo para cada IP.        | `5`             |
| `RATE_LIMIT_TOKEN`   | Número máximo de requisições permitidas por segundo para cada token.     | `10`            |
| `BAN_DURATION`       | Tempo de bloqueio (em segundos) após exceder o limite de requisições.   | `60`            |

---

## Configuração do Projeto

### Pré-requisitos

- **Go 1.20+**
- **Docker e Docker Compose**
- **Redis** (já incluído no `docker-compose.yml`)

### Passo a Passo

1. Clone o repositório:
   ```bash
   git clone https://github.com/marcosocram/fullcycle-rate-limiter.git
   cd fullcycle-rate-limiter
   ```
2. Configure o arquivo `.env`: Crie um arquivo `.env` na raiz do projeto e defina as variáveis:
    ```env
    SERVER_PORT=8080
    REDIS_ADDR=redis:6379
    RATE_LIMIT_IP=5
    RATE_LIMIT_TOKEN=10
    BAN_DURATION=60
    ```
3. Construa e inicie os serviços com Docker:
   ```bash
   docker-compose up --build
   ```
4. O servidor estará disponível em: http://localhost:8080

## Cenários de Teste

### Cenário 1: Requisições dentro do limite por IP

Se o limite é de 5 requisições por segundo por IP e fazemos 5 requisições rápidas do mesmo IP, o serviço deve responder normalmente para cada uma.

#### Comando para testar:

```bash
for i in {1..5}; do curl -i http://localhost:8080/; done
```

#### Resultado esperado:

```
HTTP/1.1 200 OK
Content-Length: 42
Content-Type: text/plain; charset=utf-8

Bem-vindo ao servidor com Rate Limiting!
...
(repetido 5 vezes)
```

### Cenário 2: Excedendo o limite por IP

Agora, fazemos 6 requisições rápidas do mesmo IP, onde o limite é de 5 por segundo. A sexta requisição deve ser bloqueada com o código HTTP 429 e a mensagem de erro.

#### Comando para testar:

```bash
for i in {1..6}; do curl -i http://localhost:8080/; done
```

#### Resultado esperado:

```
# Primeiras 5 requisições
HTTP/1.1 200 OK
Content-Length: 42
Content-Type: text/plain; charset=utf-8

Bem-vindo ao servidor com Rate Limiting!
...
(repetido 5 vezes)

# Sexta requisição
HTTP/1.1 429 Too Many Requests
Content-Length: 95
Content-Type: text/plain; charset=utf-8

you have reached the maximum number of requests or actions allowed within a certain time frame
```
Após essa resposta de **429**, o IP é bloqueado por 60 segundos, como definido em `BAN_DURATION`. Novas requisições do mesmo IP durante esse período retornarão o código **HTTP 429** com a mesma mensagem de erro.

### Cenário 3: Requisições com Token dentro do limite

Agora vamos fazer 10 requisições rápidas utilizando um token `API_KEY: abc123` que tem limite de 10 req/s.

#### Comando para testar:
    
```bash
for i in {1..10}; do curl -i -H "API_KEY: abc123" http://localhost:8080/; done
```

#### Resultado esperado:

```
HTTP/1.1 200 OK
Content-Length: 42
Content-Type: text/plain; charset=utf-8

Bem-vindo ao servidor com Rate Limiting!
...
(repetido 10 vezes)
```

### Cenário 4: Excedendo o limite por Token

Agora fazemos 11 requisições com o mesmo token `API_KEY: abc123` em menos de um segundo. A 11ª requisição será bloqueada.

#### Comando para testar:

```bash
for i in {1..11}; do curl -i -H "API_KEY: abc123" http://localhost:8080/; done
```

#### Resultado esperado:

```
# Primeiras 10 requisições
HTTP/1.1 200 OK
Content-Length: 42
Content-Type: text/plain; charset=utf-8

Bem-vindo ao servidor com Rate Limiting!
...
(repetido 10 vezes)

# Décima primeira requisição
HTTP/1.1 429 Too Many Requests
Content-Length: 95
Content-Type: text/plain; charset=utf-8

you have reached the maximum number of requests or actions allowed within a certain time frame
```

Assim como no caso do IP, o token ficará bloqueado por 60 segundos (ou o valor definido em `BAN_DURATION`). Novas requisições com esse token durante o período de bloqueio retornarão o código **HTTP 429**.

## Testes Automatizados

Este projeto inclui testes automatizados para validar os cenários do rate limiter:

1. Instale as dependências de teste:
    ```bash
    go mod tidy
    ```
2. Execute os testes:
    ```bash
    go test ./tests -v
    ```

Os testes utilizam a biblioteca miniredis para simular um servidor Redis em memória, garantindo isolamento entre os testes e independência de serviços externos.

Os resultados mostrarão se cada cenário passou ou falhou, demonstrando a eficácia e a robustez do rate limiter.

Estes testes cobrem os cenários principais e verificam se o rate limiter funciona conforme esperado em diferentes condições.