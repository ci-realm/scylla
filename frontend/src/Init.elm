module Init exposing (init, routeFor, urlChanged, urlFor)

import Browser.Navigation as Nav
import Cmd.Extra exposing (withCmd, withCmds, withNoCmd)
import Command exposing (buildLogWatch, getBuild, getLastBuilds, getOrganizationBuilds, getOrganizations, wsConnect)
import Debounce
import Model exposing (Model)
import Msgs exposing (Msg(..), Route(..))
import Task exposing (perform)
import Time
import Url exposing (Url)
import Url.Parser exposing ((</>), Parser, custom, map, oneOf, parse, s, string, top)



-- WS.open "ws://localhost:7100/socket"


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { socket = Nothing
      , socketDebounce = Debounce.init
      , error = Nothing
      , time = Time.millisToPosix 0
      , route = routeFor url
      , url = url
      , key = key
      , lastBuilds = []
      , build = Nothing
      , organizations = []
      , organizationBuilds = []
      , buildLogPaused = False
      }
    , Cmd.batch
        [ perform Tick Time.now
        , wsConnect "ws://localhost:7100/socket"
        ]
    )


urlChanged : Model -> ( Model, Cmd msg )
urlChanged model =
    let
        route =
            routeFor model.url

        updatedModel =
            { model | route = route }
    in
    case route of
        HomeRoute ->
            updatedModel |> withNoCmd

        BuildsRoute ->
            updatedModel |> withCmd (getLastBuilds model.socket)

        BuildRoute id ->
            updatedModel |> withCmds [ getBuild model.socket id, buildLogWatch model.socket id ]

        OrganizationsRoute ->
            updatedModel |> withCmd (getOrganizations model.socket)

        OrganizationBuildsRoute name ->
            updatedModel |> withCmd (getOrganizationBuilds model.socket name)

        NotFoundRoute ->
            updatedModel |> withNoCmd


routeFor : Url -> Route
routeFor url =
    parse routes url |> Maybe.withDefault NotFoundRoute


routes : Parser (Route -> b) b
routes =
    oneOf
        [ map HomeRoute top
        , map BuildsRoute (s "builds")
        , map BuildRoute (s "build" </> buildId_)
        , map OrganizationsRoute (s "organizations")
        , map OrganizationBuildsRoute (s "organizations" </> string)
        ]


buildId_ : Parser (Int -> a) a
buildId_ =
    custom "BUILD_ID" String.toInt


urlFor : Route -> String
urlFor route =
    case route of
        HomeRoute ->
            "/"

        BuildsRoute ->
            "/builds"

        BuildRoute id ->
            "/build/" ++ String.fromInt id

        OrganizationsRoute ->
            "/organizations"

        OrganizationBuildsRoute name ->
            "/organizations/" ++ name

        NotFoundRoute ->
            "/"
