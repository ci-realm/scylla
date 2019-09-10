module Page.Builds exposing (view)

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
import Msgs exposing (Msg(..), Route(..))
import Style exposing (iconButton)
import Widget.BuildsTable as BuildsTable


view : Model -> Element Msg
view model =
    column [ width (fill |> maximum 1000), centerX ]
        [ row [ width fill ]
            [ column [ width fill ]
                [ BuildsTable.view
                    { time = model.time
                    , builds = model.lastBuilds
                    }
                ]
            ]
        , row [ width fill ]
            [ column [ alignRight ]
                [ iconButton replay "Refresh" <| Just GetLastBuilds
                ]
            ]
        ]
