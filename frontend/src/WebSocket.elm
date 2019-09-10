port module WebSocket exposing
    ( createWS
    , open
    , openWithProto
    , openWithProtos
    , send
    , subscriptions
    )

import Data exposing (Socket)
import Json.Decode as JD
import Json.Decode.Pipeline as JDP
import Json.Encode as JE
import Msgs exposing (Msg(..))


port createWS : SocketSpec -> Cmd msg


port sendWS : SocketData -> Cmd msg


port newFD : (JD.Value -> msg) -> Sub msg


port closeFD : (JD.Value -> msg) -> Sub msg


port recv : (JE.Value -> msg) -> Sub msg


type alias SocketData =
    { socket : Socket
    , data : JE.Value
    }


type alias SocketSpec =
    ( String, List String )


open : String -> Cmd msg
open url =
    createWS ( url, [] )


openWithProto : String -> String -> Cmd msg
openWithProto url proto =
    createWS ( url, [ proto ] )


openWithProtos : String -> List String -> Cmd msg
openWithProtos url protos =
    createWS ( url, protos )


send : Socket -> JE.Value -> Cmd msg
send s v =
    sendWS { socket = s, data = v }


decodeSocket : JD.Value -> Result JD.Error Socket
decodeSocket =
    JD.decodeValue
        (JD.succeed Socket
            |> JDP.required "url" JD.string
            |> JDP.required "fd" JD.int
        )


processNewFD : JD.Value -> Msg
processNewFD value =
    case decodeSocket value of
        Ok socket ->
            SocketOpened socket

        Err _ ->
            SocketNotOpened


processCloseFD : JD.Value -> Msg
processCloseFD value =
    case decodeSocket value of
        Ok socket ->
            SocketClosed socket

        Err _ ->
            SocketNotOpened


subscriptions : Sub Msg
subscriptions =
    Sub.batch
        [ newFD processNewFD
        , closeFD processCloseFD
        , recv SocketReceived
        ]
