set -x

mkdir -p /organizations/peerOrganizations/match.agm.com/

export FABRIC_CA_CLIENT_HOME=/organizations/peerOrganizations/match.agm.com/



fabric-ca-client enroll -u https://admin:adminpw@ca-match:7054 --caname ca-match --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"



echo 'NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca-match-7054-ca-match.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca-match-7054-ca-match.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca-match-7054-ca-match.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca-match-7054-ca-match.pem
    OrganizationalUnitIdentifier: orderer' > "/organizations/peerOrganizations/match.agm.com/msp/config.yaml"



fabric-ca-client register --caname ca-match --id.name peer0 --id.secret peer0pw --id.type peer --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"



fabric-ca-client register --caname ca-match --id.name user1 --id.secret user1pw --id.type client --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"




fabric-ca-client register --caname ca-match --id.name matchadmin --id.secret matchadminpw --id.type admin --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"



fabric-ca-client enroll -u https://peer0:peer0pw@ca-match:7054 --caname ca-match -M "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/msp" --csr.hosts peer0.match.agm.com --csr.hosts  peer0-match --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"



cp "/organizations/peerOrganizations/match.agm.com/msp/config.yaml" "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/msp/config.yaml"



fabric-ca-client enroll -u https://peer0:peer0pw@ca-match:7054 --caname ca-match -M "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls" --enrollment.profile tls --csr.hosts peer0.match.agm.com --csr.hosts  peer0-match --csr.hosts ca-match --csr.hosts localhost --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"




cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/tlscacerts/"* "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/ca.crt"
cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/signcerts/"* "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/server.crt"
cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/keystore/"* "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/server.key"

mkdir -p "/organizations/peerOrganizations/match.agm.com/msp/tlscacerts"
cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/tlscacerts/"* "/organizations/peerOrganizations/match.agm.com/msp/tlscacerts/ca.crt"

mkdir -p "/organizations/peerOrganizations/match.agm.com/tlsca"
cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/tls/tlscacerts/"* "/organizations/peerOrganizations/match.agm.com/tlsca/tlsca.match.agm.com-cert.pem"

mkdir -p "/organizations/peerOrganizations/match.agm.com/ca"
cp "/organizations/peerOrganizations/match.agm.com/peers/peer0.match.agm.com/msp/cacerts/"* "/organizations/peerOrganizations/match.agm.com/ca/ca.match.agm.com-cert.pem"


fabric-ca-client enroll -u https://user1:user1pw@ca-match:7054 --caname ca-match -M "/organizations/peerOrganizations/match.agm.com/users/User1@match.agm.com/msp" --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"

cp "/organizations/peerOrganizations/match.agm.com/msp/config.yaml" "/organizations/peerOrganizations/match.agm.com/users/User1@match.agm.com/msp/config.yaml"

fabric-ca-client enroll -u https://matchadmin:matchadminpw@ca-match:7054 --caname ca-match -M "/organizations/peerOrganizations/match.agm.com/users/Admin@match.agm.com/msp" --tls.certfiles "/organizations/fabric-ca/match/tls-cert.pem"

cp "/organizations/peerOrganizations/match.agm.com/msp/config.yaml" "/organizations/peerOrganizations/match.agm.com/users/Admin@match.agm.com/msp/config.yaml"

{ set +x; } 2>/dev/null
