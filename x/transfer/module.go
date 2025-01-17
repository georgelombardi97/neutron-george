package transfer

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/internal/sudo"
)

/*
	In addition to original ack processing of ibc transfer acknowledgement we want to pass the acknowledgement to originating wasm contract.
	The package contains a code to achieve the purpose.
*/

type IBCModule struct {
	keeper      keeper.Keeper
	sudoHandler sudo.Handler
	transfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper, wasmKeeper *wasm.Keeper) IBCModule {
	return IBCModule{
		keeper:      k,
		IBCModule:   transfer.NewIBCModule(k),
		sudoHandler: sudo.NewSudoHandler(wasmKeeper, types.ModuleName),
	}
}

// OnAcknowledgementPacket implements the IBCModule interface.
// Wrapper struct shadows(overrides) the OnAcknowledgementPacket method to achieve the package's purpose.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to process original OnAcknowledgementPacket")
	}
	return im.HandleAcknowledgement(ctx, packet, acknowledgement)

}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to process original OnTimeoutPacket")
	}
	return im.HandleTimeout(ctx, packet)
}
