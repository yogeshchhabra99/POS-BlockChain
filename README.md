#POS-BLOCKCHAIN

This project uses golang to create a simple p2p blockchain. Read the problem statement pdf for more details about the the problem.

##Files:
p2p_blockchain.go: peer to peer blockchain implementation for the problem. To run execute: go run p2p_blockchain.go -l 8000 -secio . Here 8000 is the port number you can choose other port also. This will give you the command you need to enter in a new terminal to connect a new device to the blockchain network. Now you can enter user commands. Read below for commands.

server_blockchain.go: this is a servered blockchain i.e. there is a central server or device where all transactions are stored. This is not a true blockchain.

if you wanna create an executable then use go build and pass the same parameters while running the executable created

##User Commands:
###First command:
The first command has to be either CASH or RETAIL to select which POS terminal this is. I tried doing this via commandline but the threads break.
1. CASH : selects this as cash POS terminal
2. RETAIL: selects this as retail POS terminal.
Note that these commands work only once.

###Cash Card POS commands:
1. NEWUSER : creates a new user and gives you his id
2. RECHARGE $USERID $AMOUNT : Recharging/crediting an account with userid=$USERID e.g. RECHARGE 2 100; adds 100 rs to account with userid 2

###Retail POS commands:
1. DEBIT $USERID $AMOUNT : debit amount from account 

##Caution: 
The package go-libp2p has some issues with some versions of go. They claim that the package works with go versions >=1.12 but it gave errors with versions 1.12 and 1.13 as well. I used version 1.15 at last which works fine.

If you need help installing golang:
1.) check for older go versions using $ go version
2.) if you already have go and its not version 1.15.* . Run $ echo $GOROOT or $ echo $PATH to find where is go installed, go to that directory and remove all go files. Download 1.15 version and extract at the same location to ensure you dont have to change GOROOT or PATH variables.
3.) If you have to change the environment variables due to some version name, look for them in $HOME/~/.profile or $HOME/~/.bashrc
4.)The go packages are installed in $GOPATH directory, clean that directory to ensure all packages a re downloaded again.
5.) Make sure GOPATH and GOROOT are different and GOPATH is not inside GOROOT.

##Sample Run:
0. You can side by side see screenshots.pdf to see the blockchain working
1. In Terminal 1 run: "go run p2p_blockchain.go -l 8000 -secio" This gives you another command you need to paste in another terminal to securely connect to peer to peer network. The command would be something like "go run p2p_blockchain.go -l 8001 -d /ip4/127.0.0.1/tcp/8000/p2p/QmaPTcwcqqAyS1Nz82nnxRXN5ZE9DmXyX7qLYWniBfLP6Z -secio", this is sample it wont work in your pc.
2. Paste the copied command in another terminal, make sure you are at the hoe of the project
3. Now the network starts with two devices
4. Choose terminal 1 as CASH. To do that type "CASH" in terminal 1 and press enter. 
5. Choose terminal 2 as RETAIL. To do that type "RETAIL" in terminal 2 and press enter
6. In the terminal 1, create a new User. To do that type "NEWUSER" in terminal 1 and press enter. This shows new user created with userId =1. You will see a new transaction added to the blockchain with type="NEW_ACCOUNT" and the balance and amount fields set to 0.
7. Recharge the account of userId= 1 with an amount 1000 Rs. To do that type "RECHARGE 1 1000" in terminal 1 and press enter. This adds 1000 Rs to account with userId=1. You will see another transaction added to blockchain with type="CREDIT", amount =1000 and balance =1000. You can see now how the transactions hold all the history.
8. Debit 600 from the account with userId=1. To do that type "DEBIT 1 600" in terminal 2 and press enter. This debits 600 from account with userId=1. You will see a new transaction added to blockchain with type = DEBIT, amount =600, and balance=400. This way the latest balance is holded by the top transaction for that account in the blockchain.
9. The program informs you in case of a type, or a user with given id doesn't exist or amount you entered is -ve.
10. You can play more by using 2 cash terminals. creating a user in one and recharging it in other.


References:
https://medium.com/@mycoralhealth/code-a-simple-p2p-blockchain-in-go-46662601f417