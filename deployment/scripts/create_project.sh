#!/bin/bash

mkdir -p auth/cmd auth/internal auth/proto auth/authpb auth product/cmd product/internal product/proto product/productpb user/cmd user/internal user/proto user/userpb order/cmd order/internal order/proto order/orderpb logging/cmd logging/internal logging/proto logging/loggingpb api-gateway deploy/scripts

touch auth/proto/auth.proto auth/cmd/auth_server/main.go auth/internal/server/server.go product/proto/product.proto product/cmd/product_server/main.go product/internal/server/server.go user/proto/user.proto user/cmd/user_server/main.go user/internal/server/server.go order/proto/order.proto order/cmd/order_server/main.go order/internal/server/server.go logging/proto/logging.proto logging/cmd/logging_server/main.go logging/internal/server/server.go api-gateway/Dockerfile deploy/scripts/deploy.sh 

echo "Project structure created successfully!"