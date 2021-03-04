#grpcurl -v -plaintext -d '{\"phone\": \"8479095558\", \"fname\":\"d\", \"lname\":\"p\" }' localhost:9000 CPUser.CPUser/CreateUser

# LOCAL
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"asdfasdf\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"asdfasdf\" }, \"tokenCode\": \"12345\" }'     localhost:9000 CPUser.CPUser/SelectionInitial
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }'     localhost:9000 CPUser.CPUser/SelectionInitial
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }' localhost:9000 CPUser.CPUser/ItemPaySubscribe


# REMOTE

grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"TZJLmNQfxmJVu03i9Cteeswt\" }, \"tableToken\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/OrderInitiation
 grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"TZJLmNQfxmJVu03i9Cteeswt\" }, \"tokenCode\": \"1234\" }'     checkplease.app:9000 CPUser.CPUser/SelectionInitial
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"TZJLmNQfxmJVu03i9Cteeswt\" }, \"tokenCode\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/ItemPaySubscribe
