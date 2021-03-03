# Connecting to the Sifchain BetaNet. 

## Prerequisites / Dependencies:

- [Docker](https://www.docker.com/get-started)
- [Ruby 2.7.x](https://www.ruby-lang.org/en/documentation/installation)
- [Golang](https://golang.org/doc/install)
- [Git]

## Here are the step by step instructions.

```
sudo su -
apt update
apt upgrade
apt install ruby-full
apt install docker
apt install docker-compose
apt install git
```

Download the go tar.gz file from [here](https://golang.org/doc/install)

```
tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
export GOPATH=/usr/local/go
export PATH=$PATH:$GOPATH/bin
```
  - For convenience you can add the last two lines into your profile (.bashrc or .bash_profile file)

## Scaffold and run your node

1. Clone the repository:

```
git clone https://github.com/Sifchain/sifnode && cd sifnode
```

2. Build:

```
make clean install
```

3. Generate a mnemonic (if you don't already have one):

```
rake "keys:generate:mnemonic"
```

  - This will generate the mnemonic keys for you. Note it down on a piece of paper, create a passphrase and remember it!
  - Do not share the mnemonic keys with anyone and keep the keys safe.

4. Boot your node:

```
rake "genesis:sifnode:mainnet:boot[<moniker>,'<mnemonic>',<gas_price>,<bind_ip_address>]"
```

Where:

|Param|Description|
|-----|----------|
|`<moniker>`|A name for your node. Could be anything like MySifNode|
|`<mnemonic>`|The mnemonic phrase generated in the previous step.|
|`<gas_price>`|The minimum gas price (e.g.: 0.5rowan).|
|`<bind_ip_address>`|The IP Address to bind to (*Important:* this is what your node will advertise to the rest of the network). This should be the public IP of the host. You can usually fine your IP address using this command `ping $(hostname)`| 

and your node will start synchronizing with the network. Please note that this may take several hours or more. You can check the latest block height from this url: https://blockexplorer.sifchain.finance/. To check how much you have caught up, run this:

```
curl -Ss $(hostname):26657/status | jq -er '.result.sync_info'
```

## Verify

Once you have caught up with the latest block height, run

``` 
sifnodecli q tendermint-validator-set --node tcp://$(hostname):26657 --trust-node | grep address | wc -l
```

   • This should give you a number, for example 54. Go to https://blockexplorer.sifchain.finance/validators and see if the Active number of validators match with your previous command output. If it did, Congratulations. You are now connected to the network.

## Become a Validator

You won't be able to participate in consensus until you become a validator.

1. Import your mnemonic locally:

```
rake "keys:import[<moniker>]"
```

Where:

|Param|Description|
|-----|----------|
|`<moniker>`|A name for your node.|

   • You will need to have tokens (rowan) on your account in order to become a validator.
   • If you have rowan on another sif wallet, transfer it to the one you created using the mnemonic keys.
   • Obtain your node moniker (if you don't already know it):

``` 
cat ~/.sifnoded/config/config.toml | grep moniker
``` 

2. Go inside the docker container

```
docker ps
```

The above command should give you your container ID. Copy it. Then run

```
docker exec -it YOUR_CONTAINER_ID sh
```
   • You are now inside the docker container.
    
3. From within your running container, obtain your node's public key:

```
/root/.sifnoded/cosmovisor/genesis/bin/sifnoded tendermint show-validator
```

Note down the output of the above command. This will be your public key <pub_key>

4. Run the following command to become a validator: 

```
sifnodecli tx staking create-validator \
    --commission-max-change-rate="0.1" \
    --commission-max-rate="0.1" \
    --commission-rate="0.1" \
    --amount="<amount>" \
    --pubkey=<pub_key> \
    --moniker=<moniker> \
    --chain-id=sifchain \
    --min-self-delegation="1" \
    --gas-prices="0.5rowan" \
    --from=<moniker> \
    --keyring-backend=file \
    --node tcp://44.235.108.41:26657
```

Where:

|Param|Description|
|-----|----------|
|`Commission`|Commission rates are in %. 0.1 means 10%. Modify the amount of commission you want.|
|`<amount>`|The amount of rowan you wish to stake (the more the better). rowan amount should be multiplied into 10^18. For example if you want to stake 10 rowan, then you add 18 zeroes at the end to make it like 10000000000000000000rowan as the amount.|
|`<pub_key>`|The public key of your node, that you got in the previous step.|
|`<moniker>`|The moniker (name) of your node. Note you have to type moniker twice in the command below.|



e.g.:

```
sifnodecli tx staking create-validator \
    --commission-max-change-rate="0.1" \
    --commission-max-rate="0.1" \
    --commission-rate="0.1" \
    --amount="1000000000000000000000rowan" \
    --pubkey=$(/root/.sifnoded/cosmovisor/genesis/bin/sifnoded tendermint show-validator) \
    --moniker=MySifNode \
    --chain-id=sifchain \
    --min-self-delegation="1" \
    --gas-prices="0.5rowan" \
    --from=MySifNode \
    --keyring-backend=file \
    --node tcp://$(hostname):26657
```

## Additional Resources

Join our discord [here](https://discord.gg/pArfJZwX) if you have any other questions.

### Endpoints

|Description|Address|
|-----------|-------|
|Block Explorer|https://blockexplorer.sifchain.finance|
|RPC|https://rpc.sifchain.finance|
|API|https://api.sifchain.finance|
