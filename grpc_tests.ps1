#grpcurl -v -plaintext -d '{\"phone\": \"8479095558\", \"fname\":\"d\", \"lname\":\"p\" }' localhost:9000 CPUser.CPUser/CreateUser

# LOCAL
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"asdfasdf\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tableToken\": \"12345\" }' localhost:9000 CPUser.CPUser/OrderInitiation
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"asdfasdf\" }, \"tokenCode\": \"12345\" }'     localhost:9000 CPUser.CPUser/SelectionInitial
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }'     localhost:9000 CPUser.CPUser/SelectionInitial
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"EpLj3ssCAIsuaVMJJinsSVBY\" }, \"tokenCode\": \"12345\" }' localhost:9000 CPUser.CPUser/SelectionSubscribe


# REMOTE

#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"YE1v2VLj46UCEKdupoL1AH5e\" }, \"tableToken\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/OrderInitiation
# grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"YE1v2VLj46UCEKdupoL1AH5e\" }, \"tokenCode\": \"1234\" }'     checkplease.app:9000 CPUser.CPUser/SelectionInitial
#grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"YE1v2VLj46UCEKdupoL1AH5e\" }, \"tokenCode\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/SelectionSubscribe
