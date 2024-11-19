package main

import (
    "fmt"

    mxChainCore "github.com/multiversx/mx-chain-core-go/core"
    "github.com/multiversx/mx-chain-core-go/core/check"
    "github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
    "github.com/multiversx/mx-chain-core-go/hashing"
    "github.com/multiversx/mx-chain-core-go/hashing/keccak"
    "github.com/multiversx/mx-chain-core-go/marshal"
    "github.com/multiversx/mx-chain-go/process/factory"
    "github.com/multiversx/mx-chain-go/process/smartContract/hooks"
    "github.com/multiversx/mx-sdk-go/core"
    "github.com/multiversx/mx-sdk-go/data"
    "github.com/multiversx/mx-sdk-go/disabled"
    "github.com/multiversx/mx-sdk-go/storage"
)

const accountStartNonce = uint64(0)
var initialDNSAddress = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

type addressGenerator struct {
    blockChainHook hooks.BlockChainHookHandler
    hasher         hashing.Hasher
}

func NewAddressGenerator() (*addressGenerator, error) {
    builtInFuncs := &disabled.BuiltInFunctionContainer{}

    argsHook := hooks.ArgBlockChainHook{
        Accounts:                 &disabled.Accounts{},
        PubkeyConv:               core.AddressPublicKeyConverter,
        StorageService:           &disabled.StorageService{},
        BlockChain:               &disabled.Blockchain{},
        ShardCoordinator:         &disabled.ShardCoordinator{},
        Marshalizer:              &marshal.JsonMarshalizer{},
        Uint64Converter:          uint64ByteSlice.NewBigEndianConverter(),
        BuiltInFunctions:         builtInFuncs,
        DataPool:                 &disabled.DataPool{},
        CompiledSCPool:           storage.NewMapCacher(),
        NilCompiledSCStore:       true,
        NFTStorageHandler:        &disabled.SimpleESDTNFTStorageHandler{},
        EpochNotifier:            &disabled.EpochNotifier{},
        GlobalSettingsHandler:    &disabled.GlobalSettingsHandler{},
        EnableEpochsHandler:      &disabled.EnableEpochsHandler{},
        GasSchedule:              &disabled.GasScheduleNotifier{},
        Counter:                  &disabled.BlockChainHookCounter{},
        MissingTrieNodesNotifier: &disabled.MissingTrieNodesNotifier{},
    }
    blockchainHook, err := hooks.NewBlockChainHookImpl(argsHook)
    if err != nil {
        return nil, err
    }

    return &addressGenerator{
        blockChainHook: blockchainHook,
        hasher:         keccak.NewKeccak(),
    }, nil
}

func (ag *addressGenerator) CompatibleDNSAddress(shardId byte) (core.AddressHandler, error) {
    addressLen := len(initialDNSAddress)
    shardInBytes := []byte{0, shardId}

    newDNSPk := string(initialDNSAddress[:(addressLen-mxChainCore.ShardIdentiferLen)]) + string(shardInBytes)
    newDNSAddress, err := ag.blockChainHook.NewAddress([]byte(newDNSPk), accountStartNonce, factory.WasmVirtualMachine)
    if err != nil {
        return nil, err
    }

    return data.NewAddressFromBytes(newDNSAddress)
}

func (ag *addressGenerator) CompatibleDNSAddressFromUsername(username string) (core.AddressHandler, error) {
    hash := ag.hasher.Compute(username)
    lastByte := hash[len(hash)-1]
    return ag.CompatibleDNSAddress(lastByte)
}

func (ag *addressGenerator) ComputeWasmVMScAddress(address core.AddressHandler, nonce uint64) (core.AddressHandler, error) {
    if check.IfNil(address) {
        return nil, fmt.Errorf("nil address")
    }

    scAddressBytes, err := ag.blockChainHook.NewAddress(address.AddressBytes(), nonce, factory.WasmVirtualMachine)
    if err != nil {
        return nil, err
    }

    return data.NewAddressFromBytes(scAddressBytes)
}

func main() {
    addressGen, err := NewAddressGenerator()
    if err != nil {
        fmt.Println("Error creating address generator:", err)
        return
    }

    shardId := byte(1)
    address, err := addressGen.CompatibleDNSAddress(shardId)
    if err != nil {
        fmt.Println("Error generating compatible DNS address:", err)
        return
    }
    fmt.Println("Generated DNS Address:", address)

    username := "exampleUser"
    addressFromUsername, err := addressGen.CompatibleDNSAddressFromUsername(username)
    if err != nil {
        fmt.Println("Error generating address from username:", err)
        return
    }
    fmt.Println("Generated Address from Username:", addressFromUsername)

    ownerAddress, err := data.NewAddressFromBytes([]byte{1, 2, 3, 4})
    if err != nil {
        fmt.Println("Error creating owner address:", err)
        return
    }
    nonce := uint64(0)
    scAddress, err := addressGen.ComputeWasmVMScAddress(ownerAddress, nonce)
    if err != nil {
        fmt.Println("Error generating smart contract address:", err)
        return
    }
    fmt.Println("Generated Smart Contract Address:", scAddress)
}