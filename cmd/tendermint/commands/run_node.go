package commands

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	//cfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
)

var runNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Run the tendermint node",
	RunE:  runNode,
}

func init() {
	// bind flags

	// node flags
	runNodeCmd.Flags().String("moniker", config.Moniker,
		"Node Name")
	viperConfig.BindPFlag("moniker", runNodeCmd.Flags().Lookup("moniker"))

	runNodeCmd.Flags().Bool("fast_sync", config.FastSync,
		"Fast blockchain syncing")
	viperConfig.BindPFlag("fast_sync", runNodeCmd.Flags().Lookup("fast_sync"))

	// abci flags
	runNodeCmd.Flags().String("proxy_app", config.ProxyApp,
		"Proxy app address, or 'nilapp' or 'dummy' for local testing.")
	viperConfig.BindPFlag("proxy_app", runNodeCmd.Flags().Lookup("proxy_app"))

	runNodeCmd.Flags().String("abci", config.ABCI,
		"Specify abci transport (socket | grpc)")
	viperConfig.BindPFlag("abci", runNodeCmd.Flags().Lookup("abci"))

	// rpc flags
	runNodeCmd.Flags().String("rpc_laddr", config.RPCListenAddress,
		"RPC listen address. Port required")
	viperConfig.BindPFlag("rpc_laddr", runNodeCmd.Flags().Lookup("rpc_laddr"))

	runNodeCmd.Flags().String("grpc_laddr", config.GRPCListenAddress,
		"GRPC listen address (BroadcastTx only). Port required")
	viperConfig.BindPFlag("grpc_laddr", runNodeCmd.Flags().Lookup("grpc_laddr"))

	// p2p flags
	runNodeCmd.Flags().String("p2p.laddr", config.P2P.ListenAddress,
		"Node listen address. (0.0.0.0:0 means any interface, any port)")
	viperConfig.BindPFlag("p2p.laddr", runNodeCmd.Flags().Lookup("p2p.laddr"))

	runNodeCmd.Flags().String("p2p.seeds", config.P2P.Seeds,
		"Comma delimited host:port seed nodes")
	viperConfig.BindPFlag("p2p.seeds", runNodeCmd.Flags().Lookup("p2p.seeds"))

	runNodeCmd.Flags().Bool("p2p.skip_upnp", config.P2P.SkipUPNP,
		"Skip UPNP configuration")
	viperConfig.BindPFlag("p2p.skip_upnp", runNodeCmd.Flags().Lookup("p2p.skip_upnp"))

	// feature flags
	runNodeCmd.Flags().Bool("p2p.pex", config.P2P.PexReactor,
		"Enable Peer-Exchange (dev feature)")
	viperConfig.BindPFlag("p2p.pex", runNodeCmd.Flags().Lookup("p2p.pex"))

	RootCmd.AddCommand(runNodeCmd)
}

// Users wishing to:
//	* Use an external signer for their validators
//	* Supply an in-proc abci app
// should import github.com/tendermint/tendermint/node and implement
// their own run_node to call node.NewNode (instead of node.NewNodeDefault)
// with their custom priv validator and/or custom proxy.ClientCreator
func runNode(cmd *cobra.Command, args []string) error {

	// Wait until the genesis doc becomes available
	// This is for Mintnet compatibility.
	// TODO: If Mintnet gets deprecated or genesis_file is
	// always available, remove.
	genDocFile := config.GenesisFile
	if !cmn.FileExists(genDocFile) {
		log.Notice(cmn.Fmt("Waiting for genesis file %v...", genDocFile))
		for {
			time.Sleep(time.Second)
			if !cmn.FileExists(genDocFile) {
				continue
			}
			jsonBlob, err := ioutil.ReadFile(genDocFile)
			if err != nil {
				return fmt.Errorf("Couldn't read GenesisDoc file: %v", err)
			}
			genDoc, err := types.GenesisDocFromJSON(jsonBlob)
			if err != nil {
				return fmt.Errorf("Error reading GenesisDoc: %v", err)
			}
			if genDoc.ChainID == "" {
				return fmt.Errorf("Genesis doc %v must include non-empty chain_id", genDocFile)
			}

			// config.SetChainID("chain_id", genDoc.ChainID) TODO
		}
	}

	// Create & start node
	n := node.NewNodeDefault(config)
	if _, err := n.Start(); err != nil {
		return fmt.Errorf("Failed to start node: %v", err)
	} else {
		log.Notice("Started node", "nodeInfo", n.Switch().NodeInfo())
	}

	// Trap signal, run forever.
	n.RunForever()

	return nil
}
