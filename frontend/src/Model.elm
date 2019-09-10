module Model exposing (Model)

import Browser.Navigation as Nav
import Data exposing (Build, Organization, Socket)
import Debounce
import Msgs exposing (Route(..))
import Time
import Url exposing (Url)


type alias Model =
    { socket : Maybe Socket
    , socketDebounce : Debounce.Debounce String
    , error : Maybe String
    , time : Time.Posix
    , route : Route
    , url : Url
    , key : Nav.Key
    , lastBuilds : List Build
    , build : Maybe Build
    , organizations : List Organization
    , organizationBuilds : List Build
    , buildLogPaused : Bool
    }
