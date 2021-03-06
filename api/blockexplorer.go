package api

import (
	"fmt"
	"net/http"

	"github.com/NebulousLabs/Sia/types"
)

// Handles the call to get information about the blockchain. Returns
// several data points such as chain height, the current block, and
// file contract info
func (srv *Server) blockexplorerStatusHandler(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, srv.blocke.ExplorerStatus())
}

// Handles the call to get many data points on blocks
func (srv *Server) blockexplorerBlockDataHandler(w http.ResponseWriter, req *http.Request) {
	// Extract the start and end point from the request
	var start, finish types.BlockHeight
	_, err := fmt.Sscan(req.FormValue("start"), &start)
	if err != nil {
		writeError(w, "Malformed or no start height", http.StatusBadRequest)
		return
	}

	// If a range end is not given, assume the range end to be one
	// greater than the range start, returning one block
	_, err = fmt.Sscan(req.FormValue("finish"), &finish)
	if err != nil {
		finish = start + 1
	}

	// Bounds checking is done inside BlockInfo
	blockSummaries, err := srv.blocke.BlockInfo(start, finish)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, blockSummaries)
}
