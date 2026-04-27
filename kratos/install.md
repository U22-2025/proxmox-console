```
docker run --rm -it \
  -p 4433:4433 \
  -p 4434:4434 \
  -v $(pwd)/kratos.yml:/etc/config/kratos.yml \
  -v $(pwd)/identity.schema.json:/etc/config/identity.schema.json \
  oryd/kratos:latest serve -c /etc/config/kratos.yml
  ```