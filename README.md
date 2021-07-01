# Secure Multiparty Computations - MPC
This following repo contains two MPC protocols. A simple MPC protocol 
implemented to get a better understanding of MPC's and a protocol using Shamir Secret Sharing. 
The function to be computed in the protocols is defined in a circuit. 

### Defining a circuit
A circuit contains information about how many parties participate, the field, the secret sharing method, 
if it should preprocess and if the corrupt parties are active in the protocol. The circuit contains gates that define 
different computations to be made. The first gate number should be one larger than all the inputs given to the protocol.
The gates should come in sequential order. I.e. a gate can't take as input a gate with a higher gate number than itself.
If the active field is set to true in the circuit the number of parties to participate needs to be greater than three parties.

Note that the simple protocols circuit can only contain one gate, and it needs to be an addition or multiplication gate.

### Running the protocol
To run the protocol go into the folder MPC. Then type in the terminal 'go run MPC.go "Circuit" "secret"'. 
As an example if one would run thee circuit YaoBits2 it would type 'go run MPC.go YaoBits2 01' (if the user should give input else leave it blank)
The protocol will ask for an IP-address to connect to. If the user should be the host then hit ENTER.
If not, type the IP-address of the host. The parties will have the party number in the order they connect to the host.

### Creating a Yao circuit
It is possible to create any Yao circuit to compare two parties' bit strings. In the MPC folder type 'go run YaoCC.go "bit length" "party size" "preprocessing true/false" "active true/false"'
As an example 'go run YaoCC.go 15 45 true true'.