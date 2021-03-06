package miner

import (
	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/encoding"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

// ReceiveTransactionPoolUpdate listens to the transaction pool for changes in
// the transaction pool. These changes will be applied to the blocks being
// mined.
func (m *Miner) ReceiveTransactionPoolUpdate(cc modules.ConsensusChange, unconfirmedTransactions []types.Transaction, _ []modules.SiacoinOutputDiff) {
	lockID := m.mu.Lock()
	defer m.mu.Unlock(lockID)
	defer m.notifySubscribers()

	m.height -= types.BlockHeight(len(cc.RevertedBlocks))
	m.height += types.BlockHeight(len(cc.AppliedBlocks))

	// The total encoded size of the transactions cannot exceed the block size.
	m.transactions = nil
	remainingSize := int(types.BlockSizeLimit - 5e3)
	for {
		if len(unconfirmedTransactions) == 0 {
			break
		}
		remainingSize -= len(encoding.Marshal(unconfirmedTransactions[0]))
		if remainingSize < 0 {
			break
		}

		m.transactions = append(m.transactions, unconfirmedTransactions[0])
		unconfirmedTransactions = unconfirmedTransactions[1:]
	}

	// If no blocks have been applied, the block variables do not need to be
	// updated.
	if len(cc.AppliedBlocks) == 0 {
		if build.DEBUG {
			if len(cc.RevertedBlocks) != 0 {
				panic("blocks reverted without being added")
			}
		}
		return
	}

	// Update the parent, target, and earliest timestamp fields for the miner.
	m.parent = cc.AppliedBlocks[len(cc.AppliedBlocks)-1].ID()
	target, exists1 := m.cs.ChildTarget(m.parent)
	timestamp, exists2 := m.cs.EarliestChildTimestamp(m.parent)
	if build.DEBUG {
		if !exists1 {
			panic("could not get child target")
		}
		if !exists2 {
			panic("could not get child earliest timestamp")
		}
	}
	m.target = target
	m.earliestTimestamp = timestamp
}
