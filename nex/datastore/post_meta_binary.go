package nex_datastore

import (
	"github.com/PretendoNetwork/team-kirby-clash-deluxe-secure/database"
	"github.com/PretendoNetwork/team-kirby-clash-deluxe-secure/globals"

	"github.com/PretendoNetwork/nex-go"
	datastore "github.com/PretendoNetwork/nex-protocols-go/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/datastore/types"
)

func PostMetaBinary(err error, client *nex.Client, callID uint32, param *datastore_types.DataStorePreparePostParam) uint32 {
	if err != nil {
		globals.Logger.Error(err.Error())
		return nex.Errors.DataStore.InvalidArgument
	}

	metaBinary := database.GetMetaBinaryByOwnerPID(client.PID())

	// * Meta binary already exists
	if metaBinary.DataID != 0 && param.PersistenceInitParam.DeleteLastObject {
		// * Delete existing object before uploading new one
		err = database.DeleteMetaBinaryByDataID(metaBinary.DataID)
		if err != nil {
			globals.Logger.Critical(err.Error())
			return nex.Errors.DataStore.Unknown
		}
	}

	dataID := database.InsertMetaBinaryByDataStorePreparePostParamWithOwnerPID(param, client.PID())

	rmcResponseStream := nex.NewStreamOut(globals.NEXServer)

	rmcResponseStream.WriteUInt64LE(uint64(dataID))

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCResponse(datastore.ProtocolID, callID)
	rmcResponse.SetSuccess(datastore.MethodPreparePostObject, rmcResponseBody)

	rmcResponseBytes := rmcResponse.Bytes()

	responsePacket, _ := nex.NewPacketV1(client, nil)

	responsePacket.SetVersion(1)
	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	globals.NEXServer.Send(responsePacket)

	return 0
}
