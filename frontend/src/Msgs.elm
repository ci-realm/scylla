module Msgs exposing (Msg(..), Mutation(..), Route(..))

import Browser exposing (UrlRequest)
import Data exposing (Build, LogLine, Organization, Socket)
import Debounce
import Json.Encode as JE
import Time
import Url exposing (Url)


type Msg
    = SocketOpened Socket
    | SocketClosed Socket
    | SocketNotOpened
    | SocketReceived JE.Value
    | SocketDebounce Debounce.Msg
    | Tick Time.Posix
    | LinkClicked UrlRequest
    | UrlChanged Url
    | BuildLogWatch Int
    | BuildLogUnwatch Int
    | BuildRestart Int
    | GetLastBuilds
    | GetOrganizations
    | GetOrganizationBuilds String


type Route
    = HomeRoute
    | BuildsRoute
    | BuildRoute Int
    | OrganizationsRoute
    | OrganizationBuildsRoute String
    | NotFoundRoute


type Mutation
    = MutateLastBuilds (List Build)
    | MutateRestart
    | MutateBuild Build
    | MutateOrganizations (List Organization)
    | MutateOrganizationBuilds (List Build)
    | MutateLogLine LogLine



-- | MutateBuildLog LogLine
