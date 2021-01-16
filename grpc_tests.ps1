#grpcurl -v -plaintext -d '{\"phone\": \"8479095558\", \"fname\":\"d\", \"lname\":\"p\" }' localhost:9000 CPUser.CPUser/CreateUser
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }' localhost:9000 CPUser.CPUser/SelectionSubscribe
