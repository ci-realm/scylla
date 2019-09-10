module Github exposing (payloadDecoder)

import Data exposing (Build, LogLine, Organization)
import Iso8601
import Json.Decode as JD exposing (Decoder, andThen, at, int, list, string, succeed)
import Json.Decode.Pipeline exposing (required, requiredAt)
import Msgs exposing (Mutation(..))
import Time


payloadDecoder : Decoder Mutation
payloadDecoder =
    at [ "data", "mutation" ] string |> andThen payloadDecode


payloadDecode : String -> Decoder Mutation
payloadDecode mutation =
    case mutation of
        "lastBuilds" ->
            mapData "builds" MutateLastBuilds (list decodeBuild)

        "build" ->
            mapData "build" MutateBuild decodeBuild

        "organizations" ->
            mapData "organizations" MutateOrganizations (list decodeOrganization)

        "organizationBuilds" ->
            mapData "organizationBuilds" MutateOrganizationBuilds (list decodeBuild)

        "buildLog" ->
            JD.map MutateLogLine (at [ "data", "data" ] decodeLogLine)

        "restart" ->
            succeed MutateRestart

        _ ->
            JD.fail ("Invalid mutation type: " ++ mutation)


mapData : String -> (a -> b) -> Decoder a -> Decoder b
mapData key t decoder =
    JD.map t (at [ "data", "data", key ] decoder)


decodeOrganization : Decoder Organization
decodeOrganization =
    succeed Organization
        |> required "owner" string
        |> required "url" string
        |> required "buildCount" int


decodeBuild : Decoder Build
decodeBuild =
    succeed Build
        |> required "id" int
        |> required "status" string
        |> required "createdAt" decodeTime
        |> required "finishedAt" decodeTime
        |> requiredAt [ "hook", "pull_request", "head", "repo", "owner", "login" ] string
        |> requiredAt [ "hook", "pull_request", "head", "repo", "owner", "html_url" ] string
        |> requiredAt [ "hook", "pull_request", "head", "repo", "name" ] string
        |> requiredAt [ "hook", "pull_request", "head", "repo", "html_url" ] string
        |> requiredAt [ "hook", "pull_request", "head", "sha" ] string
        |> requiredAt [ "hook", "pull_request", "number" ] int
        |> requiredAt [ "hook", "pull_request", "title" ] string
        |> requiredAt [ "hook", "pull_request", "base", "ref" ] string
        |> requiredAt [ "hook", "pull_request", "user", "login" ] string
        |> Json.Decode.Pipeline.optional "log" (list decodeLogLine) []


decodeLogLine : Decoder LogLine
decodeLogLine =
    succeed LogLine
        |> required "buildId" int
        |> required "createdAt" decodeTime
        |> required "line" string


decodeTime : Decoder Time.Posix
decodeTime =
    string
        |> andThen
            (\iso ->
                succeed <|
                    case Iso8601.toTime iso of
                        Ok time ->
                            time

                        Err _ ->
                            Time.millisToPosix 0
            )
