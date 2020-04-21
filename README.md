# MythicalHP
MythicalHP
This is a simple introduction to enabling Primary & Secondary with Azure Standard Load Balancer. A SLB with 2 VMs, will need to support Active Passive mode. When the Primary server goes down, the Secondary will need to be Active. When Primary comes back up, Secondary will need to go passive

High Level Architecture Diagram
![Image description](./MythicalHPLB.png)

Call Sequence between Primary and Secondary
![Image description](./MythicalHP.png)


## Steps
* Ubuntu VMs
* Deploy Standard Load Balancer with Backend Pool
* Install https://github.com/golang/go/wiki/Ubuntu
* go get -u github.com/gorilla/mux
* git clone https://github.com/jaypaddy/MythicalHP.git
* cd MythicalHP
* [PRIMARY] sudo go run . -role=primary -tcpprobe=primarymq -agentport=80
* [SECONDARY] sudo go run . -role=secondary -tcpprobe=primarymq -agentport=80

role = this is the role of the server where the agent is running. Assumption is that the agent will run on the same server as the primary service i.e. the workload
tcpprobe = this will be tcp service that needs to be assessed for health
agentport = the port at which the agent will run. this should match with the port used on the Load Balancer Healthprobe. On the loadbalancer it will be specified as http://<server>:<agentport>/healthprobe