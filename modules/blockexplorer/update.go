package blockexplorer

import (
	"fmt"

	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

// Handles updates recieved from the consensus subscription. Keeps
// track of transaction volume, block timestamps and block sizes, as
// well as the current block height
func (be *BlockExplorer) ReceiveConsensusSetUpdate(cc modules.ConsensusChange) {
	lockID := be.mu.Lock()
	defer be.mu.Unlock(lockID)

	// Modify the number of file contracts and how much they costed
	for _, diff := range cc.FileContractDiffs {
		if diff.Direction == modules.DiffApply {
			be.activeContracts += 1
			be.totalContracts += 1
			be.activeContractCost = be.activeContractCost.Add(diff.FileContract.Payout)
			be.totalContractCost = be.totalContractCost.Add(diff.FileContract.Payout)
		} else {
			be.activeContracts -= 1
			be.activeContractCost = be.activeContractCost.Sub(diff.FileContract.Payout)
		}
	}

	// Reverting the blockheight and block data structs from reverted blocks
	be.blockchainHeight -= types.BlockHeight(len(cc.RevertedBlocks))

	// Handle incoming blocks
	for _, block := range cc.AppliedBlocks {
		// add the block to the database.
		err := be.addBlockDB(block)
		if err != nil {
			fmt.Printf("Error when adding block to database: " + err.Error() + "\n")
		}
		be.blockchainHeight += 1
	}
	be.currentBlock = cc.AppliedBlocks[len(cc.AppliedBlocks)-1]

	// Notify subscribers about updates
	be.updateSubscribers()
}
