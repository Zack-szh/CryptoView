# CryptoView

Useful commands: 
- start docker: 
    docker-compose up 
- close docker while deleting database: 
    docker-compose down -v 
- launch interactive psql shell: 
    docker exec -it cryptoview-postgres-1 psql -U cryptoview -d cryptoview