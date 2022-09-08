package wasmbinding

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func CustomMessageDecorator(ictx *ictxkeeper.Keeper, icq *icqkeeper.Keeper) func(messenger wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			Keeper:        *ictx,
			Wrapped:       old,
			Ictxmsgserver: ictxkeeper.NewMsgServerImpl(*ictx),
			Icqmsgserver:  icqkeeper.NewMsgServerImpl(*icq),
		}
	}
}

type CustomMessenger struct {
	Keeper        ictxkeeper.Keeper
	Wrapped       wasmkeeper.Messenger
	Ictxmsgserver ictxtypes.MsgServer
	Icqmsgserver  icqtypes.MsgServer
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		var contractMsg bindings.NeutronMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			ctx.Logger().Debug("json.Unmarshal: failed to decode incoming custom cosmos message",
				"from_address", contractAddr.String(),
				"message", string(msg.Custom),
				"error", err,
			)
			return nil, nil, sdkerrors.Wrap(err, "failed to decode incoming custom cosmos message")
		}

		if contractMsg.SubmitTx != nil {
			return m.submitTx(ctx, contractAddr, contractMsg.SubmitTx)
		}
		if contractMsg.RegisterInterchainAccount != nil {
			return m.registerInterchainAccount(ctx, contractAddr, contractMsg.RegisterInterchainAccount)
		}
		if contractMsg.RegisterInterchainQuery != nil {
			return m.registerInterchainQuery(ctx, contractAddr, contractMsg.RegisterInterchainQuery)
		}
		if contractMsg.UpdateInterchainQuery != nil {
			return m.updateInterchainQuery(ctx, contractAddr, contractMsg.UpdateInterchainQuery)
		}
		if contractMsg.RemoveInterchainQuery != nil {
			return m.removeInterchainQuery(ctx, contractAddr, contractMsg.RemoveInterchainQuery)
		}
	}

	return m.Wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *CustomMessenger) updateInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, updateQuery *bindings.UpdateInterchainQuery) ([]sdk.Event, [][]byte, error) {
	response, err := m.performUpdateInterchainQuery(ctx, contractAddr, updateQuery)
	if err != nil {
		ctx.Logger().Debug("performUpdateInterchainQuery: failed to update interchain query",
			"from_address", contractAddr.String(),
			"msg", updateQuery,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "failed to update interchain query")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal UpdateInterchainQueryResponse response to JSON",
			"from_address", contractAddr.String(),
			"msg", updateQuery,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("interchain query updated",
		"from_address", contractAddr.String(),
		"msg", updateQuery,
	)
	return nil, [][]byte{data}, nil
}

func (m *CustomMessenger) performUpdateInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, updateQuery *bindings.UpdateInterchainQuery) (*bindings.UpdateInterchainQueryResponse, error) {
	msg := icqtypes.MsgUpdateInterchainQueryRequest{
		QueryId:         updateQuery.QueryId,
		NewKeys:         updateQuery.NewKeys,
		NewUpdatePeriod: updateQuery.NewUpdatePeriod,
		Sender:          contractAddr.String(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate incoming UpdateInterchainQuery message")
	}

	response, err := m.Icqmsgserver.UpdateInterchainQuery(sdk.WrapSDKContext(ctx), &msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to update interchain query")
	}

	return (*bindings.UpdateInterchainQueryResponse)(response), nil
}

func (m *CustomMessenger) removeInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, removeQuery *bindings.RemoveInterchainQuery) ([]sdk.Event, [][]byte, error) {
	response, err := m.performRemoveInterchainQuery(ctx, contractAddr, removeQuery)
	if err != nil {
		ctx.Logger().Debug("performRemoveInterchainQuery: failed to update interchain query",
			"from_address", contractAddr.String(),
			"msg", removeQuery,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "failed to remove interchain query")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal RemoveInterchainQueryResponse response to JSON",
			"from_address", contractAddr.String(),
			"msg", removeQuery,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("interchain query removed",
		"from_address", contractAddr.String(),
		"msg", removeQuery,
	)
	return nil, [][]byte{data}, nil
}

func (m *CustomMessenger) performRemoveInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, updateQuery *bindings.RemoveInterchainQuery) (*bindings.RemoveInterchainQueryResponse, error) {
	msg := icqtypes.MsgRemoveInterchainQueryRequest{
		QueryId: updateQuery.QueryId,
		Sender:  contractAddr.String(),
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate incoming RemoveInterchainQuery message")
	}

	response, err := m.Icqmsgserver.RemoveInterchainQuery(sdk.WrapSDKContext(ctx), &msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to remove interchain query")
	}

	return (*bindings.RemoveInterchainQueryResponse)(response), nil
}

func (m *CustomMessenger) submitTx(ctx sdk.Context, contractAddr sdk.AccAddress, submitTx *bindings.SubmitTx) ([]sdk.Event, [][]byte, error) {
	response, err := m.PerformSubmitTx(ctx, contractAddr, submitTx)
	if err != nil {
		ctx.Logger().Debug("PerformSubmitTx: failed to submit interchain transaction",
			"from_address", contractAddr.String(),
			"connection_id", submitTx.ConnectionId,
			"interchain_account_id", submitTx.InterchainAccountId,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "failed to submit interchain transaction")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal submitTx response to JSON",
			"from_address", contractAddr.String(),
			"connection_id", submitTx.ConnectionId,
			"interchain_account_id", submitTx.InterchainAccountId,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("interchain transaction submitted",
		"from_address", contractAddr.String(),
		"connection_id", submitTx.ConnectionId,
		"interchain_account_id", submitTx.InterchainAccountId,
	)
	return nil, [][]byte{data}, nil
}

func (m *CustomMessenger) PerformSubmitTx(ctx sdk.Context, contractAddr sdk.AccAddress, submitTx *bindings.SubmitTx) (*bindings.SubmitTxResponse, error) {
	tx := ictxtypes.MsgSubmitTx{
		FromAddress:         contractAddr.String(),
		ConnectionId:        submitTx.ConnectionId,
		Memo:                submitTx.Memo,
		InterchainAccountId: submitTx.InterchainAccountId,
		Timeout:             submitTx.Timeout,
	}
	for _, msg := range submitTx.Msgs {
		tx.Msgs = append(tx.Msgs, &types.Any{
			TypeUrl: msg.TypeURL,
			Value:   msg.Value,
		})
	}
	//if err := tx.UnpackInterfaces(m.Keeper.Codec); err != nil {
	//	return nil, sdkerrors.Wrap(err, "failed to unpack interfaces to send interchain transaction")
	//}

	if err := tx.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate incoming SubmitTx message")
	}

	response, err := m.Ictxmsgserver.SubmitTx(sdk.WrapSDKContext(ctx), &tx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to submit interchain transaction")
	}

	return (*bindings.SubmitTxResponse)(response), nil
}

func (m *CustomMessenger) registerInterchainAccount(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainAccount) ([]sdk.Event, [][]byte, error) {
	response, err := m.PerformRegisterInterchainAccount(ctx, contractAddr, reg)
	if err != nil {
		ctx.Logger().Debug("PerformRegisterInterchainAccount: failed to register interchain account",
			"from_address", contractAddr.String(),
			"connection_id", reg.ConnectionId,
			"interchain_account_id", reg.InterchainAccountId,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "failed to register interchain account")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal register interchain account response to JSON",
			"from_address", contractAddr.String(),
			"connection_id", reg.ConnectionId,
			"interchain_account_id", reg.InterchainAccountId,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("registered interchain account",
		"from_address", contractAddr.String(),
		"connection_id", reg.ConnectionId,
		"interchain_account_id", reg.InterchainAccountId,
	)
	return nil, [][]byte{data}, nil
}

func (m *CustomMessenger) PerformRegisterInterchainAccount(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainAccount) (*bindings.RegisterInterchainAccountResponse, error) {
	msg := ictxtypes.MsgRegisterInterchainAccount{
		FromAddress:         contractAddr.String(),
		ConnectionId:        reg.ConnectionId,
		InterchainAccountId: reg.InterchainAccountId,
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate incoming RegisterInterchainAccount message")
	}

	response, err := m.Ictxmsgserver.RegisterInterchainAccount(sdk.WrapSDKContext(ctx), &msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to register interchain account")
	}

	return (*bindings.RegisterInterchainAccountResponse)(response), nil
}

func (m *CustomMessenger) registerInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainQuery) ([]sdk.Event, [][]byte, error) {
	response, err := m.PerformRegisterInterchainQuery(ctx, contractAddr, reg)
	if err != nil {
		ctx.Logger().Debug("PerformRegisterInterchainQuery: failed to register interchain query",
			"from_address", contractAddr.String(),
			"query_type", reg.QueryType,
			"kv_keys", icqtypes.KVKeys(reg.Keys).String(),
			"transactions_filter", reg.TransactionsFilter,
			"zone_id", reg.ZoneId,
			"connection_id", reg.ConnectionId,
			"update_period", reg.UpdatePeriod,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "failed to register interchain query")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal register interchain query response to JSON",
			"from_address", contractAddr.String(),
			"kv_keys", icqtypes.KVKeys(reg.Keys).String(),
			"transactions_filter", reg.TransactionsFilter,
			"zone_id", reg.ZoneId,
			"connection_id", reg.ConnectionId,
			"update_period", reg.UpdatePeriod,
			"error", err,
		)
		return nil, nil, sdkerrors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("registered interchain query",
		"from_address", contractAddr.String(),
		"query_type", reg.QueryType,
		"kv_keys", icqtypes.KVKeys(reg.Keys).String(),
		"transactions_filter", reg.TransactionsFilter,
		"zone_id", reg.ZoneId,
		"connection_id", reg.ConnectionId,
		"update_period", reg.UpdatePeriod,
		"query_id", response.Id,
	)
	return nil, [][]byte{data}, nil
}

func (m *CustomMessenger) PerformRegisterInterchainQuery(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainQuery) (*bindings.RegisterInterchainQueryResponse, error) {
	msg := icqtypes.MsgRegisterInterchainQuery{
		Keys:               reg.Keys,
		TransactionsFilter: reg.TransactionsFilter,
		QueryType:          reg.QueryType,
		ZoneId:             reg.ZoneId,
		ConnectionId:       reg.ConnectionId,
		UpdatePeriod:       reg.UpdatePeriod,
		Sender:             contractAddr.String(),
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to validate incoming RegisterInterchainQuery message")
	}

	response, err := m.Icqmsgserver.RegisterInterchainQuery(sdk.WrapSDKContext(ctx), &msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to register interchain query")
	}

	return (*bindings.RegisterInterchainQueryResponse)(response), nil
}
