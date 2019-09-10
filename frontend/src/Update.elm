module Update exposing (update)

import Browser
import Browser.Navigation as Nav
import Cmd.Extra exposing (withCmd, withNoCmd)
import Command exposing (buildLogUnwatch, buildLogWatch, buildRestart, getLastBuilds, getOrganizationBuilds, getOrganizations, wsConnect)
import Data exposing (LogLine)
import Debounce
import Github exposing (payloadDecoder)
import Init exposing (urlChanged)
import Json.Decode as JD
import Model exposing (Model)
import Msgs exposing (Msg(..), Mutation(..), Route(..))
import Return exposing (Return)
import Url
import Utils exposing (scrollIntoView)


update : Msg -> Model -> Return Msg Model
update message model =
    case message of
        Tick newTime ->
            { model | time = newTime } |> withNoCmd

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    model |> withCmd (Nav.pushUrl model.key (Url.toString url))

                Browser.External href ->
                    model |> withCmd (Nav.load href)

        UrlChanged url ->
            urlChanged { model | url = url }

        SocketDebounce msg_ ->
            let
                ( debounce, cmd ) =
                    Debounce.update debounceConfig (Debounce.takeLast wsConnect) msg_ model.socketDebounce
            in
            { model | socketDebounce = debounce } |> withCmd cmd

        SocketOpened newsocket ->
            urlChanged { model | socket = Just newsocket }

        SocketClosed oldSocket ->
            let
                ( debounce, cmd ) =
                    Debounce.push debounceConfig oldSocket.url model.socketDebounce
            in
            { model | socketDebounce = debounce } |> withCmd cmd

        SocketNotOpened ->
            model |> withNoCmd

        SocketReceived value ->
            case JD.decodeValue Github.payloadDecoder value of
                Ok decoded ->
                    case decoded of
                        MutateLastBuilds obj ->
                            { model | lastBuilds = obj } |> withNoCmd

                        MutateRestart ->
                            model |> withNoCmd

                        MutateBuild obj ->
                            { model | build = Just obj } |> withNoCmd

                        MutateOrganizations obj ->
                            { model | organizations = obj } |> withNoCmd

                        MutateOrganizationBuilds obj ->
                            { model | organizationBuilds = obj } |> withNoCmd

                        MutateLogLine obj ->
                            appendLog model obj |> withCmd (scrollIntoView "last-log-line")

                Err error ->
                    { model | error = Just (JD.errorToString error) } |> withNoCmd

        GetLastBuilds ->
            model |> withCmd (getLastBuilds model.socket)

        GetOrganizations ->
            model |> withCmd (getOrganizations model.socket)

        GetOrganizationBuilds name ->
            model |> withCmd (getOrganizationBuilds model.socket name)

        BuildLogWatch id ->
            { model | buildLogPaused = False } |> withCmd (buildLogWatch model.socket id)

        BuildLogUnwatch id ->
            { model | buildLogPaused = True } |> withCmd (buildLogUnwatch model.socket id)

        BuildRestart id ->
            model |> withCmd (buildRestart model.socket id)


debounceConfig : Debounce.Config Msg
debounceConfig =
    { strategy = Debounce.later 10000
    , transform = SocketDebounce
    }


appendLog : Model -> LogLine -> Model
appendLog model line =
    case model.build of
        Just build ->
            { model | build = Just { build | log = build.log ++ [ line ] } }

        Nothing ->
            model
