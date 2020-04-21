package utreexo

import (
	"fmt"
	"testing"
)

func TestForestAddDel(t *testing.T) {

	numAdds := uint32(10)

	f := NewForest(nil)

	sc := NewSimChain(0x07)
	sc.lookahead = 400

	for b := 0; b < 1000; b++ {

		adds, _, delHashes := sc.NextBlock(numAdds)

		bp, err := f.ProveBatch(delHashes)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.Modify(adds, bp.Targets)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("nl %d %s", f.numLeaves, f.ToString())
	}
}

// Add 2. delete 1.  Repeat.
func Test2Fwd1Back(t *testing.T) {
	f := NewForest(nil)
	var absidx uint32
	adds := make([]LeafTXO, 2)

	for i := 0; i < 100; i++ {

		for j := range adds {
			adds[j].Hash[0] = uint8(absidx>>8) | 0xa0
			adds[j].Hash[1] = uint8(absidx)
			adds[j].Hash[3] = 0xaa
			absidx++
			//		if i%30 == 0 {
			//			utree.Track(adds[i])
			//			trax = append(trax, adds[i])
			//		}
		}

		//		t.Logf("-------- block %d\n", i)
		fmt.Printf("\t\t\t########### block %d ##########\n\n", i)

		// add 2
		_, err := f.Modify(adds, nil)
		if err != nil {
			t.Fatal(err)
		}

		s := f.ToString()
		fmt.Printf(s)

		// get proof for the first
		_, err = f.Prove(adds[0].Hash)
		if err != nil {
			t.Fatal(err)
		}

		// delete the first
		//		err = f.Modify(nil, []Hash{p.Payload})
		//		if err != nil {
		//			t.Fatal(err)
		//		}

		//		s = f.ToString()
		//		fmt.Printf(s)

		// get proof for the 2nd
		keep, err := f.Prove(adds[1].Hash)
		if err != nil {
			t.Fatal(err)
		}
		// check proof

		worked := f.Verify(keep)
		if !worked {
			t.Fatalf("proof at position %d, length %d failed to verify\n",
				keep.Position, len(keep.Siblings))
		}
	}
}

// Add and delete variable numbers, repeat.
// deletions are all on the left side and contiguous.
func TestAddxDelyLeftFullBatchProof(t *testing.T) {
	for x := 0; x < 10; x++ {
		for y := 0; y < x; y++ {
			err := AddDelFullBatchProof(x, y)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

}

// Add x, delete y, construct & reconstruct blockproof
func AddDelFullBatchProof(nAdds, nDels int) error {
	if nDels > nAdds-1 {
		return fmt.Errorf("too many deletes")
	}

	f := NewForest(nil)
	adds := make([]LeafTXO, nAdds)

	for j := range adds {
		adds[j].Hash[0] = uint8(j>>8) | 0xa0
		adds[j].Hash[1] = uint8(j)
		adds[j].Hash[3] = 0xaa
	}

	// add x
	_, err := f.Modify(adds, nil)
	if err != nil {
		return err
	}
	addHashes := make([]Hash, len(adds))
	for i, h := range adds {
		addHashes[i] = h.Hash
	}

	// get block proof
	bp, err := f.ProveBatch(addHashes[:nDels])
	if err != nil {
		return err
	}

	// check block proof.  Note this doesn't delete anything, just proves inclusion
	worked, _ := VerifyBatchProof(bp, f.GetTops(), f.numLeaves, f.height)
	//	worked := f.VerifyBatchProof(bp)

	if !worked {
		return fmt.Errorf("VerifyBatchProof failed")
	}
	fmt.Printf("VerifyBatchProof worked\n")
	return nil
}

func TestDeleteNonExisting(t *testing.T) {
	f := NewForest(nil)
	deletions := []uint64{0}
	_, err := f.Modify(nil, deletions)
	if err == nil {
		t.Fatal(fmt.Errorf(
			"shouldn't be able to delete non-existing leaf 0 from empty forest"))
	}
}
