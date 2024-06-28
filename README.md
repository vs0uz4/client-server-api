# Go Client-Server API
### Desafio GoLang Pós GoExpert - Client-Server API

> [!IMPORTANT]  
> Para poder executar os projetos contidos neste repositório é necessário que se tenha o Go instalado no computador. Para maiores informações siga o site https://go.dev/

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
