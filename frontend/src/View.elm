module View exposing (view)

import Browser
import Element
    exposing
        ( Element
        , centerX
        , column
        , el
        , fill
        , map
        , maximum
        , padding
        , px
        , row
        , spacing
        , text
        , width
        )
import Element.Background as Background
import Element.Font as Font
import Init exposing (urlFor)
import Model exposing (Model)
import Msgs exposing (Msg, Route(..))
import Page.Build
import Page.Builds
import Page.OrganizationBuilds
import Page.Organizations
import Style exposing (grey, styledLink, white)


view : Model -> Browser.Document Msg
view model =
    let
        title =
            titleFor model.route

        show =
            skeleton model title
    in
    case model.route of
        NotFoundRoute ->
            show [ text "Not Found" ]

        BuildsRoute ->
            show [ Page.Builds.view model ]

        BuildRoute _ ->
            show [ Page.Build.view model ]

        OrganizationsRoute ->
            show [ Page.Organizations.view model ]

        OrganizationBuildsRoute name ->
            show [ Page.OrganizationBuilds.view model name ]

        HomeRoute ->
            show [ text "Home" ]


skeleton : Model -> String -> List (Element Msg) -> Browser.Document Msg
skeleton model title kids =
    { title = title
    , body =
        [ Element.layout
            [ Font.size 20, Background.color grey, Font.color white, width fill ]
          <|
            column
                [ width (fill |> maximum 1000), spacing 36, padding 10, centerX ]
                [ row [ width fill ]
                    [ column [ width fill, Element.paddingXY 0 20 ] [ text "Scylla" ]
                    , column [ width fill ] [ row [ width fill ] (List.map mapNavItem navItems) ]
                    ]
                , row [ width fill ] [ el [ Font.size 30, centerX ] <| text title ]
                , row [ width fill ] kids
                , row [ width fill ]
                    [ column [ width fill ] [ text <| Maybe.withDefault "" model.error ] ]
                ]
        ]
    }


titleFor : Route -> String
titleFor route =
    case route of
        HomeRoute ->
            "Home"

        BuildRoute id ->
            "Build #" ++ String.fromInt id

        BuildsRoute ->
            "Builds"

        OrganizationsRoute ->
            "Organizations"

        OrganizationBuildsRoute name ->
            "Recent Builds for " ++ name

        NotFoundRoute ->
            "Not Found"


navItems : List ( String, String )
navItems =
    List.map (\r -> ( titleFor r, urlFor r )) [ HomeRoute, BuildsRoute, OrganizationsRoute ]


mapNavItem : ( String, String ) -> Element msg
mapNavItem ( title, link ) =
    column [ width fill ] [ styledLink { url = link, label = text title } ]
