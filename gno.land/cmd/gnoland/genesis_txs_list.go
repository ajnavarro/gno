package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newTxsListCmd list all transactions on the specified genesis file
func newTxsListCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "txs list [flags] [<arg>...]",
			ShortHelp:  "lists transactions existing on genesis.json",
			LongHelp:   "Lists transactions existing on genesis.json",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsListCmd(io, txsCfg)
		},
	)

	return cmd
}

func execTxsListCmd(io commands.IO, cfg *txsCfg) error {
	genesis, err := types.GenesisDocFromFile(cfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	gs, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return fmt.Errorf("genesis state is not using the correct Gno Genesis type.")
	}

	tw := tabwriter.NewWriter(io.Out(), 0, 8, 2, '\t', 0)
	for _, tx := range gs.Txs {
		for _, msg := range tx.Msgs {
			switch m := msg.(type) {
			case vmm.MsgAddPackage:
				fmt.Fprintf(tw, "create\tpath:%s\tfiles:%d\tcreator:%s\t\n", m.Package.Path, len(m.Package.Files), m.Creator.String())
			case vmm.MsgCall:
				fmt.Fprintf(tw, "call\tpath:%s\tparams:%d\tcaller:%s\t\n", m.PkgPath, len(m.Args), m.Caller.String())
			}
		}
	}

	return tw.Flush()
}
