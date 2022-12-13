package polybft

import (
	"bytes"
	"math"
	"testing"

	bls "github.com/0xPolygon/polygon-edge/consensus/polybft/signer"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposerCalculator_SetIndex(t *testing.T) {
	t.Parallel()

	validators := newTestValidatorsWithAliases([]string{"A", "B", "C", "D", "E"}, []uint64{10, 100, 1, 50, 30})
	metadata := validators.getPublicIdentities()

	vs, err := validators.toValidatorSet()
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(1, metadata), hclog.NewNullLogger())

	// validate no changes to validator set positions
	for i, v := range vs.Accounts() {
		assert.Equal(t, metadata[i].Address, v.Address)
	}

	proposer, err := pc.CalcProposer(0, 1)
	require.NoError(t, err)
	assert.Equal(t, proposer, metadata[1].Address)
	// validate no changes to validator set positions
	for i, v := range vs.Accounts() {
		assert.Equal(t, metadata[i].Address, v.Address)
	}
}

func TestProposerCalculator_CalculateProposer(t *testing.T) {
	t.Parallel()

	validators := newTestValidatorsWithAliases([]string{"A", "B", "C", "D", "E"}, []uint64{1, 2, 3, 4, 5})
	metadata := validators.getPublicIdentities()

	vs, err := validators.toValidatorSet()
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, metadata), hclog.NewNullLogger())

	assert.Equal(t, int64(15), vs.totalVotingPower)

	currProposerAddress, err := pc.CalcProposer(0, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[4].Address, currProposerAddress)

	proposerAddressR1, err := pc.CalcProposer(1, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[3].Address, proposerAddressR1)

	proposerAddressR2, err := pc.CalcProposer(2, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[2].Address, proposerAddressR2)

	proposerAddressR3, err := pc.CalcProposer(3, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[1].Address, proposerAddressR3)

	proposerAddressR4, err := pc.CalcProposer(4, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[4].Address, proposerAddressR4)

	proposerAddressR5, err := pc.CalcProposer(5, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[3].Address, proposerAddressR5)

	proposerAddressR6, err := pc.CalcProposer(6, 0)
	require.NoError(t, err)
	assert.Equal(t, metadata[0].Address, proposerAddressR6)
}

func TestProposerCalculator_SamePriority(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(5)
	require.NoError(t, err)

	// at some point priorities will be the same and bytes address will be compared
	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 2,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 3,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())

	proposerR0, err := pc.CalcProposer(0, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x3}, proposerR0)

	proposerR1, err := pc.CalcProposer(1, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerR1)

	proposerR2, err := pc.CalcProposer(2, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerR2)

	proposerR2, err = pc.CalcProposer(2, 0) // call again same round
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerR2)
}

func TestProposerCalculator_ProposerSelection1(t *testing.T) {
	t.Parallel()

	const numberOfIteration = 99

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	vset, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1000,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 300,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 330,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(4, vset.Accounts()), hclog.NewNullLogger())

	var proposers = make([]types.Address, numberOfIteration)

	for i := uint64(0); i < numberOfIteration; i++ {
		proposers[i], err = pc.CalcProposer(i, 4)
		require.NoError(t, err)
	}

	expected := []types.Address{
		{0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1},
		{0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1},
		{0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3},
		{0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x1}, {0x3}, {0x2}, {0x1}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2},
		{0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1},
		{0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1},
		{0x3}, {0x1}, {0x1},
	}

	for i, p := range proposers {
		assert.True(t, bytes.Equal(expected[i].Bytes(), p.Bytes()))
	}
}

// Test that Update requires correct height
// func TestProposerCalculator_UpdateWrongHeight(t *testing.T) {
// 	t.Parallel()

// 	validators := newTestValidatorsWithAliases([]string{"A", "B", "C"}, []uint64{1000, 300, 330})
// 	metadata := validators.getPublicIdentities()

// 	pc, err := NewProposerCalculator(NewProposerCalculatorSnapshot(0, metadata), hclog.NewNullLogger())
// 	require.NoError(t, err)

// 	proposerAddressR0, err := pc.CalcProposer(0, 0)
// 	assert.Equal(t, metadata[0].Address, proposerAddressR0)

// 	err = pc.updateToBlock(0, 0, nil)
// 	require.Error(t, err)

// 	err = pc.Update(0, 2, nil)
// 	require.Error(t, err)

// 	err = pc.Update(0, 1, nil)
// 	require.NoError(t, err)
// }

func TestProposerCalculator_IncrementProposerPrioritySameVotingPower(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 1,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)
	assert.Equal(t, int64(3), vs.totalVotingPower)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())

	// when voting power is the same order is by address
	currProposerAddress, err := pc.CalcProposer(0, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, currProposerAddress)

	proposerAddresR1, err := pc.CalcProposer(1, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerAddresR1)

	proposerAddressR2, err := pc.CalcProposer(2, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x3}, proposerAddressR2)

	proposerAddressR3, err := pc.CalcProposer(3, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerAddressR3)

	proposerAddressR4, err := pc.CalcProposer(4, 0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerAddressR4)
}

func TestProposerCalculator_AveragingInIncrementProposerPriorityWithVotingPower(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	// Other than TestAveragingInIncrementProposerPriority this is a more complete test showing
	// how each ProposerPriority changes in relation to the validator's voting power respectively.
	// average is zero in each round:
	vp0 := int64(10)
	vp1 := int64(1)
	vp2 := int64(1)
	total := vp0 + vp1 + vp2
	avg := (vp0 + vp1 + vp2 - total) / 3
	valz := []*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: uint64(vp0),
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: uint64(vp1),
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: uint64(vp2),
		},
	}

	tcs := []struct {
		wantProposerPrioritys []int64
		times                 uint64
		wantProposerIndex     int64
	}{

		0: {
			[]int64{
				// Acumm+VotingPower-Avg:
				0 + vp0 - total - avg, // mostest will be subtracted by total voting power (12)
				0 + vp1,
				0 + vp2},
			1,
			0,
		},
		1: {
			[]int64{
				(0 + vp0 - total) + vp0 - total - avg, // this will be mostest on 2nd iter, too
				(0 + vp1) + vp1,
				(0 + vp2) + vp2},
			2,
			0,
		}, // increment twice -> expect average to be subtracted twice
		2: {
			[]int64{
				0 + 3*(vp0-total) - avg, // still mostest
				0 + 3*vp1,
				0 + 3*vp2},
			3,
			0,
		},
		3: {
			[]int64{
				0 + 4*(vp0-total), // still mostest
				0 + 4*vp1,
				0 + 4*vp2},
			4,
			0,
		},
		4: {
			[]int64{
				0 + 4*(vp0-total) + vp0, // 4 iters was mostest
				0 + 5*vp1 - total,       // now this val is mostest for the 1st time (hence -12==totalVotingPower)
				0 + 5*vp2},
			5,
			1,
		},
		5: {
			[]int64{
				0 + 6*vp0 - 5*total, // mostest again
				0 + 6*vp1 - total,   // mostest once up to here
				0 + 6*vp2},
			6,
			0,
		},
		6: {
			[]int64{
				0 + 7*vp0 - 6*total, // in 7 iters this val is mostest 6 times
				0 + 7*vp1 - total,   // in 7 iters this val is mostest 1 time
				0 + 7*vp2},
			7,
			0,
		},
		7: {
			[]int64{
				0 + 8*vp0 - 7*total, // mostest again
				0 + 8*vp1 - total,
				0 + 8*vp2},
			8,
			0,
		},
		8: {
			[]int64{
				0 + 9*vp0 - 7*total,
				0 + 9*vp1 - total,
				0 + 9*vp2 - total}, // mostest
			9,
			2,
		},
		9: {
			[]int64{
				0 + 10*vp0 - 8*total, // after 10 iters this is mostest again
				0 + 10*vp1 - total,   // after 6 iters this val is "mostest" once and not in between
				0 + 10*vp2 - total},  // in between 10 iters this val is "mostest" once
			10,
			0,
		},
		10: {
			[]int64{
				0 + 11*vp0 - 9*total,
				0 + 11*vp1 - total,  // after 6 iters this val is "mostest" once and not in between
				0 + 11*vp2 - total}, // after 10 iters this val is "mostest" once
			11,
			0,
		},
	}

	for i, tc := range tcs {
		snap := NewProposerCalculatorSnapshot(1, valz)
		pc := NewProposerCalculator(snap, hclog.NewNullLogger())

		require.NoError(t, pc.incrementProposerPriorityNTimes(tc.times))

		address, _ := pc.GetLatestProposer(0, 1)

		assert.Equal(t, pc.snapshot.Validators[tc.wantProposerIndex].Metadata.Address, address,
			"test case: %v",
			i)

		for valIdx, val := range pc.snapshot.Validators {
			assert.Equal(t,
				tc.wantProposerPrioritys[valIdx],
				val.ProposerPriority,
				"test case: %v, validator: %v",
				i,
				valIdx)
		}
	}
}

func TestProposerCalculator_TotalVotingPowerErrorOnOverflow(t *testing.T) {
	t.Parallel()

	// NewValidatorSet calls IncrementProposerPriority which calls TotalVotingPower()
	// which should return error on overflows:
	_, err := NewValidatorSet([]*ValidatorMetadata{
		{Address: types.Address{0x1}, VotingPower: math.MaxInt64},
		{Address: types.Address{0x2}, VotingPower: math.MaxInt64},
		{Address: types.Address{0x3}, VotingPower: math.MaxInt64},
	}, hclog.NewNullLogger())
	require.Error(t, err)
}

func TestProposerCalculator_UpdatesForNewValidatorSet(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(2)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 100}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 100}

	accountSet := []*ValidatorMetadata{v1, v2}
	vs, err := NewValidatorSet(accountSet, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())

	_, err = pc.CalcProposer(1, 0)
	require.NoError(t, err)

	// verify that the capacity and length of validators is the same
	assert.Equal(t, len(vs.Accounts()), cap(pc.snapshot.Validators))
	// verify that validator priorities are centered
	valsCount := int64(len(pc.snapshot.Validators))

	sum := int64(0)
	for _, val := range pc.snapshot.Validators {
		// mind overflow
		sum = safeAddClip(sum, val.ProposerPriority)
	}

	assert.True(t, sum < valsCount && sum > -valsCount,
		"expected total priority in (-%d, %d). Got %d", valsCount, valsCount, sum)
	// verify that priorities are scaled
	dist := computeMaxMinPriorityDiff(pc.snapshot.Validators)
	assert.True(t, dist <= priorityWindowSizeFactor*vs.totalVotingPower,
		"expected priority distance < %d. Got %d", priorityWindowSizeFactor*vs.totalVotingPower, dist)
}

func TestProposerCalculator_GetLatestProposer(t *testing.T) {
	const (
		bestIdx = 5
		count   = 10
	)

	t.Parallel()

	validatorSet := newTestValidators(count).getPublicIdentities()
	snap := NewProposerCalculatorSnapshot(0, validatorSet)

	proposerCalculator := NewProposerCalculator(snap, hclog.NewNullLogger())

	snap.Validators[bestIdx].ProposerPriority = 1000000

	// not set
	_, exists := proposerCalculator.GetLatestProposer(0, 0)
	assert.False(t, exists)

	_, err := proposerCalculator.CalcProposer(0, 0)
	require.NoError(t, err)

	// wrong round
	_, exists = proposerCalculator.GetLatestProposer(1, 0)
	assert.False(t, exists)

	// ok
	address, exists := proposerCalculator.GetLatestProposer(0, 0)
	assert.True(t, exists)
	assert.Equal(t, validatorSet[bestIdx].Address, address)
}

func TestProposerCalculator_UpdateValidatorsSameVpUpdatedAndNewAdded(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(8)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 100}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 100}
	v3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 100}
	v4 := &ValidatorMetadata{Address: types.Address{0x4}, BlsKey: keys[3].PublicKey(), VotingPower: 100}
	v5 := &ValidatorMetadata{Address: types.Address{0x5}, BlsKey: keys[4].PublicKey(), VotingPower: 100}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2, v3, v4, v5}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())

	// iterate one cycle should bring back priority to 0
	require.NoError(t, pc.incrementProposerPriorityNTimes(5))

	for _, v := range pc.snapshot.Validators {
		assert.Equal(t, int64(0), v.ProposerPriority)
	}

	// updated old validators
	u1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[1].PublicKey(), VotingPower: 10}
	u2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[2].PublicKey(), VotingPower: 10}
	// added new validator
	a1 := &ValidatorMetadata{Address: types.Address{0x9}, BlsKey: keys[7].PublicKey(), VotingPower: 100}

	newAccountSet := []*ValidatorMetadata{u1, u2, a1}

	err = pc.updateValidators(newAccountSet)
	require.NoError(t, err)

	assert.Equal(t, 3, len(pc.snapshot.Validators))

	// removedVp := sum(v3, v4, v5) = 300
	// newVp := sum(u1, u2, a1) = 120
	// sum(removedVp, newVp) = 420; priority(a1) = -1.125*420 = -472
	// scale: difMax = 2 * 120; diff(-475, 0); ratio ~ 2
	// priority(a1) = -472/2 = 236; u1 = 0, u2 = 0
	// center: avg = 236/3 = 79; priority(a1)= 236 - 79

	// check voting power after update
	assert.Equal(t, uint64(10), pc.snapshot.Validators[0].Metadata.VotingPower)
	assert.Equal(t, uint64(10), pc.snapshot.Validators[1].Metadata.VotingPower)
	assert.Equal(t, uint64(100), pc.snapshot.Validators[2].Metadata.VotingPower)
	// newly added validator
	assert.Equal(t, uint64(100), pc.snapshot.Validators[2].Metadata.VotingPower)
	assert.Equal(t, types.Address{0x9}, pc.snapshot.Validators[2].Metadata.Address)
	assert.Equal(t, int64(-157), pc.snapshot.Validators[2].ProposerPriority) // a1
	// check priority
	assert.Equal(t, int64(79), pc.snapshot.Validators[0].ProposerPriority) // u1
	assert.Equal(t, int64(79), pc.snapshot.Validators[1].ProposerPriority) // u2

	err = pc.incrementProposerPriorityNTimes(1)
	require.NoError(t, err)

	// 79 + 10 - (100+10+10)
	assert.Equal(t, int64(-31), pc.snapshot.Validators[0].ProposerPriority)
	// 79 + 10
	assert.Equal(t, int64(89), pc.snapshot.Validators[1].ProposerPriority)
	// -157+100
	assert.Equal(t, int64(-57), pc.snapshot.Validators[2].ProposerPriority)
}

func TestProposerCalculator_UpdateValidators(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(4)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 10}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 20}
	v3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 30}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2, v3}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())
	require.Equal(t, int64(60), pc.totalVotingPower)
	// 	init priority must be 0
	require.Zero(t, pc.snapshot.Validators[0].ProposerPriority)
	require.Zero(t, pc.snapshot.Validators[1].ProposerPriority)
	require.Zero(t, pc.snapshot.Validators[2].ProposerPriority)
	// vp must be initialized
	require.Equal(t, uint64(10), pc.snapshot.Validators[0].Metadata.VotingPower)
	require.Equal(t, uint64(20), pc.snapshot.Validators[1].Metadata.VotingPower)
	require.Equal(t, uint64(30), pc.snapshot.Validators[2].Metadata.VotingPower)

	// increment once
	err = pc.incrementProposerPriorityNTimes(1)
	require.NoError(t, err)

	// updated
	u1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 100}
	u2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 200}
	u3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 300}
	// added
	a1 := &ValidatorMetadata{Address: types.Address{0x4}, BlsKey: keys[3].PublicKey(), VotingPower: 400}

	err = pc.updateValidators([]*ValidatorMetadata{u1, u2, u3, a1})
	require.NoError(t, err)

	require.Equal(t, 4, len(pc.snapshot.Validators))
	// priorities are from previous iteration
	require.Equal(t, int64(292), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(302), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(252), pc.snapshot.Validators[2].ProposerPriority)
	// new added a1
	require.Equal(t, types.Address{0x4}, pc.snapshot.Validators[3].Metadata.Address)
	require.Equal(t, int64(-843), pc.snapshot.Validators[3].ProposerPriority)
	// total vp is updated
	require.Equal(t, int64(1000), pc.totalVotingPower)
}

func TestProposerCalculator_ScaleAfterDelete(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 10}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 10}
	v3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 80000}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2, v3}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())
	assert.Equal(t, int64(80020), pc.totalVotingPower)

	err = pc.incrementProposerPriorityNTimes(1)
	require.NoError(t, err)

	// priorities are from previous iteration
	require.Equal(t, int64(10), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(10), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(-20), pc.snapshot.Validators[2].ProposerPriority)

	// another increment
	err = pc.incrementProposerPriorityNTimes(4000)
	require.NoError(t, err)
	// priorities are from previous iteration
	assert.Equal(t, types.Address{0x3}, pc.proposer.Metadata.Address)

	// 	reduce validator voting power from 8k to 1
	u1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 10}
	u2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 10}

	require.Equal(t, int64(-40010), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(40010), pc.snapshot.Validators[1].ProposerPriority)

	err = pc.updateValidators([]*ValidatorMetadata{u1, u2})
	require.NoError(t, err)

	// maxdiff = 2*tvp = 40
	// diff(min,max) (-40010, 40010) = 80020
	// ratio := (diff + diffMax - 1) / diffMax; (80020 + 20 - 1)/20 = 2001
	// priority = priority / ratio; u1 = -40010 / 4001 ~ -19; u2 = 40010 / 4001 ~ 19
	require.Equal(t, int64(-19), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(19), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(20), pc.totalVotingPower)
}

func TestProposerCalculator_ShiftAfterUpdate(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 50}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 80}
	v3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 100000}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2, v3}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())
	assert.Equal(t, int64(100130), pc.totalVotingPower)

	err = pc.incrementProposerPriorityNTimes(4000)
	require.NoError(t, err)

	// updates
	u1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 5}
	u2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 8}

	err = pc.updateValidators([]*ValidatorMetadata{u1, u2})
	require.NoError(t, err)

	// maxdiff = 2*tvp = 26
	// diff(min,max) (-260, 19610) = 19870
	// ratio := (diff + diffMax - 1) / diffMax; (19870 + 26 - 1)/26 =765
	// scale priority = priority / ratio; p1 = 0; p2 = 25
	// shift with avg=(25+0)/2=12; p = priority - avg; u1 = -12; u2= 13
	require.Equal(t, int64(-12), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(13), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(13), pc.totalVotingPower)
}

func TestProposerCalculator_UpdateValidatorSet(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 1}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 8}
	v3 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 15}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2, v3}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())
	assert.Equal(t, int64(24), pc.totalVotingPower)

	err = pc.incrementProposerPriorityNTimes(2)
	require.NoError(t, err)

	// modified validator
	u1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 5}
	// added validator
	a1 := &ValidatorMetadata{Address: types.Address{0x4}, BlsKey: keys[1].PublicKey(), VotingPower: 8}

	err = pc.updateValidators([]*ValidatorMetadata{u1, a1})
	require.NoError(t, err)
	// expecting 2 validators with updated voting power and total voting power
	require.Equal(t, 2, len(pc.snapshot.Validators))
	require.Equal(t, types.Address{0x1}, pc.snapshot.Validators[0].Metadata.Address)
	require.Equal(t, uint64(5), pc.snapshot.Validators[0].Metadata.VotingPower)
	require.Equal(t, int64(11), pc.snapshot.Validators[0].ProposerPriority)

	require.Equal(t, types.Address{0x4}, pc.snapshot.Validators[1].Metadata.Address)
	require.Equal(t, uint64(8), pc.snapshot.Validators[1].Metadata.VotingPower)
	require.Equal(t, int64(-10), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(13), pc.totalVotingPower)
}

func TestProposerCalculator_AddValidator(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, BlsKey: keys[0].PublicKey(), VotingPower: 3}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, BlsKey: keys[1].PublicKey(), VotingPower: 1}

	vs, err := NewValidatorSet([]*ValidatorMetadata{v1, v2}, hclog.NewNullLogger())
	require.NoError(t, err)

	pc := NewProposerCalculator(NewProposerCalculatorSnapshot(0, vs.Accounts()), hclog.NewNullLogger())
	assert.Equal(t, int64(4), pc.totalVotingPower)
	err = pc.incrementProposerPriorityNTimes(1)
	require.NoError(t, err)
	require.Equal(t, types.Address{0x1}, pc.proposer.Metadata.Address)
	require.Equal(t, int64(-1), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(1), pc.snapshot.Validators[1].ProposerPriority)

	err = pc.incrementProposerPriorityNTimes(1)
	require.NoError(t, err)

	require.Equal(t, int64(-2), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(2), pc.snapshot.Validators[1].ProposerPriority)

	a1 := &ValidatorMetadata{Address: types.Address{0x3}, BlsKey: keys[2].PublicKey(), VotingPower: 8}

	err = pc.updateValidators([]*ValidatorMetadata{v1, v2, a1})
	require.NoError(t, err)

	// updated vp: 8+3+1 = 12
	// added validator priority = -1.125*8 ~ -13
	// scaling: max(-13, 3) = 16 < 2* 12; no scaling
	// centring: avg = (13+3+1)/3=5; v1=-2+5, v2=2+5; u3=-13+5
	require.Equal(t, int64(3), pc.snapshot.Validators[0].ProposerPriority)
	require.Equal(t, int64(7), pc.snapshot.Validators[1].ProposerPriority)
	require.Equal(t, int64(-8), pc.snapshot.Validators[2].ProposerPriority)
}