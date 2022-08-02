### ECR Login
aws ecr get-login-password --region ap-northeast-1 --profile tdt | docker login --username AWS --password-stdin 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/tdt-broadcast
aws ecr get-login-password --profile tdt | docker login --username AWS --password-stdin 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/tdt-broadcast
docker login -u AWS -p $(aws ecr get-login-password --region ap-northeast-1 --profile tdt) 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com

### Pull
docker pull 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/tdt-broadcast:latest

### Build & Push
sdocker tag tdt:latest 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/tdt-broadcast:latest
docker push 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/tdt-broadcast:latest

### ssh 
ssh root@127.0.0.1 -p 22

### redis
redis-cli -c -h tdt-broadcast.v81cls.ng.0001.apne1.cache.amazonaws.com:6379 -p 6379

### 
docker run -itd tdt /bin/sh


docker build -f Dockerfile -t ssh .
docker tag ssh:latest 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/ssh-test:latest
docker push 713687319441.dkr.ecr.ap-northeast-1.amazonaws.com/ssh-test:latest


AWS_ACCESS_KEY=xxxxx AWS_SECRET_ACCESS_KEY=xxxxx go run ./main.go -c config.toml
