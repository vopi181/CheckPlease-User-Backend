#grpcurl -v -plaintext -d '{\"phone\": \"8479095558\", \"fname\":\"d\", \"lname\":\"p\" }' localhost:9000 CPUser.CPUser/CreateUser
#grpcurl -v -plaintext -d '{ \"authrequest\": { \"token\": \"w3rci8o1i6lvbobcahewqztv\" }, \"tabletoken\": \"1234\" }' checkplease.app:9000 cpuser.cpuser/orderinitiation
#grpcurl -v -plaintext -d '{ \"authrequest\": { \"token\": \"w3rci8o1i6lvbobcahewqztv\" }, \"tokencode\": \"1234\" }' checkplease.app:9000 cpuser.cpuser/ItemPaySubscribe

grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }' localhost:9000 CPUser.CPUser/ItemPaySubscribe
