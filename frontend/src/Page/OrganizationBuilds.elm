module Page.OrganizationBuilds exposing (view)

import Element
    exposing
        ( Element
        , alignRight
        , centerX
        , column
        , fill
        , maximum
        , row
        , width
        )
import Material.Icons.Av exposing (replay)
import Model exposing (Model)
import Msgs exposing (Msg(..))
import Style exposing (iconButton)
import Widget.BuildsTable as BuildsTable


view : Model -> String -> Element Msg
view model name =
    column [ width (fill |> maximum 1000), centerX ]
        [ row [ width fill ]
            [ column [ width fill ]
                [ BuildsTable.view { time = model.time, builds = model.organizationBuilds }
                ]
            ]
        , row [ width fill ]
            [ column [ alignRight ]
                [ iconButton replay "Refresh" <| Just (GetOrganizationBuilds name)
                ]
            ]
        ]
