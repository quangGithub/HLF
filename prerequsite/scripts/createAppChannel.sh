peer channel create -o orderer:7050 -c mychannel -f ./channel-artifacts/mychannel.tx --outputBlock ./channel-artifacts/mychannel.block --tls --cafile /organizations/ordererOrganizations/agm.com/orderers/orderer.agm.com/msp/tlscacerts/tlsca.agm.com-cert.pem