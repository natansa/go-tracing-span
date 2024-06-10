## Testes do projeto
- Execute a aplicacao com docker-compose: docker-compose up -d
- Os testes unitarios estao no arquivo: ./tests/unit_test.go
- O teste manual localhost pode ser realizado usando o Send Request no arquivo: ./api/get-local.http
- O teste manual cloudrun pode ser realizado usando o Send Request no arquivo: ./api/get-cloudrun.http
- O teste de carga local automatizado pode ser realizado executando o run test do arquivo: ./tests/load_test_local.go
- O teste de carga cloudrun automatizado pode ser realizado executando o run test do arquivo: ./tests/load_test_cloudrun.go