# Go Client-Server API

> [!IMPORTANT]  
> Para poder executar os projetos contidos neste repositório é necessário que se tenha o Go instalado no computador. Para maiores informações siga o site https://go.dev/

### Desafio GoLang Pós GoExpert - Client-Server API

Este projeto faz parte como desafio da Pós GoExpert, nele são cobertos os conhecimentos em http webserver, contextos, banco de dados e também manipulação de aquivos.

O Desafio consiste em entregar dois sistemas em Go, sendo eles:
- client.go
- server.go

Dos quais devem obdecer os seguintes requisitos:
- O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
- O server.go deverá consumir uma API contendo o câmbio de Dólar e Real e em seguida deverá retornar no formato JSON o resultado para o cliente. Abaixo segue o link da API a ser consumida:

    ```
    https://economia.awesomeapi.com.br/json/last/USD-BRL 
    ```

- Usando o package `context`, o server.go deverá registrar no banco de dados _SQLite_ cada cotação recebida, sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de **200ms** e o timeout máximo para conseguir persistir os dados no banco deverá ser de **10ms**.

- O client.go precisará receber do server.go apenas o valor atual do câmbio (campo `bid` do JSON). Utilizando o package `context`, o client.go terá um timeout máximo de **300ms** para receber o resultado do server.go.

- Os 3 contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.

- O client.go terá que salvar a cotação atual em um arquivo `cotacao.txt` no formato: Dólar: {valor}

- O endpoint necessário gerado pelo server.go para este desafio será: `/cotacao` e a porta a ser utilizada pelo servidor HTTP será a **8080**.

### Executando os sistemas
#### Server

Para poder executar o `server` você precisa estar dentro da pasta `server` e executar o seguinte comando abaixo a partir do terminal:
```shell
❯ go run server.go
```
Na janela do terminal você deverá ver uma mensagem parecida com o exemplo abaixo:
```shell
❯ go run server.go
2024/07/05 01:46:07 Server listening on localhost:8080
```
> Esta mensagem informa que o servidor já encontra-se disponível.

#### Client

Para poder executar o `client` você precisa estar dentro da pasta `client` e executar o seguinte comando abaixo a partir do terminal:
```shell
❯ go run client.go
```
Na janela do terminal você deverá ver uma mensagem parecida com o exemplo abaixo:
```shell
❯ go run main.go
2024/07/05 01:47:41 Sending request :: [GET] - localhost:8080/cotacao
2024/07/05 01:47:41 Log file created :: cotacao.txt - (14 bytes)
```
> O exemplo de mensagem acima, apresenta `logs` informando os momentos em que a requisição é disparada e em que o aquivo `cotacao.txt` é criado.

#### Endpoint /health
De modo a aprofundar mais os estudos desenvolvi um `endpoint` para exemplificar um `health-check`, neste endpoint ao acionarmos receberemos um JSON como resposta contendo as seguintes informações:
```json
{
"cpu": {
    "cores": 10,
    "percent_used": [
        12.418727008294281,
        10.632701093725466,
        7.77921444771416,
        5.79172610556409,
        6.34332287344979,
        4.237920560011274,
        2.494750104997976,
        6.14680347276916,
        4.1662288536303205,
        2.457975031761513
    ]
},
"duration": "266.542µs",
"memory": {
    "available": 5113233408,
    "free": 209010688,
    "percent_used": 70.23706436157227,
    "total": 17179869184,
    "used": 12066635776
},
"status": "pass",
"uptime": "16m1.870507292s"
}
```