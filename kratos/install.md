```
<!-- docker run -d --rm -it   -p 4433:4433   -p 4434:4434   -v $(pwd)/kratos.yaml:/etc/config/kratos.yml   -v $(pwd)/identity.schema.json:/etc/config/identity.schema.json   oryd/kratos:latest serve -c /etc/config/kratos.yml

docker run -d -p 3000:3000 oryd/kratos-selfservice-ui-node -->
  
docker compose up
  ```