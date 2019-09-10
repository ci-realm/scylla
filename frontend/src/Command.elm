module Command exposing
    ( buildLogUnwatch
    , buildLogWatch
    , buildRestart
    , getBuild
    , getLastBuilds
    , getOrganizationBuilds
    , getOrganizations
    , wsConnect
    )

import Data exposing (Socket)
import Json.Encode as JE
import WebSocket as WS


wsConnect : String -> Cmd msg
wsConnect url =
    WS.open url


buildRestart : Maybe Socket -> Int -> Cmd msg
buildRestart socket id =
    sendKind socket "restart" [ ( "id", JE.string (String.fromInt id) ) ]


buildLogWatch : Maybe Socket -> Int -> Cmd msg
buildLogWatch socket id =
    sendKind socket "buildLogWatch" [ ( "id", JE.string (String.fromInt id) ) ]


buildLogUnwatch : Maybe Socket -> Int -> Cmd msg
buildLogUnwatch socket id =
    sendKind socket "buildLogUnwatch" [ ( "id", JE.string (String.fromInt id) ) ]


getBuild : Maybe Socket -> Int -> Cmd msg
getBuild socket id =
    sendKind socket "build" [ ( "id", JE.string (String.fromInt id) ) ]


getLastBuilds : Maybe Socket -> Cmd msg
getLastBuilds socket =
    sendKind socket "lastBuilds" []


getOrganizationBuilds : Maybe Socket -> String -> Cmd msg
getOrganizationBuilds socket name =
    sendKind socket "organizationBuilds" [ ( "orgName", JE.string name ) ]


getOrganizations : Maybe Socket -> Cmd msg
getOrganizations socket =
    sendKind socket "organizations" []


sendKind : Maybe Socket -> String -> List ( String, JE.Value ) -> Cmd msg
sendKind socket kind data =
    send socket (JE.object [ ( "kind", JE.string kind ), ( "data", JE.object data ) ])


send : Maybe Socket -> JE.Value -> Cmd msg
send ms v =
    case ms of
        Nothing ->
            Cmd.none

        Just s ->
            WS.send s v
