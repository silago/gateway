docker:
	docker build  --tag=pool-gateway-api .
	docker tag pool-gateway-api:latest 072388890876.dkr.ecr.eu-central-1.amazonaws.com/pool-gateway-api:latest
	docker push 072388890876.dkr.ecr.eu-central-1.amazonaws.com/pool-gateway-api:latest