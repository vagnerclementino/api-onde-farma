# Prompt para IA: Estrutura de Projeto Go para AWS Lambda

Atue como um desenvolvedor Go sênior especializado em arquitetura de software. Crie a estrutura de diretórios e o código base de um projeto em Go que rodará como uma AWS Lambda (integrada ao API Gateway) para emular uma API REST.

Aplique rigorosamente as seguintes boas práticas e padrões arquiteturais:

1. **Screaming Architecture e Pacotes de Domínio**: A estrutura de diretórios deve "gritar" a regra de negócio do projeto (com pastas como `book` ou `user`), separando claramente o domínio das regras de negócio de pacotes auxiliares ou de suporte (como `config` ou `auth`).
2. **Diretório `cmd`**: Utilize a pasta `cmd` (por exemplo, `cmd/api/main.go`) para abrigar o ponto de entrada e a compilação do executável. É na função `main` que deve ocorrer toda a injeção de dependências, instanciando repositórios, serviços, handlers e o controle de contexto.
3. **Pacote `internal`**: Proteja os pacotes e implementações que não devem ser expostos publicamente ou importados por outros projetos, como os handlers HTTP e rotas, colocando-os dentro do diretório `internal`.
4. **Domínio Puro (Sem Tags)**: As structs que representam as entidades de negócio devem ser mantidas puras e isoladas, ou seja, não devem conter tags de JSON, banco de dados ou ferramentas de ORM. O domínio não pode depender da camada de transporte.
5. **Camada HTTP e "De-Para"**: Na camada que processa o request (os handlers HTTP), crie structs específicas de entrada e saída (como `bookRequest` e `bookResponse`) com as devidas tags JSON. Faça o *parse*, a validação e a conversão de dados ("De-Para") nesta camada, antes de enviar os dados para a struct de domínio puro no serviço. Utilize bibliotecas como `chi` ou a biblioteca padrão para o roteamento.
6. **Uso de `context.Context`**: Por convenção, o primeiro parâmetro de todas as funções da camada de serviço e de repositório deve ser um `context.Context` (para lidar com timeouts e graceful shutdown), e o último valor retornado deve ser um `error`.
7. **Segregação de Interfaces**: Defina interfaces para os Casos de Uso e Repositórios. Siga a boa prática de criar interfaces pequenas e específicas (como `Reader` para leitura e `Writer` para gravação) e faça a composição delas na interface principal do repositório.
8. **Retorno de Estruturas Concretas**: As funções construtoras dos serviços devem sempre receber interfaces via parâmetro (o que facilita a criação de *mocks* para testes), mas devem retornar tipos concretos (como um ponteiro para a struct `*service`), garantindo maior performance e segurança no uso da estrutura.

Com base nessas regras, gere a árvore de diretórios do projeto e os arquivos principais com exemplos de código para um CRUD simples.
