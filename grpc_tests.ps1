#grpcurl -v -plaintext -d '{\"phone\": \"8479095558\", \"fname\":\"d\", \"lname\":\"p\" }' localhost:9000 CPUser.CPUser/CreateUser
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"W3rci8O1I6LvBoBcAhEwQzTv\" }, \"tableToken\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/OrderInitiation
grpcurl -v -plaintext -d '{ \"authRequest\": { \"Token\": \"W3rci8O1I6LvBoBcAhEwQzTv\" }, \"tokenCode\": \"1234\" }' checkplease.app:9000 CPUser.CPUser/SelectionSubscribe
